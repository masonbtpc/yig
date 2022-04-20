package meta

import "github.com/journeymidnight/yig/meta/types"

// Insert object to `garbageCollection` table
func (m *Meta) PutObjectToGarbageCollection(object *types.Object) error {
	return m.Client.PutObjectToGarbageCollection(object, nil)
}

func (m *Meta) ScanGarbageCollection(limit int, startRowKey string) ([]types.GarbageCollection, error) {
	return m.Client.ScanGarbageCollection(limit, startRowKey)
}

func (m *Meta) RemoveGarbageCollection(garbage types.GarbageCollection) error {
	return m.Client.RemoveGarbageCollection(garbage)
}
