package migrations

import (
	_ "embed"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/db/types"
	treeMigrations "github.com/0xPolygon/cdk/tree/migrations"
)

//go:embed bridgesync0001.sql
var mig001 string

func RunMigrations(dbPath string) error {
	migrations := []types.Migration{
		{
			ID:  "bridgesync0001",
			SQL: mig001,
		},
	}
	migrations = append(migrations, treeMigrations.Migrations...)
	return db.RunMigrations(dbPath, migrations)
}
