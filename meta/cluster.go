package meta

import (
	e "github.com/journeymidnight/yig/error"
	"github.com/journeymidnight/yig/helper"
	"github.com/journeymidnight/yig/meta/types"
	"github.com/journeymidnight/yig/redis"
)

func (m *Meta) GetClusters() (cluster []types.Cluster, err error) {
	rowKey := "cephClusters"
	getCluster := func() (c interface{}, err error) {
		helper.Logger.Info("GetClusters CacheMiss")
		return m.Client.GetClusters()
	}
	unmarshaller := func(in []byte) (interface{}, error) {
		var cluster types.Cluster
		err := helper.MsgPackUnMarshal(in, &cluster)
		return cluster, err
	}
	c, err := m.Cache.Get(redis.ClusterTable, rowKey, getCluster, unmarshaller, true)
	if err != nil {
		return
	}
	cluster, ok := c.([]types.Cluster)
	if !ok {
		err = e.ErrInternalError
		return
	}
	return cluster, nil
}
