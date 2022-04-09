package cockroachdb

import (
	"strconv"

	. "github.com/journeymidnight/yig/meta/types"
)

//objmap
func (t *CockroachDBClient) GetObjectMap(bucketName, objectName string) (objMap *ObjMap, err error) {
	objMap = &ObjMap{}
	sqltext := "select bucketname,objectname,nullvernum from objmap where bucketname=$1 and objectName=$2;"
	err = t.Client.QueryRow(sqltext, bucketName, objectName).Scan(
		&objMap.BucketName,
		&objMap.Name,
		&objMap.NullVerNum,
	)
	if err != nil {
		return
	}
	objMap.NullVerId = strconv.FormatUint(objMap.NullVerNum, 10)
	return
}

func (t *CockroachDBClient) PutObjectMap(objMap *ObjMap, tx DB) (err error) {
	if tx == nil {
		tx = t.Client
	}
	sqltext := "insert into objmap(bucketname,objectname,nullvernum) values($1,$2,$3);"
	_, err = tx.Exec(sqltext, objMap.BucketName, objMap.Name, objMap.NullVerNum)
	return err
}

func (t *CockroachDBClient) DeleteObjectMap(objMap *ObjMap, tx DB) (err error) {
	if tx == nil {
		tx = t.Client
	}
	sqltext := "delete from objmap where bucketname=$1 and objectname=$2;"
	_, err = tx.Exec(sqltext, objMap.BucketName, objMap.Name)
	return err
}
