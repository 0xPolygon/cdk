package db

import (
	"testing"

	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"
)

func Test_checkMigrations(t *testing.T) {
	embedMigration := embedMigrations[AggregatorMigrationName]
	migrationSource := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: embedMigration,
	}

	_, err := migrationSource.FileSystem.ReadFile("migrations/0001.sql")
	assert.NoError(t, err)
}
