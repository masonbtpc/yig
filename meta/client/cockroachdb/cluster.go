package cockroachdb

import (
	"github.com/journeymidnight/yig/meta/types"
)

func (t *CockroachDBClient) GetClusters() (cluster []types.Cluster, err error) {
	sqltext := "select fsid,pool,weight from cluster"
	rows, err := t.Client.Query(sqltext)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		c := types.Cluster{}
		err = rows.Scan(&c.Fsid, &c.Pool, &c.Weight)
		cluster = append(cluster, c)
		if err != nil {
			return nil, err
		}
	}
	return cluster, nil
}
