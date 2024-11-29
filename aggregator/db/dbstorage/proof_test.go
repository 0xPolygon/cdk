package dbstorage

import (
	"context"
	"math"
	"path"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/aggregator/db"
	"github.com/0xPolygon/cdk/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	proofID  = "proof_1"
	prover   = "prover_1"
	proverID = "prover_id"
)

func Test_Proof(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "Test_Proof.sqlite")
	err := db.RunMigrationsUp(dbPath, db.AggregatorMigrationName)
	assert.NoError(t, err)

	ctx := context.Background()
	now := time.Now()

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
		Proof:            "proof content",
		InputProver:      "input prover",
		ProofID:          &proofID,
		Prover:           &prover,
		ProverID:         &proofID,
		GeneratingSince:  nil,
		CreatedAt:        now,
		UpdatedAt:        now,
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

	require.Equal(t, proof.BatchNumber, proof2.BatchNumber)
	require.Equal(t, proof.BatchNumberFinal, proof2.BatchNumberFinal)
	require.Equal(t, proof.Proof, proof2.Proof)
	require.Equal(t, *proof.ProofID, *proof2.ProofID)
	require.Equal(t, proof.InputProver, proof2.InputProver)
	require.Equal(t, *proof.Prover, *proof2.Prover)
	require.Equal(t, *proof.ProverID, *proof2.ProverID)
	require.Equal(t, proof.CreatedAt.Unix(), proof2.CreatedAt.Unix())
	require.Equal(t, proof.UpdatedAt.Unix(), proof2.UpdatedAt.Unix())

	proof = state.Proof{
		BatchNumber:      1,
		BatchNumberFinal: 1,
		Proof:            "proof content",
		InputProver:      "input prover",
		ProofID:          &proofID,
		Prover:           &prover,
		ProverID:         &proofID,
		GeneratingSince:  &now,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err = DBStorage.UpdateGeneratedProof(ctx, &proof, dbtxer)
	assert.NoError(t, err)

	sequence := state.Sequence{FromBatchNumber: 3, ToBatchNumber: 4}

	proof3 := state.Proof{
		BatchNumber:      3,
		BatchNumberFinal: 3,
		GeneratingSince:  nil,
	}

	proof4 := state.Proof{
		BatchNumber:      4,
		BatchNumberFinal: 4,
		GeneratingSince:  nil,
	}

	err = DBStorage.AddSequence(ctx, sequence, dbtxer)
	assert.NoError(t, err)

	err = DBStorage.AddGeneratedProof(ctx, &proof3, dbtxer)
	assert.NoError(t, err)

	err = DBStorage.AddGeneratedProof(ctx, &proof4, dbtxer)
	assert.NoError(t, err)

	proof5, proof6, err := DBStorage.GetProofsToAggregate(ctx, dbtxer)
	assert.NoError(t, err)
	assert.NotNil(t, proof5)
	assert.NotNil(t, proof6)

	err = DBStorage.DeleteGeneratedProofs(ctx, 1, math.MaxInt, dbtxer)
	assert.NoError(t, err)

	err = DBStorage.CleanupGeneratedProofs(ctx, 1, dbtxer)
	assert.NoError(t, err)

	now = time.Now()

	proof3.GeneratingSince = &now
	proof4.GeneratingSince = &now

	err = DBStorage.AddGeneratedProof(ctx, &proof3, dbtxer)
	assert.NoError(t, err)

	err = DBStorage.AddGeneratedProof(ctx, &proof4, dbtxer)
	assert.NoError(t, err)

	time.Sleep(5 * time.Second)

	affected, err := DBStorage.CleanupLockedProofs(ctx, "4s", dbtxer)
	assert.NoError(t, err)
	require.Equal(t, int64(2), affected)

	proof5, proof6, err = DBStorage.GetProofsToAggregate(ctx, dbtxer)
	assert.EqualError(t, err, state.ErrNotFound.Error())
	assert.Nil(t, proof5)
	assert.Nil(t, proof6)

	err = DBStorage.DeleteUngeneratedProofs(ctx, dbtxer)
	assert.NoError(t, err)
}
