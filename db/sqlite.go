package db

import (
	"database/sql"

	"github.com/russross/meddler"
	_ "modernc.org/sqlite"
)

// NewSQLiteDB creates a new SQLite DB
func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	meddler.Default = meddler.SQLite
	meddler.Mapper = meddler.SnakeCase
	return sql.Open("sqlite", dbPath)
}
