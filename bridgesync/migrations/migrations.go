package migrations

import (
	"strings"

	migrate "github.com/rubenv/sql-migrate"

	_ "embed"
)

const upDownSeparator = "-- +migrate Up"

//go:embed 0001.sql
var mig001 string
var mig001splitted = strings.Split(mig001, upDownSeparator)

var Migrations = &migrate.MemoryMigrationSource{
	Migrations: []*migrate.Migration{
		{
			Id:   "001",
			Up:   []string{mig001splitted[1]},
			Down: []string{mig001splitted[0]},
		},
	},
}
