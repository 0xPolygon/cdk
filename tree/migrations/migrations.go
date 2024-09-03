package migrations

import (
	"strings"

	"github.com/0xPolygon/cdk/db"
	migrate "github.com/rubenv/sql-migrate"

	_ "embed"
)

const upDownSeparator = "-- +migrate Up"

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
