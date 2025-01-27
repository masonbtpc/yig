package cockroachdb

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v4"
	e "github.com/journeymidnight/yig/error"
	"github.com/journeymidnight/yig/helper"
	"github.com/journeymidnight/yig/meta/types"
)

func (t *CockroachDBClient) GetBucket(bucketName string) (bucket *types.Bucket, err error) {
	var acl, cors, logging, lc, policy, website, encryption, createTime string
	sqltext := "select bucketname,acl,cors,COALESCE(logging,null),lc,uid,policy,website,COALESCE(encryption,null),createtime,usages,versioning from buckets where bucketname=$1;"
	bucket = new(types.Bucket)
	err = t.Client.QueryRow(sqltext, bucketName).Scan(
		&bucket.Name,
		&acl,
		&cors,
		&logging,
		&lc,
		&bucket.OwnerId,
		&policy,
		&website,
		&encryption,
		&createTime,
		&bucket.Usage,
		&bucket.Versioning,
	)
	if err != nil && err == sql.ErrNoRows {
		err = e.ErrNoSuchBucket
		return
	} else if err != nil {
		return
	}
	bucket.CreateTime, err = time.Parse(helper.CONFIG.TimeFormat, createTime)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(acl), &bucket.ACL)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(cors), &bucket.CORS)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(logging), &bucket.BucketLogging)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(lc), &bucket.Lifecycle)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(policy), &bucket.Policy)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(website), &bucket.Website)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(encryption), &bucket.Encryption)
	if err != nil {
		return
	}
	return
}

func (t *CockroachDBClient) GetBuckets() (buckets []types.Bucket, err error) {
	sqltext := "select bucketname,acl,cors,COALESCE(logging,null),lc,uid,policy,website,COALESCE(encryption,null),createtime,usages,versioning from buckets;"
	rows, err := t.Client.Query(sqltext)
	if err == sql.ErrNoRows {
		err = nil
		return
	} else if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tmp types.Bucket
		var acl, cors, logging, lc, policy, website, encryption, createTime string
		err = rows.Scan(
			&tmp.Name,
			&acl,
			&cors,
			&logging,
			&lc,
			&tmp.OwnerId,
			&policy,
			&website,
			&encryption,
			&createTime,
			&tmp.Usage,
			&tmp.Versioning)
		if err != nil {
			return
		}
		tmp.CreateTime, err = time.Parse(helper.CONFIG.TimeFormat, createTime)
		if err != nil {
			return
		}
		err = json.Unmarshal([]byte(acl), &tmp.ACL)
		if err != nil {
			return
		}
		err = json.Unmarshal([]byte(cors), &tmp.CORS)
		if err != nil {
			return
		}
		err = json.Unmarshal([]byte(logging), &tmp.BucketLogging)
		if err != nil {
			return
		}
		err = json.Unmarshal([]byte(lc), &tmp.Lifecycle)
		if err != nil {
			return
		}
		err = json.Unmarshal([]byte(policy), &tmp.Policy)
		if err != nil {
			return
		}
		err = json.Unmarshal([]byte(website), &tmp.Website)
		if err != nil {
			return
		}
		err = json.Unmarshal([]byte(encryption), &tmp.Encryption)
		if err != nil {
			return
		}
		buckets = append(buckets, tmp)
	}
	return
}

//Actually this method is used to update bucket
func (t *CockroachDBClient) PutBucket(bucket types.Bucket) error {
	sql, args := bucket.GetUpdateSql("crdb")
	_, err := t.Client.Exec(sql, args...)
	if err != nil {
		return err
	}
	return nil
}

func (t *CockroachDBClient) CheckAndPutBucket(bucket types.Bucket) (bool, error) {
	var processed bool
	_, err := t.GetBucket(bucket.Name)
	if err == nil {
		processed = false
		return processed, err
	} else if err != nil && err != e.ErrNoSuchBucket {
		processed = false
		return processed, err
	} else {
		processed = true
	}
	sql, args := bucket.GetCreateSql("crdb")
	_, err = t.Client.Exec(sql, args...)
	return processed, err
}

func (t *CockroachDBClient) ListObjects(bucketName, marker, verIdMarker, prefix, delimiter string, versioned bool, maxKeys int) (retObjects []*types.Object, prefixes []string, truncated bool, nextMarker, nextVerIdMarker string, err error) {
	if versioned {
		return
	}
	var count int
	var exit bool
	objectNum := make(map[string]int)
	commonPrefixes := make(map[string]struct{})
	omarker := marker
	for {
		var loopcount int
		var sqltext string
		var rows *sql.Rows
		if prefix == "" {
			if marker == "" {
				sqltext = `select bucketname,name,version,nullversion,deletemarker 
					from objects 
					where bucketName=$1 
					order by bucketname,name,version 
					limit $2`
				rows, err = t.Client.Query(sqltext, bucketName, maxKeys)
				if err != nil {
					continue
				}
			} else {
				sqltext = `select bucketname,name,version,nullversion,deletemarker 
					from objects 
					where bucketName=$1 
					and name >=$2 
					order by bucketname,name,version 
					offset $3 limit $4`
				rows, err = t.Client.Query(sqltext, bucketName, marker, objectNum[marker], objectNum[marker]+maxKeys)
				if err != nil {
					continue
				}
			}
		} else { // prefix not empty
			prefixPattern := prefix + "%"
			if marker == "" {
				sqltext = `select bucketname,name,version,nullversion,deletemarker 
					from objects 
					where bucketName=$1 
					and name like $2
					order by bucketname,name,version 
					limit $3`
				rows, err = t.Client.Query(sqltext, bucketName, prefixPattern, maxKeys)
				if err != nil {
					continue
				}
			} else {
				sqltext = `select bucketname,name,version,nullversion,deletemarker 
					from objects 
					where bucketName=$1 
					and name >=$2 
					and name like $3
					order by bucketname,name,version 
					offset $4 limit $5`
				rows, err = t.Client.Query(sqltext, bucketName, marker, prefixPattern,
					objectNum[marker], objectNum[marker]+maxKeys)
				if err != nil {
					continue
				}
			}
		}
		if err != nil {
			return
		}
		for rows.Next() {
			loopcount += 1
			//fetch related date
			var bucketname, name string
			var version uint64
			var nullversion, deletemarker bool
			err = rows.Scan(
				&bucketname,
				&name,
				&version,
				&nullversion,
				&deletemarker,
			)
			if err != nil {
				_ = rows.Close()
				return
			}
			//prepare next marker
			//TODU: be sure how tidb/mysql compare strings
			if _, ok := objectNum[name]; !ok {
				objectNum[name] = 0
			}
			objectNum[name] += 1
			marker = name
			//filte row
			//filte by prefix
			hasPrefix := strings.HasPrefix(name, prefix)
			if !hasPrefix {
				continue
			}
			//filte by objectname
			if objectNum[name] > 1 {
				continue
			}
			//filte by deletemarker
			if deletemarker {
				continue
			}
			if name == omarker {
				continue
			}
			//filte by delemiter
			if len(delimiter) != 0 {
				subStr := strings.TrimPrefix(name, prefix)
				n := strings.Index(subStr, delimiter)
				if n != -1 {
					prefixKey := prefix + subStr[0:(n+1)]
					if prefixKey == omarker {
						continue
					}
					if _, ok := commonPrefixes[prefixKey]; !ok {
						if count == maxKeys {
							truncated = true
							exit = true
							break
						}
						commonPrefixes[prefixKey] = struct{}{}
						nextMarker = prefixKey
						count += 1
					}
					continue
				}
			}
			var o *types.Object
			Strver := strconv.FormatUint(version, 10)
			o, err = t.GetObject(bucketname, name, Strver)
			if err == e.ErrNoSuchKey {
				// it's possible the object is already deleted
				continue
			}
			if err != nil {
				_ = rows.Close()
				return
			}
			count += 1
			if count == maxKeys {
				nextMarker = name
			}
			if count == 0 {
				continue
			}
			if count > maxKeys {
				truncated = true
				exit = true
				break
			}
			retObjects = append(retObjects, o)
		}
		_ = rows.Close()
		if loopcount == 0 {
			exit = true
		}
		if exit {
			break
		}
	}
	prefixes = helper.Keys(commonPrefixes)
	return
}

func (t *CockroachDBClient) DeleteBucket(bucket types.Bucket) error {
	sqltext := "delete from buckets where bucketname=$1;"
	_, err := t.Client.Exec(sqltext, bucket.Name)
	if err != nil {
		return err
	}
	return nil
}

func (t *CockroachDBClient) UpdateUsage(bucketName string, size int64, tx types.DB) (err error) {
	if !helper.CONFIG.PiggybackUpdateUsage {
		return nil
	}

	if tx == nil {
		tx = t.Client
	}
	sql := "update buckets set usages= usages + $1 where bucketname=$2;"
	_, err = tx.Exec(sql, size, bucketName)
	return
}
