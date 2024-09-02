package db

import (
	"fmt"

	"github.com/0xPolygon/cdk/log"
	migrate "github.com/rubenv/sql-migrate"
	_ "modernc.org/sqlite"
)

// RunMigrations will execute pending migrations if needed to keep
// the database updated with the latest changes in either direction,
// up or down.
func RunMigrations(dbPath string, migrations migrate.MigrationSource) error {
	db, err := NewSQLiteDB(dbPath)
	if err != nil {
		return fmt.Errorf("error creating DB %w", err)
	}

	nMigrations, err := migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("error executing migration %w", err)
	}

	log.Infof("successfully ran %d migrations", nMigrations)
	return nil
}
