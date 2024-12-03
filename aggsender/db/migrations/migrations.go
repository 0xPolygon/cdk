package migrations

import (
	"database/sql"
	_ "embed"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/db/types"
	"github.com/0xPolygon/cdk/log"
)

//go:embed 0001.sql
var mig001 string

func RunMigrations(logger *log.Logger, database *sql.DB) error {
	migrations := []types.Migration{
		{
			ID:  "0001",
			SQL: mig001,
		},
	}

	return db.RunMigrationsDB(logger, database, migrations)
}
