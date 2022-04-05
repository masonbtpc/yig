package cockroachdb

import "database/sql"

func (t *CockroachDBClient) NewTrans() (tx *sql.Tx, err error) {
	tx, err = t.Client.Begin()
	return
}

func (t *CockroachDBClient) AbortTrans(tx *sql.Tx) (err error) {
	err = tx.Rollback()
	return
}

func (t *CockroachDBClient) CommitTrans(tx *sql.Tx) (err error) {
	err = tx.Commit()
	return
}
