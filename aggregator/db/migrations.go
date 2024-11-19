package db

import (
	"embed"
	"fmt"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/log"
	migrate "github.com/rubenv/sql-migrate"
)

const (
	// AggregatorMigrationName is the name of the migration used to associate with the migrations dir
	AggregatorMigrationName = "aggregator-db"
)

var (
	//go:embed migrations/*.sql
	embedAggregatorMigrations embed.FS

	// embedMigrations is a map of migrations with the name
	embedMigrations = map[string]embed.FS{}
)

func init() {
	embedMigrations[AggregatorMigrationName] = embedAggregatorMigrations
}

// RunMigrationsUp runs migrate-up for the given config.
func RunMigrationsUp(dbPath string, name string) error {
	log.Info("running migrations up")

	return runMigrations(dbPath, name, migrate.Up)
}

// CheckMigrations runs migrate-up for the given config.
func CheckMigrations(dbPath string, name string) error {
	return checkMigrations(dbPath, name)
}

// RunMigrationsDown runs migrate-down for the given config.
func RunMigrationsDown(dbPath string, name string) error {
	log.Info("running migrations down")

	return runMigrations(dbPath, name, migrate.Down)
}

// runMigrations will execute pending migrations if needed to keep
// the database updated with the latest changes in either direction,
// up or down.
func runMigrations(dbPath string, name string, direction migrate.MigrationDirection) error {
	db, err := db.NewSQLiteDB(dbPath)
	if err != nil {
		return err
	}

	embedMigration, ok := embedMigrations[name]
	if !ok {
		return fmt.Errorf("migration not found with name: %v", name)
	}

	var migrations = &migrate.EmbedFileSystemMigrationSource{
		FileSystem: embedMigration,
		Root:       "migrations",
	}

	nMigrations, err := migrate.Exec(db, "sqlite3", migrations, direction)
	if err != nil {
		return err
	}

	log.Info("successfully ran ", nMigrations, " migrations")

	return nil
}

func checkMigrations(dbPath string, name string) error {
	db, err := db.NewSQLiteDB(dbPath)
	if err != nil {
		return err
	}

	embedMigration, ok := embedMigrations[name]
	if !ok {
		return fmt.Errorf("migration not found with name: %v", name)
	}

	migrationSource := &migrate.EmbedFileSystemMigrationSource{FileSystem: embedMigration}

	migrations, err := migrationSource.FindMigrations()
	if err != nil {
		log.Errorf("error getting migrations from source: %v", err)

		return err
	}

	var expected int
	for _, migration := range migrations {
		if len(migration.Up) != 0 {
			expected++
		}
	}

	var actual int
	query := `SELECT COUNT(1) FROM public.gorp_migrations`
	err = db.QueryRow(query).Scan(&actual)
	if err != nil {
		log.Error("error getting migrations count: ", err)

		return err
	}
	if expected == actual {
		log.Infof("Found %d migrations as expected", actual)
	} else {
		return fmt.Errorf(
			"error the component needs to run %d migrations before starting. DB only contains %d migrations",
			expected, actual,
		)
	}

	return nil
}
