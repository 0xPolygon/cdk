package migrations

import (
	_ "embed"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/db/types"
	treeMigrations "github.com/0xPolygon/cdk/tree/migrations"
)

const (
	RollupExitTreePrefix = "rollup_exit_"
	L1InfoTreePrefix     = "l1_info_"
)

//go:embed l1infotreesync0001.sql
var mig001 string

func RunMigrations(dbPath string) error {
	migrations := []types.Migration{
		{
			ID:  "l1infotreesync0001",
			SQL: mig001,
		},
	}
	for _, tm := range treeMigrations.Migrations {
		migrations = append(migrations, types.Migration{
			ID:     tm.ID,
			SQL:    tm.SQL,
			Prefix: RollupExitTreePrefix,
		})
		migrations = append(migrations, types.Migration{
			ID:     tm.ID,
			SQL:    tm.SQL,
			Prefix: L1InfoTreePrefix,
		})
	}
	return db.RunMigrations(dbPath, migrations)
}
