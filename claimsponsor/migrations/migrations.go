package migrations

import (
	_ "embed"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/db/types"
)

//go:embed claimsponsor0001.sql
var mig001 string

func RunMigrations(dbPath string) error {
	migrations := []types.Migration{
		{
			ID:  "claimsponsor0001",
			SQL: mig001,
		},
	}
	return db.RunMigrations(dbPath, migrations)
}
