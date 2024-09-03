package migrations

import (
	"strings"

	"github.com/0xPolygon/cdk/db"
	treeMigrations "github.com/0xPolygon/cdk/tree/migrations"
	migrate "github.com/rubenv/sql-migrate"

	_ "embed"
)

const upDownSeparator = "-- +migrate Up"

//go:embed bridgesync0001.sql
var mig001 string
var mig001splitted = strings.Split(mig001, upDownSeparator)

var bridgeMigrations = &migrate.MemoryMigrationSource{
	Migrations: []*migrate.Migration{
		{
			Id:   "bridgesync001",
			Up:   []string{mig001splitted[1]},
			Down: []string{mig001splitted[0]},
		},
	},
}

func RunMigrations(dbPath string) error {
	bridgeMigrations.Migrations = append(
		bridgeMigrations.Migrations,
		treeMigrations.Migrations.Migrations...,
	)
	return db.RunMigrations(dbPath, bridgeMigrations)
}
