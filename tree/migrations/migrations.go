package migrations

import (
	_ "embed"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/db/types"
)

//go:embed tree0001.sql
var mig001 string

var Migrations = []types.Migration{
	{
		ID:  "tree001",
		SQL: mig001,
	},
}

func RunMigrations(dbPath string) error {
	return db.RunMigrations(dbPath, Migrations)
}
