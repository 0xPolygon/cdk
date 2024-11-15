package db

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const (
	UniqueConstrain = 1555
)

var (
	ErrNotFound = errors.New("not found")
)

// NewSQLiteDB creates a new SQLite DB
func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	return sql.Open("sqlite3", fmt.Sprintf("file:%s?_txlock=exclusive&_foreign_keys=on", dbPath))
	// db,err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", dbPath))
	// if err != nil {
	// 	return nil, err
	// }
	// _, err = db.Exec(`
	// 	PRAGMA foreign_keys = ON;
	// 	pragma journal_mode = WAL;
	// 	pragma synchronous = normal;
	// 	pragma journal_size_limit  = 6144000;
	// `)
	// return db, err
}

func ReturnErrNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
