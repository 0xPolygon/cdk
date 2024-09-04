package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// NewSQLiteDB creates a new SQLite DB
func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	initMeddler()
	return sql.Open("sqlite", dbPath)
}
