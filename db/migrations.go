package db

import (
	"fmt"

	"github.com/0xPolygon/cdk/log"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	migrate "github.com/rubenv/sql-migrate"
)

// RunMigrationsUp runs migrate-up for the given config.
func RunMigrationsUp(cfg Config, name string) error {
	log.Info("running migrations up")
	return runMigrations(cfg, name, migrate.Up)
}

// CheckMigrations runs migrate-up for the given config.
func CheckMigrations(cfg Config, name string) error {
	return checkMigrations(cfg, name, migrate.Up)
}

// RunMigrationsDown runs migrate-down for the given config.
func RunMigrationsDown(cfg Config, name string) error {
	log.Info("running migrations down")
	return runMigrations(cfg, name, migrate.Down)
}

// runMigrations will execute pending migrations if needed to keep
// the database updated with the latest changes in either direction,
// up or down.
func runMigrations(cfg Config, migrations migrate.MigrationSource, direction migrate.MigrationDirection) error {
	c, err := pgx.ParseConfig(fmt.Sprintf("sqlite://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name))
	if err != nil {
		return err
	}
	db := stdlib.OpenDB(*c)

	nMigrations, err := migrate.Exec(db, "sqlite", migrations, direction)
	if err != nil {
		return err
	}

	log.Infof("successfully ran %d migrations", nMigrations)
	return nil
}

func checkMigrations(cfg Config, packrName string, direction migrate.MigrationDirection) error {
	c, err := pgx.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name))
	if err != nil {
		return err
	}
	db := stdlib.OpenDB(*c)

	box, ok := packrMigrations[packrName]
	if !ok {
		return fmt.Errorf("packr box not found with name: %v", packrName)
	}

	migrationSource := &migrate.PackrMigrationSource{Box: box}
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
