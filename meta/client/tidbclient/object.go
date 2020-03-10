package tidbclient

import (
	"database/sql"
	. "database/sql/driver"
	"encoding/json"
	"math"
	"strconv"
	"time"

	. "github.com/journeymidnight/yig/error"
	. "github.com/journeymidnight/yig/meta/types"
)

func (t *TidbClient) GetObject(bucketName, objectName, version string) (object *Object, err error) {
	var ibucketname, iname, customattributes, acl, lastModifiedTime string
	var iversion uint64

	var row *sql.Row
	sqltext := "select bucketname,name,version,location,pool,ownerid,size,objectid,lastmodifiedtime,etag,contenttype," +
		"customattributes,acl,nullversion,deletemarker,ssetype,encryptionkey,initializationvector,type,storageclass,createtime from objects where bucketname=? and name=? "
	if version == "" || version == NullVersion {
		sqltext += "and version='0'"
		row = t.Client.QueryRow(sqltext, bucketName, objectName)
	} else {
		sqltext += "and version=?;"
		row = t.Client.QueryRow(sqltext, bucketName, objectName, version)
	}
	object = &Object{}
	err = row.Scan(
		&ibucketname,
		&iname,
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
	if err == sql.ErrNoRows {
		err = ErrNoSuchKey
		return
	} else if err != nil {
		return
	}
	object.LastModifiedTime, err = time.Parse("2006-01-02 15:04:05", lastModifiedTime)
	if err != nil {
		return
	}
	object.Name = objectName
	object.BucketName = bucketName
	err = json.Unmarshal([]byte(acl), &object.ACL)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(customattributes), &object.CustomAttributes)
	if err != nil {
		return
	}
	if object.Type == ObjectTypeMultipart {
		iversion = math.MaxUint64 - object.CreateTime
		object.Parts, err = getParts(object.BucketName, object.Name, iversion, t.Client)
		//build simple index for multipart
		if len(object.Parts) != 0 {
			var sortedPartNum = make([]int64, len(object.Parts))
			for k, v := range object.Parts {
				sortedPartNum[k-1] = v.Offset
			}
			object.PartsIndex = &SimpleIndex{Index: sortedPartNum}
		}
	}
	return
}

func (t *TidbClient) UpdateObjectAttrs(object *Object) error {
	sql, args := object.GetUpdateAttrsSql()
	_, err := t.Client.Exec(sql, args...)
	return err
}

func (t *TidbClient) UpdateObjectAcl(object *Object) error {
	sql, args := object.GetUpdateAclSql()
	_, err := t.Client.Exec(sql, args...)
	return err
}

func (t *TidbClient) RenameObject(object *Object, sourceObject string) (err error) {
	sql, args := object.GetUpdateNameSql(sourceObject)
	if len(object.Parts) != 0 {
		tx, err := t.Client.Begin()
		if err != nil {
			return err
		}
		defer func() {
			if err == nil {
				err = tx.Commit()
			}
			if err != nil {
				tx.Rollback()
			}
		}()
		_, err = tx.Exec(sql, args...)
		if err != nil {
			return err
		}

		// rename parts
		sql, args = object.GetUpdateObjectPartNameSql(sourceObject)
		_, err = tx.Exec(sql, args...)
		return err
	}
	_, err = t.Client.Exec(sql, args...)
	return
}

func (t *TidbClient) ReplaceObjectMetas(object *Object, tx Tx) (err error) {
	sql, args := object.GetReplaceObjectMetasSql()
	_, err = t.Client.Exec(sql, args...)
	return
}

func (t *TidbClient) UpdateAppendObject(object *Object) (err error) {
	tx, err := t.Client.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	lastModifiedTime := object.LastModifiedTime.Format(TIME_LAYOUT_TIDB)
	sql := "update objects set lastmodifiedtime=?, size=? where bucketname=? and name=? and version=?"
	args := []interface{}{lastModifiedTime, object.Size, object.BucketName, object.Name, object.VersionId}
	_, err = tx.Exec(sql, args...)

	return t.UpdateUsage(object.BucketName, object.Size, tx)
}

func (t *TidbClient) PutObjectWithoutMultiPart(object *Object) error {
	sql, args := object.GetCreateSql()
	_, err := t.Client.Exec(sql, args...)
	return err
}

func (t *TidbClient) UpdateObjectWithoutMultiPart(object *Object) error {
	sql, args := object.GetUpdateSql()
	_, err := t.Client.Exec(sql, args...)
	return err
}

func (t *TidbClient) PutObject(object *Object, multipart *Multipart, updateUsage bool) (err error) {
	tx, err := t.Client.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	sql, args := object.GetCreateSql()
	_, err = tx.Exec(sql, args...)
	if object.Parts != nil {
		v := math.MaxUint64 - uint64(object.LastModifiedTime.UnixNano())
		version := strconv.FormatUint(v, 10)
		for _, p := range object.Parts {
			psql, args := p.GetCreateSql(object.BucketName, object.Name, version)
			_, err = tx.Exec(psql, args...)
			if err != nil {
				return err
			}
		}
	}

	if multipart != nil {
		return t.DeleteMultipart(multipart, tx)
	}

	if updateUsage {
		return t.UpdateUsage(object.BucketName, object.Size, tx)
	}

	return nil
}

func (t *TidbClient) UpdateObject(object *Object, multipart *Multipart, updateUsage bool) (err error) {

	tx, err := t.Client.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	sql, args := object.GetUpdateSql()
	_, err = tx.Exec(sql, args...)
	if object.Parts != nil {
		v := math.MaxUint64 - uint64(object.LastModifiedTime.UnixNano())
		version := strconv.FormatUint(v, 10)
		for _, p := range object.Parts {
			psql, args := p.GetCreateSql(object.BucketName, object.Name, version)
			_, err = tx.Exec(psql, args...)
			if err != nil {
				return err
			}
		}
	}

	if multipart != nil {
		err = t.DeleteMultipart(multipart, tx)
		if err != nil {
			return err
		}
	}

	if updateUsage {
		err = t.UpdateUsage(object.BucketName, object.Size, tx)
		if err != nil {
			return err
		}
	}

	return err
}

func (t *TidbClient) UpdateFreezerObject(object *Object, tx Tx) (err error) {
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
	txn := tx.(*sql.Tx)
	v := math.MaxUint64 - uint64(object.LastModifiedTime.UnixNano())
	version := strconv.FormatUint(v, 10)
	sqltext := "delete from objectpart where objectname=? and bucketname=? and version=?;"
	_, err = txn.Exec(sqltext, object.Name, object.BucketName, version)
	if err != nil {
		return err
	}

	sql, args := object.GetGlacierUpdateSql()
	_, err = txn.Exec(sql, args...)
	if object.Parts != nil {
		for _, p := range object.Parts {
			psql, args := p.GetCreateSql(object.BucketName, object.Name, version)
			_, err = txn.Exec(psql, args...)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *TidbClient) UpdateFreezerObject(object *Object, tx DB) (err error) {
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

	v := math.MaxUint64 - uint64(object.LastModifiedTime.UnixNano())
	version := strconv.FormatUint(v, 10)
	sqltext := "delete from objectpart where objectname=? and bucketname=? and version=?;"
	_, err = tx.Exec(sqltext, object.Name, object.BucketName, version)
	if err != nil {
		return err
	}

	sql, args := object.GetGlacierUpdateSql()
	_, err = tx.Exec(sql, args...)
	if object.Parts != nil {
		for _, p := range object.Parts {
			psql, args := p.GetCreateSql(object.BucketName, object.Name, version)
			_, err = tx.Exec(psql, args...)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *TidbClient) DeleteObject(object *Object, tx Tx) (err error) {
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

	sqltext := "delete from objects where name=? and bucketname=? and version=?;"
	_, err = tx.(*sql.Tx).Exec(sqltext, object.Name, object.BucketName, object.VersionId)
	if err != nil {
		return err
	}

	v := math.MaxUint64 - object.CreateTime
	version := strconv.FormatUint(v, 10)
	sqltext = "delete from objectpart where objectname=? and bucketname=? and version=?;"
	_, err = tx.(*sql.Tx).Exec(sqltext, object.Name, object.BucketName, version)
	if err != nil {
		return err
	}
	return nil
}

func (t *TidbClient) DeleteOldObjects(latestObject *Object, tx Tx) (err error) {
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

	sqltext := "delete from objects where bucketname=? and name=? and version>?;"
	_, err = tx.(*sql.Tx).Exec(sqltext, latestObject.Name, latestObject.BucketName, latestObject.VersionId)
	if err != nil {
		return err
	}
	sqltext = "delete from objectpart where bucketname=? and objectname=? and version>?;"
	_, err = tx.(*sql.Tx).Exec(sqltext, latestObject.Name, latestObject.BucketName, latestObject.VersionId)
	if err != nil {
		return err
	}
	return nil
}

//util function
func getParts(bucketName, objectName string, version uint64, cli *sql.DB) (parts map[int]*Part, err error) {
	parts = make(map[int]*Part)
	sqltext := "select partnumber,size,objectid,offset,etag,lastmodified,initializationvector from objectpart where bucketname=? and objectname=? and version=?;"
	rows, err := cli.Query(sqltext, bucketName, objectName, version)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var p *Part = &Part{}
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
