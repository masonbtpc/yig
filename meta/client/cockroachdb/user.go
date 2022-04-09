package cockroachdb

import (
	"database/sql"
)

func (t *CockroachDBClient) GetUserBuckets(userId string) (buckets []string, err error) {
	sqltext := "select bucketname from users where userid=$1;"
	rows, err := t.Client.Query(sqltext, userId)
	if err == sql.ErrNoRows {
		err = nil
		return
	} else if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tmp string
		err = rows.Scan(&tmp)
		if err != nil {
			return
		}
		buckets = append(buckets, tmp)
	}
	return
}

func (t *CockroachDBClient) AddBucketForUser(bucketName, userId string) (err error) {
	sql := "insert into users(userid,bucketname) values($2,$3)"
	_, err = t.Client.Exec(sql, userId, bucketName)
	return
}

func (t *CockroachDBClient) RemoveBucketForUser(bucketName string, userId string) (err error) {
	sql := "delete from users where userid=$1 and bucketname=$2;"
	_, err = t.Client.Exec(sql, userId, bucketName)
	return
}
