package cockroachdb

import (
	. "github.com/journeymidnight/yig/meta/types"
	"strconv"
)

//objmap
func (t *CockroachDBClient) GetObjectMap(bucketName, objectName string) (objMap *ObjMap, err error) {
	objMap = &ObjMap{}
	sqltext := "select bucketname,objectname,nullvernum from objmap where bucketname=? and objectName=?;"
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
	sqltext := "insert into objmap(bucketname,objectname,nullvernum) values(?,?,?);"
	_, err = tx.Exec(sqltext, objMap.BucketName, objMap.Name, objMap.NullVerNum)
	return err
}

func (t *CockroachDBClient) DeleteObjectMap(objMap *ObjMap, tx DB) (err error) {
	if tx == nil {
		tx = t.Client
	}
	sqltext := "delete from objmap where bucketname=? and objectname=?;"
	_, err = tx.Exec(sqltext, objMap.BucketName, objMap.Name)
	return err
}