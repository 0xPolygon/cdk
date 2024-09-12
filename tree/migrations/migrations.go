package migrations

import (
	_ "embed"
	"strings"

	"github.com/0xPolygon/cdk/db"
	migrate "github.com/rubenv/sql-migrate"
)

const (
	upDownSeparator  = "-- +migrate Up"
	dbPrefixReplacer = "/*dbprefix*/"
)

//go:embed tree0001.sql
var mig001 string
var mig001splitted = strings.Split(mig001, upDownSeparator)

var Migrations = &migrate.MemoryMigrationSource{
	Migrations: []*migrate.Migration{
		{
			Id:   "tree001",
			Up:   []string{mig001splitted[1]},
			Down: []string{mig001splitted[0]},
		},
	},
}

func RunMigrations(dbPath string) error {
	return db.RunMigrations(dbPath, Migrations)
}

func MigrationsWithPrefix(prefix string) []*migrate.Migration {
	return []*migrate.Migration{
		{
			Id:   prefix + "tree001",
			Up:   []string{strings.ReplaceAll(mig001splitted[1], dbPrefixReplacer, prefix)},
			Down: []string{strings.ReplaceAll(mig001splitted[0], dbPrefixReplacer, prefix)},
		},
	}
}
