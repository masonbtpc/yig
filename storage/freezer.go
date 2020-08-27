package storage

import (
	. "github.com/journeymidnight/yig/error"
	meta "github.com/journeymidnight/yig/meta/types"
)

func (yig *YigStorage) GetFreezerStatus(bucketName string, objectName string, version string) (freezer *meta.Freezer, err error) {
	return yig.MetaStorage.GetFreezerStatus(bucketName, objectName, version)
}

func (yig *YigStorage) CreateFreezer(freezer *meta.Freezer, isDeceiver bool) (err error) {
	if isDeceiver {
		return yig.MetaStorage.CreateFreezerDeceiver(freezer)
	}
	return yig.MetaStorage.CreateFreezer(freezer)
}

func (yig *YigStorage) GetFreezer(bucketName string, objectName string, version string) (freezer *meta.Freezer, err error) {
	return yig.MetaStorage.GetFreezer(bucketName, objectName, version)
}

func (yig *YigStorage) UpdateFreezerDate(freezer *meta.Freezer, date int, isIncrement bool) (err error) {
	if date > 30 || date < 1 {
		return ErrInvalidRestoreDate
	}
	var lifeTime int
	if isIncrement {
		freezerInfo, err := yig.GetFreezer(freezer.BucketName, freezer.Name, freezer.VersionId)
		if err != nil {
			return err
		}
		lifeTime = freezerInfo.LifeTime + date
		if lifeTime > 30 {
			return ErrInvalidRestoreDate
		}
	} else {
		lifeTime = date
	}
	freezer.LifeTime = lifeTime
	return yig.MetaStorage.UpdateFreezerDate(freezer)
}
