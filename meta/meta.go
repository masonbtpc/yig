package meta

import (
	"github.com/journeymidnight/yig/helper"
	"github.com/journeymidnight/yig/meta/client"
	"github.com/journeymidnight/yig/meta/client/cockroachdb"
	"github.com/journeymidnight/yig/meta/client/tidbclient"
)

const (
	ENCRYPTION_KEY_LENGTH = 32 // 32 bytes for AES-"256"
)

type Meta struct {
	Client client.Client
	Cache  MetaCache
	Type   string
}

func New(myCacheType CacheType) *Meta {
	meta := Meta{
		Cache: newMetaCache(myCacheType),
	}
	if helper.CONFIG.MetaStore == "tidb" {
		meta.Client = tidbclient.NewTidbClient()
		meta.Type = "tidb"
	} else if helper.CONFIG.MetaStore == "cockroachdb" {
		meta.Client = cockroachdb.NewCockroachDBclient()
		meta.Type = "cockroachdb"
	} else {
		panic("unsupported metastore")
	}
	return &meta
}
