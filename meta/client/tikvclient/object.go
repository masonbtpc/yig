package tikvclient

import (
	"context"
	. "database/sql/driver"

	. "github.com/journeymidnight/yig/error"
	"github.com/journeymidnight/yig/helper"
	. "github.com/journeymidnight/yig/meta/types"
	"github.com/tikv/client-go/key"
)

// **Key**: {BucketName}\\{ObjectName}
// **Versioned Key**: {BucketName}\\{ObjectName}\\{Version}
// Version = hex.EncodeToString(BigEndian(MaxUint64 - multipart.InitialTime))
func genObjectKey(bucketName, objectName, version string) []byte {
	// TODO: GetLatestObject
	if version == NullVersion || version == "" {
		return GenKey(bucketName, objectName)
	} else {
		return GenKey(bucketName, objectName, version)
	}
}

//object
func (c *TiKVClient) GetObject(bucketName, objectName, version string) (*Object, error) {
	key := genObjectKey(bucketName, objectName, version)
	var o Object
	ok, err := c.TxGet(key, &o)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNoSuchKey
	}

	if o.Parts != nil && len(o.Parts) != 0 {
		var sortedPartNum = make([]int64, len(o.Parts))
		for k, v := range o.Parts {
			sortedPartNum[k-1] = v.Offset
		}
		o.PartsIndex = &SimpleIndex{Index: sortedPartNum}
	}
	return &o, nil
}

func (c *TiKVClient) GetLatestVersionedObject(bucketName, objectName string) (object *Object, err error) {
	objKey := genObjectKey(bucketName, objectName, NullVersion)
	var o, vo Object
	nullObjExist, err := c.TxGet(objKey, &o)
	if err != nil {
		return nil, err
	}

	versionStartKey := genObjectKey(bucketName, objectName, TableMinKeySuffix)
	versionEndKey := genObjectKey(bucketName, objectName, TableMaxKeySuffix)
	kvs, err := c.TxScan(key.Key(versionStartKey), key.Key(versionEndKey), 1)
	if err != nil {
		return nil, err
	}
	if !nullObjExist && len(kvs) == 0 {
		return nil, ErrNoSuchKey
	} else if !nullObjExist {
		err = helper.MsgPackUnMarshal(kvs[0].V, &vo)
		if err != nil {
			return nil, err
		}
		return &vo, nil
	} else if len(kvs) == 0 {
		return &o, nil
	} else {
		reto := helper.Ternary(o.LastModifiedTime.After(vo.LastModifiedTime), &o, &vo)
		return reto.(*Object), nil
	}
}

func (c *TiKVClient) PutObject(object *Object, multipart *Multipart, updateUsage bool) error {
	objectKey := genObjectKey(object.BucketName, object.Name, object.VersionId)
	tx, err := c.NewTrans()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = c.CommitTrans(tx)
		}
		if err != nil {
			c.AbortTrans(tx)
		}
	}()

	txn := tx.(*TikvTx).tx
	if multipart != nil {
		object.Parts = multipart.Parts
		err = c.DeleteMultipart(multipart, tx)
		if err != nil {
			return err
		}
	}

	objectVal, err := helper.MsgPackMarshal(object)
	if err != nil {
		return err
	}

	err = txn.Set(objectKey, objectVal)
	if err != nil {
		return err
	}

	if updateUsage {
		return c.UpdateUsage(object.BucketName, object.Size, tx)
	}
	return nil
}

func (c *TiKVClient) PutVersionedObject(object *Object, multipart *Multipart, updateUsage bool) error {
	return c.PutObject(object, multipart, updateUsage)
}

func (c *TiKVClient) PutObjectWithoutMultiPart(object *Object) error {
	objectKey := genObjectKey(object.BucketName, object.Name, object.VersionId)
	tx, err := c.NewTrans()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = c.CommitTrans(tx)
		}
		if err != nil {
			c.AbortTrans(tx)
		}
	}()

	txn := tx.(*TikvTx).tx
	objectVal, err := helper.MsgPackMarshal(object)
	if err != nil {
		return err
	}

	err = txn.Set(objectKey, objectVal)
	if err != nil {
		return err
	}
	return c.UpdateUsage(object.BucketName, object.Size, tx)
}

func (c *TiKVClient) UpdateObject(object *Object, multipart *Multipart, updateUsage bool) (err error) {
	return c.PutObject(object, multipart, updateUsage)
}

func (c *TiKVClient) UpdateObjectWithoutMultiPart(object *Object) error {
	return c.PutObjectWithoutMultiPart(object)
}

func (c *TiKVClient) UpdateAppendObject(object *Object) error {
	return c.PutObjectWithoutMultiPart(object)
}

func (c *TiKVClient) RenameObject(object *Object, sourceObject string) (err error) {
	oldKey := genObjectKey(object.BucketName, sourceObject, NullVersion)
	newKey := genObjectKey(object.BucketName, object.Name, NullVersion)

	tx, err := c.TxnCli.Begin(context.TODO())
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit(context.Background())
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	v, err := helper.MsgPackMarshal(*object)
	if err != nil {
		return err
	}
	err = tx.Set(newKey, v)
	if err != nil {
		return err
	}
	err = tx.Delete(oldKey)
	if err != nil {
		return err
	}
	return nil
}

func (c *TiKVClient) DeleteObject(object *Object, tx Tx) error {
	key := genObjectKey(object.BucketName, object.Name, object.VersionId)
	if tx == nil {
		tx, err := c.NewTrans()
		if err != nil {
			return err
		}
		defer func() {
			if err == nil {
				err = c.CommitTrans(tx)
			}
			if err != nil {
				c.AbortTrans(tx)
			}
		}()
	}

	txn := tx.(*TikvTx).tx
	return txn.Delete(key)
}

func (c *TiKVClient) DeleteVersionedObject(object *Object, tx Tx) error {
	return nil
}

func (c *TiKVClient) UpdateObjectAcl(object *Object) error {
	return c.PutObjectWithoutMultiPart(object)
}

func (c *TiKVClient) UpdateObjectAttrs(object *Object) error {
	return c.PutObjectWithoutMultiPart(object)
}
