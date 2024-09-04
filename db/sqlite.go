package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// NewSQLiteDB creates a new SQLite DB
func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	initMeddler()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`PRAGMA foreign_keys = ON`)
	return db, err
}
