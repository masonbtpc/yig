package meta

import (
	"database/sql"

	e "github.com/journeymidnight/yig/error"
	"github.com/journeymidnight/yig/helper"
	"github.com/journeymidnight/yig/meta/types"
	"github.com/journeymidnight/yig/redis"
)

func (m *Meta) GetObject(bucketName string, objectName string, willNeed bool) (object *types.Object, err error) {
	getObject := func() (o interface{}, err error) {
		helper.Logger.Info("GetObject CacheMiss. bucket:", bucketName,
			"object:", objectName)
		object, err := m.Client.GetObject(bucketName, objectName, "")
		if err != nil {
			return
		}
		helper.Logger.Info("GetObject object.Name:", object.Name)
		if object.Name != objectName {
			err = e.ErrNoSuchKey
			return
		}
		return object, nil
	}
	unmarshaller := func(in []byte) (interface{}, error) {
		var object types.Object
		err := helper.MsgPackUnMarshal(in, &object)
		return &object, err
	}

	o, err := m.Cache.Get(redis.ObjectTable, bucketName+":"+objectName+":",
		getObject, unmarshaller, willNeed)
	if err != nil {
		return
	}
	object, ok := o.(*types.Object)
	if !ok {
		err = e.ErrInternalError
		return
	}
	return object, nil
}

func (m *Meta) GetAllObject(bucketName string, objectName string) (object []*types.Object, err error) {
	return m.Client.GetAllObject(bucketName, objectName, "")
}

func (m *Meta) GetObjectMap(bucketName, objectName string) (objMap *types.ObjMap, err error) {
	m.Client.GetObjectMap(bucketName, objectName)
	return
}

func (m *Meta) GetObjectVersion(bucketName, objectName, version string, willNeed bool) (object *types.Object, err error) {
	getObjectVersion := func() (o interface{}, err error) {
		object, err := m.Client.GetObject(bucketName, objectName, version)
		if err != nil {
			return
		}
		if object.Name != objectName {
			err = e.ErrNoSuchKey
			return
		}
		return object, nil
	}
	unmarshaller := func(in []byte) (interface{}, error) {
		var object types.Object
		err := helper.MsgPackUnMarshal(in, &object)
		return &object, err
	}
	o, err := m.Cache.Get(redis.ObjectTable, bucketName+":"+objectName+":"+version,
		getObjectVersion, unmarshaller, willNeed)
	if err != nil {
		return
	}
	object, ok := o.(*types.Object)
	if !ok {
		err = e.ErrInternalError
		return
	}
	return object, nil
}

func (m *Meta) PutObject(object *types.Object, multipart *types.Multipart, objMap *types.ObjMap, updateUsage bool) error {
	tx, err := m.Client.NewTrans()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			m.Client.AbortTrans(tx)
		}
	}()

	err = m.Client.PutObject(object, tx)
	if err != nil {
		return err
	}

	if objMap != nil {
		err = m.Client.PutObjectMap(objMap, tx)
		if err != nil {
			return err
		}
	}

	if multipart != nil {
		err = m.Client.DeleteMultipart(multipart, tx)
		if err != nil {
			return err
		}
	}

	if updateUsage {
		err = m.Client.UpdateUsage(object.BucketName, object.Size, tx)
		if err != nil {
			return err
		}
	}
	return m.Client.CommitTrans(tx)
}

func (m *Meta) PutObjectEntry(object *types.Object) error {
	err := m.Client.PutObject(object, nil)
	return err
}

func (m *Meta) UpdateObjectAcl(object *types.Object) error {
	err := m.Client.UpdateObjectAcl(object)
	return err
}

func (m *Meta) UpdateObjectAttrs(object *types.Object) error {
	err := m.Client.UpdateObjectAttrs(object)
	return err
}

func (m *Meta) RenameObject(object *types.Object, sourceObject string) error {
	err := m.Client.RenameObject(object, sourceObject, nil)
	return err
}

func (m *Meta) ReplaceObjectMetas(object *types.Object) error {
	err := m.Client.ReplaceObjectMetas(object, nil)
	return err
}

func (m *Meta) PutObjMapEntry(objMap *types.ObjMap) error {
	err := m.Client.PutObjectMap(objMap, nil)
	return err
}

func (m *Meta) DeleteObject(object *types.Object, DeleteMarker bool, objMap *types.ObjMap) (err error) {
	var tx *sql.Tx
	tx, err = m.Client.NewTrans()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = m.Client.CommitTrans(tx)
		}
		if err != nil {
			m.Client.AbortTrans(tx)
		}
	}()

	err = m.Client.DeleteObject(object, tx)
	if err != nil {
		return err
	}

	if objMap != nil {
		err = m.Client.DeleteObjectMap(objMap, tx)
		if err != nil {
			return err
		}
	}

	if DeleteMarker {
		return nil
	}

	err = m.Client.PutObjectToGarbageCollection(object, tx)
	if err != nil {
		return err
	}

	return m.Client.UpdateUsage(object.BucketName, -object.Size, tx)
}

func (m *Meta) UpdateGlacierObject(targetObject, sourceObject *types.Object, isFreezer bool) (err error) {
	var tx *sql.Tx
	tx, err = m.Client.NewTrans()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = m.Client.CommitTrans(tx)
		}
		if err != nil {
			m.Client.AbortTrans(tx)
		}
	}()

	if isFreezer {
		err = m.Client.UpdateObject(targetObject, tx)
		if err != nil {
			return err
		}

		err = m.Client.DeleteFreezer(sourceObject.BucketName, sourceObject.Name, tx)
		if err != nil {
			return err
		}
	} else {
		err = m.Client.PutObject(targetObject, tx)
		if err != nil {
			return err
		}
	}

	err = m.Client.PutObjectToGarbageCollection(sourceObject, tx)
	if err != nil {
		return err
	}

	return err
}

func (m *Meta) AppendObject(object *types.Object, isExist bool) error {
	tx, err := m.Client.NewTrans()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			m.Client.AbortTrans(tx)
		}
	}()
	if !isExist {
		err = m.Client.PutObject(object, tx)
	} else {
		err = m.Client.UpdateAppendObject(object, tx)
	}
	if err != nil {
		return err
	}
	err = m.Client.UpdateUsage(object.BucketName, object.Size, tx)
	if err != nil {
		return err
	}
	return m.Client.CommitTrans(tx)
}
