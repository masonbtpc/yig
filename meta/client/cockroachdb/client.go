package cockroachdb

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/journeymidnight/yig/helper"
)

type CockroachDBClient struct {
	Client *sql.DB
}

func NewCockroachDBclient() *CockroachDBClient {
	cli := &CockroachDBClient{}
	conn, err := sql.Open("pgx", helper.CONFIG.DBInfo)
	if err != nil {
		os.Exit(1)
	}
	conn.SetMaxIdleConns(helper.CONFIG.DbMaxIdleConns)
	conn.SetMaxOpenConns(helper.CONFIG.DbMaxOpenConns)
	conn.SetConnMaxLifetime(time.Duration(helper.CONFIG.DbConnMaxLifeSeconds) * time.Second)
	cli.Client = conn
	return cli
}
