package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bsm/redislock"
	"github.com/journeymidnight/yig/backend"
	"github.com/journeymidnight/yig/crypto"
	. "github.com/journeymidnight/yig/error"
	"github.com/journeymidnight/yig/helper"
	"github.com/journeymidnight/yig/log"
	"github.com/journeymidnight/yig/meta"
	"github.com/journeymidnight/yig/meta/client/tidbclient"
	. "github.com/journeymidnight/yig/meta/common"
	"github.com/journeymidnight/yig/meta/types"
	"github.com/journeymidnight/yig/mods"
	"github.com/journeymidnight/yig/redis"
	"github.com/journeymidnight/yig/storage"
)

const (
	WATER_LOW           = 120
	TASKQ_MAX_LENGTH    = 200
	SCAN_LIMIT          = 50
	DEFAULT_MG_LOG_PATH = "/var/log/yig/migrate.log"
	MIGRATE_JOB_MUTEX   = "MIGRATE_JOB_MUTEX"
)

var (
	yigs             []*storage.YigStorage
	signalQueue      chan os.Signal
	mgWaitgroup      sync.WaitGroup
	mgStop           bool
	mgTaskQ          chan types.Object
	mgObjectCoolDown int
	mgScanInterval   int
	mutexs           map[string]*redislock.Lock
	mux              sync.Mutex
)

func autoRefreshLock() {
	c := time.Tick(5 * time.Second)
	for {
		<-c
		if mgStop {
			helper.Logger.Info("Shutting down...")
			return
		}
		for key, lock := range mutexs {
			err := lock.Refresh(10*time.Second, nil)
			if err != nil {
				if err == redislock.ErrNotObtained {
					helper.Logger.Info("No longer hold lock ...", key)
				} else {
					helper.Logger.Info("Refresh lock failed ...", key, err.Error())
				}
				mux.Lock()
				delete(mutexs, key)
				mux.Unlock()
				continue
			}
			helper.Logger.Debug("Refresh lock success...", key)
		}
	}
}

func checkAndDoMigrate(index int) {
	for {
		if mgStop {
			helper.Logger.Info("Shutting down...")
			return
		}
		var oid, newOid string
		var bytesWritten uint64
		var err error
		var sourceCluster, destCluster backend.Cluster
		var reader io.ReadCloser
		var sourceObject, newSourceObject *types.Object
		var mutex *redislock.Lock
		object := <-mgTaskQ
		mgWaitgroup.Add(1)

		sourceObject, err = yigs[index].MetaStorage.GetObject(object.BucketName, object.Name, object.VersionId, true)
		if err != nil {
			if err == ErrNoSuchKey {
				yigs[index].MetaStorage.RemoveHotObject(&object, nil)
				goto loop
			}
			goto quit
		}

		//sometime it`s possible an object has been migrated, but still enter here, so pass it
		if sourceObject.Pool == backend.BIG_FILE_POOLNAME {
			helper.Logger.Info("object already migrated, so pass it", sourceObject.Name, sourceObject.Pool, sourceObject.ObjectId)
			goto loop
		}

		//check if object is cooldown
		if sourceObject.LastModifiedTime.Add(time.Second * time.Duration(mgObjectCoolDown)).After(time.Now()) {
			goto loop
		}

		// Try to obtain lock.
		mutex, err = redis.Locker.Obtain(redis.GenMutexKey(&object), 10*time.Second, nil)
		if err == redislock.ErrNotObtained {
			helper.Logger.Error("Lock object failed:", object.BucketName, object.ObjectId, object.VersionId)
			goto loop
		} else if err != nil {
			helper.Logger.Error("Lock seems does not work, so quit", err.Error())
			signalQueue <- syscall.SIGQUIT
			return
		}

		//add lock to mutexs map
		mux.Lock()
		mutexs[mutex.Key()] = mutex
		mux.Unlock()

		newSourceObject, err = yigs[index].MetaStorage.GetObject(object.BucketName, object.Name, object.VersionId, true)
		if err != nil {
			if err == ErrNoSuchKey {
				yigs[index].MetaStorage.RemoveHotObject(&object, nil)
				goto release
			}
			goto quit
		}

		helper.Logger.Info("start migrate for :", newSourceObject.BucketName+":"+newSourceObject.Name+":"+newSourceObject.VersionId+":"+newSourceObject.ObjectId)
		if newSourceObject.StorageClass == ObjectStorageClassGlacier {
			yigs[index].MetaStorage.RemoveHotObject(&object, nil)
			helper.Logger.Info("abort migrate because StorageClass changed for :", newSourceObject.BucketName+":"+newSourceObject.Name+":"+newSourceObject.VersionId+":"+newSourceObject.ObjectId)
			goto release
		}

		sourceCluster = yigs[index].DataStorage[newSourceObject.Location]
		reader, err = sourceCluster.GetReader(newSourceObject.Pool, newSourceObject.ObjectId, 0, uint64(newSourceObject.Size))
		if err != nil {
			helper.Logger.Error("checkIfNeedMigrate GetReader failed:", newSourceObject.Pool, newSourceObject.ObjectId, err.Error())
			goto quit
		}

		destCluster = yigs[index].PickSpecificCluster(backend.BIG_FILE_POOLNAME)
		newOid, bytesWritten, err = destCluster.Append(backend.BIG_FILE_POOLNAME, newOid, reader, 0, 0)
		if err != nil {
			helper.Logger.Error("cephCluster.Append err:", err, newOid)
			goto quit
		}
		if bytesWritten != uint64(sourceObject.Size) {
			destCluster.Remove(backend.BIG_FILE_POOLNAME, newOid)
			helper.Logger.Error("cephCluster.Append write length to hdd not equel the object size:", newOid, bytesWritten, newSourceObject.Size)
			goto release
		}

		//update object fileds
		newSourceObject.Location = destCluster.ID()
		newSourceObject.Pool = backend.BIG_FILE_POOLNAME
		oid = newSourceObject.ObjectId
		newSourceObject.ObjectId = newOid
		//update objects table and remove entry from hotobjects
		err = yigs[index].MetaStorage.MigrateObject(newSourceObject)
		if err != nil {
			destCluster.Remove(backend.BIG_FILE_POOLNAME, newOid)
			helper.Logger.Error("cephCluster.Append MigrateObject failed:", err.Error(), newSourceObject.Pool, newSourceObject.ObjectId)
			goto quit
		}
		//remove data from ssd cluster
		err = sourceCluster.Remove(backend.SMALL_FILE_POOLNAME, oid)
		if err != nil {
			helper.Logger.Error("cephCluster.Append Remove data from rabbit failed:", err.Error(), newSourceObject.Pool, newSourceObject.ObjectId)
			if !strings.Contains(err.Error(), "ret=-2") {
				goto quit
			}
		}
		//invalid redis cache
		yigs[index].MetaStorage.Cache.Remove(redis.ObjectTable, sourceObject.BucketName+":"+sourceObject.Name+":"+sourceObject.VersionId)
		yigs[index].DataCache.Remove(sourceObject.BucketName + ":" + sourceObject.Name + ":" + sourceObject.VersionId)
		helper.Logger.Info("migrate success for bucket: ", sourceObject.BucketName+" name: "+sourceObject.Name+" version: "+sourceObject.VersionId+" oldoid: "+sourceObject.ObjectId+" newoid: "+newSourceObject.ObjectId)
		goto release
	quit:
		signalQueue <- syscall.SIGQUIT
	release:
		mutex.Release()
		mux.Lock()
		delete(mutexs, mutex.Key())
		mux.Unlock()
	loop:
		mgWaitgroup.Done()
	}
}

func getHotObjects() {

	helper.Logger.Info("getHotObjects thread start")
	var customattributes, acl, lastModifiedTime string
	var sqltext string
	var rows *sql.Rows
	var err error
	var mutex *redislock.Lock

	for {
		// Try to obtain lock.
		mutex, err = redis.Locker.Obtain(MIGRATE_JOB_MUTEX, 10*time.Second, nil)
		if err == redislock.ErrNotObtained {
			helper.Logger.Info("Lock object failed, sleep 30s:", MIGRATE_JOB_MUTEX)
			time.Sleep(30 * time.Second)
			continue
		} else if err != nil {
			helper.Logger.Error("Lock seems does not work, so quit", err.Error())
			signalQueue <- syscall.SIGQUIT
			return
		}
		break
	}
	defer func() {
		mutex.Release()
		mux.Lock()
		delete(mutexs, mutex.Key())
		mux.Unlock()
	}()

	mux.Lock()
	mutexs[mutex.Key()] = mutex
	mux.Unlock()
	client := tidbclient.NewTidbClient()
	for {
		if mgStop {
			helper.Logger.Info("shutting down...")
			return
		}

		for len(mgTaskQ) > 0 {
			time.Sleep(time.Duration(1) * time.Second)
			helper.Logger.Info("wait for last round migrate jobs finished...")
			if mgStop {
				helper.Logger.Info("shutting down...")
				return
			}
		}

		sqltext = "select bucketname,name,version,location,pool,ownerid,size,objectid,lastmodifiedtime,etag,contenttype," +
			"customattributes,acl,nullversion,deletemarker,ssetype,encryptionkey,initializationvector,type,storageclass,createtime" +
			" from hotobjects order by bucketname,name,version;"
		rows, err = client.Client.Query(sqltext)
		if err != nil {
			helper.Logger.Error("getHotObjects err:", err)
			goto quit
		}
		helper.Logger.Info("query tidb success")
		for rows.Next() {
			//fetch related date
			object := &types.Object{}
			err = rows.Scan(
				&object.BucketName,
				&object.Name,
				&object.VersionId,
				&object.Location,
				&object.Pool,
				&object.OwnerId,
				&object.Size,
				&object.ObjectId,
				&lastModifiedTime,
				&object.Etag,
				&object.ContentType,
				&customattributes,
				&acl,
				&object.NullVersion,
				&object.DeleteMarker,
				&object.SseType,
				&object.EncryptionKey,
				&object.InitializationVector,
				&object.Type,
				&object.StorageClass,
				&object.CreateTime,
			)
			if err != nil {
				goto quit
			}
			object.LastModifiedTime, err = time.Parse("2006-01-02 15:04:05", lastModifiedTime)
			if err != nil {
				goto quit
			}
			err = json.Unmarshal([]byte(acl), &object.ACL)
			if err != nil {
				goto quit
			}
			err = json.Unmarshal([]byte(customattributes), &object.CustomAttributes)
			if err != nil {
				goto quit
			}
			mgTaskQ <- *object

			for len(mgTaskQ) >= WATER_LOW {
				time.Sleep(time.Duration(10) * time.Millisecond)
			}
			if mgStop {
				helper.Logger.Info("shutting down...")
				return
			}
		}
		for i := 0; i < mgScanInterval; i++ {
			time.Sleep(time.Duration(1) * time.Second)
			if mgStop {
				helper.Logger.Info("shutting down...")
				return
			}
		}
	}
quit:
	if rows != nil {
		rows.Close()
	}
	signalQueue <- syscall.SIGQUIT
	return
}

func main() {
	mgStop = false

	helper.SetupConfig()
	logLevel := log.ParseLevel(helper.CONFIG.LogLevel)

	helper.Logger = log.NewFileLogger(DEFAULT_MG_LOG_PATH, logLevel)
	defer helper.Logger.Close()
	mgTaskQ = make(chan types.Object, TASKQ_MAX_LENGTH)
	signal.Ignore()
	signalQueue = make(chan os.Signal)
	if helper.CONFIG.MetaCacheType > 0 || helper.CONFIG.EnableDataCache {
		redis.Initialize()
		defer redis.RedisConn.Close()
	}
	// Read all *.so from plugins directory, and fill the variable allPlugins
	allPluginMap := mods.InitialPlugins()
	kms := crypto.NewKMS(allPluginMap)

	numOfWorkers := helper.CONFIG.MgThread
	yigs = make([]*storage.YigStorage, helper.CONFIG.MgThread+1)
	yigs[0] = storage.New(int(meta.NoCache), false, kms)
	helper.Logger.Info("start migrate thread:", numOfWorkers)
	if helper.CONFIG.DebugMode == true {
		mgObjectCoolDown = 1
		mgScanInterval = 5
	} else {
		mgObjectCoolDown = helper.CONFIG.MgObjectCooldown
		mgScanInterval = helper.CONFIG.MgScanInterval
	}
	helper.Logger.Info("migrate service parameters:", mgObjectCoolDown, mgScanInterval)
	mutexs = make(map[string]*redislock.Lock)
	for i := 0; i < numOfWorkers; i++ {
		yigs[i+1] = storage.New(helper.CONFIG.MetaCacheType, helper.CONFIG.EnableDataCache, kms)
		if helper.CONFIG.CacheCircuitCheckInterval != 0 && helper.CONFIG.MetaCacheType != 0 {
			go func(i int) {
				yigs[i+1].PingCache(time.Duration(helper.CONFIG.CacheCircuitCheckInterval) * time.Second)
			}(i)
		}
		go checkAndDoMigrate(i + 1)
	}
	go getHotObjects()
	go autoRefreshLock()
	signal.Notify(signalQueue, syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGQUIT, syscall.SIGHUP)
	for {
		s := <-signalQueue
		switch s {
		case syscall.SIGHUP:
			// reload config file
			helper.SetupConfig()
		default:
			// coolStop YIG server, order matters
			mgStop = true
			mgWaitgroup.Wait()
			return
		}
	}

}
