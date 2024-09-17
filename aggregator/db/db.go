package db

import (
	"context"
	"fmt"

	"github.com/0xPolygon/cdk/log"
	"github.com/jackc/pgx/v4/pgxpool"
)

// NewSQLDB creates a new SQL DB
func NewSQLDB(logger *log.Logger, cfg Config) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s:%s/%s?pool_max_conns=%d",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.MaxConns))
	if err != nil {
		logger.Errorf("Unable to parse DB config: %v\n", err)
		return nil, err
	}

	if cfg.EnableLog {
		config.ConnConfig.Logger = dbLoggerImpl{}
	}

	conn, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		logger.Errorf("Unable to connect to database: %v\n", err)
		return nil, err
	}

	return conn, nil
}
