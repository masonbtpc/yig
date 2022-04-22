package cockroachdb

import (
	"database/sql"
	"time"

	e "github.com/journeymidnight/yig/error"
	"github.com/journeymidnight/yig/helper"
	"github.com/journeymidnight/yig/meta/types"
)

func (t *CockroachDBClient) CreateFreezer(freezer *types.Freezer) (err error) {
	sql, args := freezer.GetCreateSql("crdb")
	_, err = t.Client.Exec(sql, args...)
	return
}

func (t *CockroachDBClient) GetFreezer(bucketName, objectName, version string) (freezer *types.Freezer, err error) {
	var lastmodifiedtime string
	sqltext := "select bucketname,objectname,IFNULL(version,''),status,lifetime,lastmodifiedtime,IFNULL(location,''),IFNULL(pool,''),IFNULL(ownerid,''),IFNULL(size,'0'),IFNULL(objectid,''),IFNULL(etag,'') from restoreobjects where bucketname=$1 and objectname=$2;"
	row := t.Client.QueryRow(sqltext, bucketName, objectName)
	freezer = &types.Freezer{}
	err = row.Scan(
		&freezer.BucketName,
		&freezer.Name,
		&freezer.VersionId,
		&freezer.Status,
		&freezer.LifeTime,
		&lastmodifiedtime,
		&freezer.Location,
		&freezer.Pool,
		&freezer.OwnerId,
		&freezer.Size,
		&freezer.ObjectId,
		&freezer.Etag,
	)
	if err == sql.ErrNoRows {
		err = e.ErrNoSuchKey
		return
	} else if err != nil {
		return
	}
	local, _ := time.LoadLocation("Local")
	freezer.LastModifiedTime, _ = time.ParseInLocation(helper.CONFIG.TimeFormat, lastmodifiedtime, local)
	freezer.Parts, err = getFreezerParts(freezer.BucketName, freezer.Name, t.Client)
	//build simple index for multipart
	if len(freezer.Parts) != 0 {
		var sortedPartNum = make([]int64, len(freezer.Parts))
		for k, v := range freezer.Parts {
			sortedPartNum[k-1] = v.Offset
		}
		freezer.PartsIndex = &types.SimpleIndex{Index: sortedPartNum}
	}
	return
}

func (t *CockroachDBClient) GetFreezerStatus(bucketName, objectName, version string) (freezer *types.Freezer, err error) {
	sqltext := "select bucketname,objectname,IFNULL(version,''),status from restoreobjects where bucketname=$1 and objectname=$2;"
	row := t.Client.QueryRow(sqltext, bucketName, objectName)
	freezer = &types.Freezer{}
	err = row.Scan(
		&freezer.BucketName,
		&freezer.Name,
		&freezer.VersionId,
		&freezer.Status,
	)
	if err == sql.ErrNoRows || freezer.Name != objectName {
		err = e.ErrNoSuchKey
		return
	}
	return
}

func (t *CockroachDBClient) UploadFreezerDate(bucketName, objectName string, lifetime int) (err error) {
	sqltext := "update restoreobjects set lifetime=$1 where bucketname=$2 and objectname=$3;"
	_, err = t.Client.Exec(sqltext, lifetime, bucketName, objectName)
	if err != nil {
		return err
	}
	return nil
}

func (t *CockroachDBClient) DeleteFreezer(bucketName, objectName string, tx types.DB) (err error) {
	if tx == nil {
		tx, err = t.Client.Begin()
		if err != nil {
			return err
		}
		defer func() {
			if err == nil {
				err = tx.(*sql.Tx).Commit()
			}
			if err != nil {
				tx.(*sql.Tx).Rollback()
			}
		}()
	}
	sqltext := "delete from restoreobjects where bucketname=$1 and objectname=$2;"
	_, err = tx.Exec(sqltext, bucketName, objectName)
	if err != nil {
		return err
	}
	sqltext = "delete from restoreobjectpart where objectname=$1 and bucketname=$2;"
	_, err = tx.Exec(sqltext, bucketName, objectName)
	if err != nil {
		return err
	}
	return nil
}

//util function
func getFreezerParts(bucketName, objectName string, cli *sql.DB) (parts map[int]*types.Part, err error) {
	parts = make(map[int]*types.Part)
	sqltext := "select partnumber,size,objectid,\"offset\",etag,lastmodified,initializationvector from restoreobjectpart where bucketname=$1 and objectname=$2;"
	rows, err := cli.Query(sqltext, bucketName, objectName)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var p *types.Part = &types.Part{}
		err = rows.Scan(
			&p.PartNumber,
			&p.Size,
			&p.ObjectId,
			&p.Offset,
			&p.Etag,
			&p.LastModified,
			&p.InitializationVector,
		)
		parts[p.PartNumber] = p
	}
	return
}
