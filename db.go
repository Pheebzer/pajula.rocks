package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var insertTrackQuery string = `
INSERT INTO tracks (id,name,album,artist,user,added_at,duration_ms,snapshot_id)
  VALUES (?,?,?,?,?,?,?,?);
`
var deleteTracksQuery = `
DELETE FROM tracks;
`
var getSnapshotQuery = `
SELECT snapshot_id FROM metadata;
`
var updateSnapshotQuery = `
UPDATE metadata SET snapshot_id=?;
`

func InitDB(cfg *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@/%s", cfg.Db.User, cfg.Db.Password, cfg.Db.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}
