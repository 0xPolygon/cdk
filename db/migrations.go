package db

import (
	"fmt"
	"strings"

	"github.com/0xPolygon/cdk/db/types"
	"github.com/0xPolygon/cdk/log"
	_ "github.com/mattn/go-sqlite3"
	migrate "github.com/rubenv/sql-migrate"
)

const (
	upDownSeparator  = "-- +migrate Up"
	dbPrefixReplacer = "/*dbprefix*/"
)

// RunMigrations will execute pending migrations if needed to keep
// the database updated with the latest changes in either direction,
// up or down.
func RunMigrations(dbPath string, migrations []types.Migration) error {
	db, err := NewSQLiteDB(dbPath)
	if err != nil {
		return fmt.Errorf("error creating DB %w", err)
	}
	migs := &migrate.MemoryMigrationSource{Migrations: []*migrate.Migration{}}
	for _, m := range migrations {
		prefixed := strings.ReplaceAll(m.SQL, dbPrefixReplacer, m.Prefix)
		splitted := strings.Split(prefixed, upDownSeparator)
		migs.Migrations = append(migs.Migrations, &migrate.Migration{
			Id:   m.Prefix + m.ID,
			Up:   []string{splitted[1]},
			Down: []string{splitted[0]},
		})
	}

	log.Debugf("running migrations:")
	for _, m := range migs.Migrations {
		log.Debugf("%+v", m.Id)
	}
	nMigrations, err := migrate.Exec(db, "sqlite3", migs, migrate.Up)
	if err != nil {
		return fmt.Errorf("error executing migration %w", err)
	}

	log.Infof("successfully ran %d migrations", nMigrations)
	return nil
}
