package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type ImporterTx struct {
	tx         *sql.Tx
	Statements Statements
}

type Statements struct {
	InsertUpdateTrack *sql.Stmt
	DeleteTracks      *sql.Stmt
	GetSnapshotId     *sql.Stmt
	UpdateSnapshotId  *sql.Stmt
}

func InitDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}

func StartTransaction(db *sql.DB) (*ImporterTx, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	statements := getStatements(tx)
	return &ImporterTx{
		tx:         tx,
		Statements: statements,
	}, nil
}

func getStatements(tx *sql.Tx) Statements {
	return Statements{
		InsertUpdateTrack: prepareStmt(tx, `
			INSERT INTO tracks (id,name,album,artist,user,added_at,duration_ms,snapshot_id)
			VALUES (?,?,?,?,?,?,?,?)
			ON DUPLICATE KEY UPDATE snapshot_id = VALUES(snapshot_id);`),
		DeleteTracks: prepareStmt(tx, `
			DELETE FROM tracks WHERE snapshot_id = ?;`),
		GetSnapshotId: prepareStmt(tx, `
			SELECT snapshot_id FROM metadata;`),
		UpdateSnapshotId: prepareStmt(tx, `
			UPDATE metadata SET snapshot_id = ?;`),
	}
}

func prepareStmt(tx *sql.Tx, query string) *sql.Stmt {
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	return stmt
}

func RollBackTx(itx *ImporterTx) error {
	log.Println("INFO - Attempting to roll back tx")
	err := itx.tx.Rollback()
	if err != nil {
		return err
	}
	log.Println("INFO - Transaction rolled back succesfully")
	return nil
}

func Commit(itx *ImporterTx) error {
	err := itx.tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
