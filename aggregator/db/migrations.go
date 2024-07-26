package db

import (
	"embed"
	"fmt"

	"github.com/0xPolygon/cdk/log"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	migrate "github.com/rubenv/sql-migrate"
)

const (
	// AggregatorMigrationName is the name of the migration used to associate with the migrations dir
	AggregatorMigrationName = "zkevm-aggregator-db"
)

var (
	//go:embed migrations/aggregator/*.sql
	embedAggregatorMigrations embed.FS

	// embedMigrations is a map of migrations with the name
	embedMigrations = map[string]embed.FS{}
)

func init() {
	embedMigrations[AggregatorMigrationName] = embedAggregatorMigrations
}

// RunMigrationsUp runs migrate-up for the given config.
func RunMigrationsUp(cfg Config, name string) error {
	log.Info("running migrations up")
	return runMigrations(cfg, name, migrate.Up)
}

// CheckMigrations runs migrate-up for the given config.
func CheckMigrations(cfg Config, name string) error {
	return checkMigrations(cfg, name)
}

// RunMigrationsDown runs migrate-down for the given config.
func RunMigrationsDown(cfg Config, name string) error {
	log.Info("running migrations down")
	return runMigrations(cfg, name, migrate.Down)
}

// runMigrations will execute pending migrations if needed to keep
// the database updated with the latest changes in either direction,
// up or down.
func runMigrations(cfg Config, name string, direction migrate.MigrationDirection) error {
	c, err := pgx.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name))
	if err != nil {
		return err
	}

	db := stdlib.OpenDB(*c)

	embedMigration, ok := embedMigrations[name]
	if !ok {
		return fmt.Errorf("migration not found with name: %v", name)
	}

	var migrations = &migrate.EmbedFileSystemMigrationSource{FileSystem: embedMigration, Root: "migrations/aggregator/"}
	nMigrations, err := migrate.Exec(db, "postgres", migrations, direction)
	if err != nil {
		return err
	}

	log.Info("successfully ran ", nMigrations, " migrations")
	return nil
}

func checkMigrations(cfg Config, name string) error {
	c, err := pgx.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name))
	if err != nil {
		return err
	}

	db := stdlib.OpenDB(*c)

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
		return fmt.Errorf("error the component needs to run %d migrations before starting. DB only contains %d migrations", expected, actual)
	}
	return nil
}
