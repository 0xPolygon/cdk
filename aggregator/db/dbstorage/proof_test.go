package dbstorage

import (
	"context"
	"math"
	"testing"

	"github.com/0xPolygon/cdk/aggregator/db"
	"github.com/0xPolygon/cdk/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Proof(t *testing.T) {
	dbPath := "file::memory:?cache=shared"
	err := db.RunMigrationsUp(dbPath, db.AggregatorMigrationName)
	assert.NoError(t, err)

	ctx := context.Background()

	DBStorage, err := NewDBStorage(dbPath)
	assert.NoError(t, err)

	dbtxer, err := DBStorage.BeginTx(ctx, nil)
	require.NoError(t, err)

	exists, err := DBStorage.CheckProofExistsForBatch(ctx, 1, dbtxer)
	assert.NoError(t, err)
	assert.False(t, exists)

	proof := state.Proof{
		BatchNumber:      1,
		BatchNumberFinal: 1,
		GeneratingSince:  nil,
	}

	err = DBStorage.AddGeneratedProof(ctx, &proof, dbtxer)
	assert.NoError(t, err)

	err = DBStorage.AddSequence(ctx, state.Sequence{FromBatchNumber: 1, ToBatchNumber: 1}, dbtxer)
	assert.NoError(t, err)

	contains, err := DBStorage.CheckProofContainsCompleteSequences(ctx, &proof, dbtxer)
	assert.NoError(t, err)
	assert.True(t, contains)

	proof2, err := DBStorage.GetProofReadyToVerify(ctx, 0, dbtxer)
	assert.NoError(t, err)
	assert.NotNil(t, proof2)

	proof = state.Proof{
		BatchNumber:      1,
		BatchNumberFinal: 2,
	}

	err = DBStorage.UpdateGeneratedProof(ctx, &proof, dbtxer)
	assert.NoError(t, err)

	err = DBStorage.DeleteGeneratedProofs(ctx, 1, math.MaxInt, dbtxer)
	assert.NoError(t, err)

	err = DBStorage.CleanupGeneratedProofs(ctx, 1, dbtxer)
	assert.NoError(t, err)

	_, err = DBStorage.CleanupLockedProofs(ctx, "1s", dbtxer)
	assert.NoError(t, err)

	err = DBStorage.DeleteUngeneratedProofs(ctx, dbtxer)
	assert.NoError(t, err)
}
