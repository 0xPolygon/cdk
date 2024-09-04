package migrations

import (
	"strings"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/log"
	treeMigrations "github.com/0xPolygon/cdk/tree/migrations"
	migrate "github.com/rubenv/sql-migrate"

	_ "embed"
)

const (
	upDownSeparator      = "-- +migrate Up"
	RollupExitTreePrefix = "rollup_exit_"
	L1InfoTreePrefix     = "l1_info_"
)

//go:embed l1infotreesync0001.sql
var mig001 string
var mig001splitted = strings.Split(mig001, upDownSeparator)

var Migrations = &migrate.MemoryMigrationSource{
	Migrations: []*migrate.Migration{
		{
			Id:   "l1infotreesync0001",
			Up:   []string{mig001splitted[1]},
			Down: []string{mig001splitted[0]},
		},
	},
}

func RunMigrations(dbPath string) error {
	migs := []*migrate.Migration{}
	retMigs := treeMigrations.MigrationsWithPrefix(RollupExitTreePrefix)
	migs = append(migs, retMigs...)
	l1InfoMigs := treeMigrations.MigrationsWithPrefix(L1InfoTreePrefix)
	migs = append(migs, l1InfoMigs...)
	migs = append(migs, Migrations.Migrations...)
	for _, m := range migs {
		log.Debugf("%+v", m.Id)
	}
	return db.RunMigrations(dbPath, &migrate.MemoryMigrationSource{Migrations: migs})
}
