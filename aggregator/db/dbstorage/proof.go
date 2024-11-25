package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/state"
)

// CheckProofExistsForBatch checks if the batch is already included in any proof
func (d *DBStorage) CheckProofExistsForBatch(ctx context.Context, batchNumber uint64, dbTx db.Txer) (bool, error) {
	const checkProofExistsForBatchSQL = `
		SELECT EXISTS (SELECT 1 FROM proof p WHERE $1 >= p.batch_num AND $1 <= p.batch_num_final)
		`
	e := d.getExecQuerier(dbTx)
	var exists bool
	err := e.QueryRow(checkProofExistsForBatchSQL, batchNumber).Scan(&exists)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return exists, err
	}
	return exists, nil
}

// CheckProofContainsCompleteSequences checks if a recursive proof contains complete sequences
func (d *DBStorage) CheckProofContainsCompleteSequences(
	ctx context.Context, proof *state.Proof, dbTx db.Txer,
) (bool, error) {
	const getProofContainsCompleteSequencesSQL = `
		SELECT EXISTS (SELECT 1 FROM sequence s1 WHERE s1.from_batch_num = $1) AND
			   EXISTS (SELECT 1 FROM sequence s2 WHERE s2.to_batch_num = $2)
		`
	e := d.getExecQuerier(dbTx)
	var exists bool
	err := e.QueryRow(getProofContainsCompleteSequencesSQL, proof.BatchNumber, proof.BatchNumberFinal).Scan(&exists)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return exists, err
	}
	return exists, nil
}

// GetProofReadyToVerify return the proof that is ready to verify
func (d *DBStorage) GetProofReadyToVerify(
	ctx context.Context, lastVerfiedBatchNumber uint64, dbTx db.Txer,
) (*state.Proof, error) {
	const getProofReadyToVerifySQL = `
		SELECT 
			p.batch_num, 
			p.batch_num_final,
			p.proof,
			p.proof_id,
			p.input_prover,
			p.prover,
			p.prover_id,
			p.generating_since,
			p.created_at,
			p.updated_at
		FROM proof p
		WHERE batch_num = $1 AND generating_since IS NULL AND
			EXISTS (SELECT 1 FROM sequence s1 WHERE s1.from_batch_num = p.batch_num) AND
			EXISTS (SELECT 1 FROM sequence s2 WHERE s2.to_batch_num = p.batch_num_final)		
		`

	var proof = &state.Proof{}

	e := d.getExecQuerier(dbTx)
	row := e.QueryRow(getProofReadyToVerifySQL, lastVerfiedBatchNumber+1)

	var (
		generatingSince *uint64
		createdAt       *uint64
		updatedAt       *uint64
	)
	err := row.Scan(
		&proof.BatchNumber, &proof.BatchNumberFinal, &proof.Proof, &proof.ProofID,
		&proof.InputProver, &proof.Prover, &proof.ProverID, &generatingSince,
		&createdAt, &updatedAt,
	)

	if generatingSince != nil {
		timeSince := time.Unix(int64(*generatingSince), 0)
		proof.GeneratingSince = &timeSince
	}

	if createdAt != nil {
		proof.CreatedAt = time.Unix(int64(*createdAt), 0)
	}

	if updatedAt != nil {
		proof.UpdatedAt = time.Unix(int64(*updatedAt), 0)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, state.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return proof, err
}

// GetProofsToAggregate return the next to proof that it is possible to aggregate
func (d *DBStorage) GetProofsToAggregate(ctx context.Context, dbTx db.Txer) (*state.Proof, *state.Proof, error) {
	var (
		proof1 = &state.Proof{}
		proof2 = &state.Proof{}
	)

	// TODO: add comments to explain the query
	const getProofsToAggregateSQL = `
		SELECT 
			p1.batch_num as p1_batch_num, 
			p1.batch_num_final as p1_batch_num_final, 
			p1.proof as p1_proof,	
			p1.proof_id as p1_proof_id, 
			p1.input_prover as p1_input_prover, 
			p1.prover as p1_prover,
			p1.prover_id as p1_prover_id,
			p1.generating_since as p1_generating_since,
			p1.created_at as p1_created_at,
			p1.updated_at as p1_updated_at,
			p2.batch_num as p2_batch_num, 
			p2.batch_num_final as p2_batch_num_final, 
			p2.proof as p2_proof,	
			p2.proof_id as p2_proof_id, 
			p2.input_prover as p2_input_prover, 
			p2.prover as p2_prover,
			p2.prover_id as p2_prover_id,
			p2.generating_since as p2_generating_since,
			p2.created_at as p2_created_at,
			p2.updated_at as p2_updated_at
		FROM proof p1 INNER JOIN proof p2 ON p1.batch_num_final = p2.batch_num - 1
		WHERE p1.generating_since IS NULL AND p2.generating_since IS NULL AND 
		 	  p1.proof IS NOT NULL AND p2.proof IS NOT NULL AND
			  (
					EXISTS (
					SELECT 1 FROM sequence s
					WHERE p1.batch_num >= s.from_batch_num AND p1.batch_num <= s.to_batch_num AND
						p1.batch_num_final >= s.from_batch_num AND p1.batch_num_final <= s.to_batch_num AND
						p2.batch_num >= s.from_batch_num AND p2.batch_num <= s.to_batch_num AND
						p2.batch_num_final >= s.from_batch_num AND p2.batch_num_final <= s.to_batch_num
					)
					OR
					(
						EXISTS ( SELECT 1 FROM sequence s WHERE p1.batch_num = s.from_batch_num) AND
						EXISTS ( SELECT 1 FROM sequence s WHERE p1.batch_num_final = s.to_batch_num) AND
						EXISTS ( SELECT 1 FROM sequence s WHERE p2.batch_num = s.from_batch_num) AND
						EXISTS ( SELECT 1 FROM sequence s WHERE p2.batch_num_final = s.to_batch_num)
					)
				)
		ORDER BY p1.batch_num ASC
		LIMIT 1
		`

	e := d.getExecQuerier(dbTx)
	row := e.QueryRow(getProofsToAggregateSQL)

	var (
		generatingSince1, generatingSince2 *uint64
		createdAt1, createdAt2             *uint64
		updatedAt1, updatedAt2             *uint64
	)

	err := row.Scan(
		&proof1.BatchNumber, &proof1.BatchNumberFinal, &proof1.Proof, &proof1.ProofID,
		&proof1.InputProver, &proof1.Prover, &proof1.ProverID, &generatingSince1,
		&createdAt1, &updatedAt1,
		&proof2.BatchNumber, &proof2.BatchNumberFinal, &proof2.Proof, &proof2.ProofID,
		&proof2.InputProver, &proof2.Prover, &proof2.ProverID, &generatingSince2,
		&createdAt1, &updatedAt1,
	)

	if generatingSince1 != nil {
		timeSince1 := time.Unix(int64(*generatingSince1), 0)
		proof1.GeneratingSince = &timeSince1
	}

	if generatingSince2 != nil {
		timeSince2 := time.Unix(int64(*generatingSince2), 0)
		proof2.GeneratingSince = &timeSince2
	}

	if createdAt1 != nil {
		proof1.CreatedAt = time.Unix(int64(*createdAt1), 0)
	}

	if createdAt2 != nil {
		proof2.CreatedAt = time.Unix(int64(*createdAt2), 0)
	}

	if updatedAt1 != nil {
		proof1.UpdatedAt = time.Unix(int64(*updatedAt1), 0)
	}

	if updatedAt2 != nil {
		proof2.UpdatedAt = time.Unix(int64(*updatedAt2), 0)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, state.ErrNotFound
	} else if err != nil {
		return nil, nil, err
	}

	return proof1, proof2, err
}

// AddGeneratedProof adds a generated proof to the storage
func (d *DBStorage) AddGeneratedProof(ctx context.Context, proof *state.Proof, dbTx db.Txer) error {
	const addGeneratedProofSQL = `
		INSERT INTO proof (
			batch_num, batch_num_final, proof, proof_id, input_prover, prover, 
			prover_id, generating_since, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`
	e := d.getExecQuerier(dbTx)
	now := time.Now().UTC().Round(time.Microsecond)

	var (
		generatingSince *uint64
		createdAt       *uint64
		updatedAt       *uint64
	)

	if proof.GeneratingSince != nil {
		generatingSince = new(uint64)
		*generatingSince = uint64(proof.GeneratingSince.Unix())
	}

	if !proof.CreatedAt.IsZero() {
		createdAt = new(uint64)
		*createdAt = uint64(proof.CreatedAt.Unix())
	} else {
		createdAt = new(uint64)
		*createdAt = uint64(now.Unix())
	}

	if !proof.UpdatedAt.IsZero() {
		updatedAt = new(uint64)
		*updatedAt = uint64(proof.UpdatedAt.Unix())
	} else {
		updatedAt = new(uint64)
		*updatedAt = uint64(now.Unix())
	}

	_, err := e.Exec(
		addGeneratedProofSQL, proof.BatchNumber, proof.BatchNumberFinal, proof.Proof, proof.ProofID,
		proof.InputProver, proof.Prover, proof.ProverID, generatingSince, createdAt, updatedAt,
	)
	return err
}

// UpdateGeneratedProof updates a generated proof in the storage
func (d *DBStorage) UpdateGeneratedProof(ctx context.Context, proof *state.Proof, dbTx db.Txer) error {
	const updateGeneratedProofSQL = `
	UPDATE proof 
	SET proof = $3, 
		proof_id = $4, 
		input_prover = $5, 
		prover = $6, 
		prover_id = $7, 
		generating_since = $8, 
		updated_at = $9 
	WHERE batch_num = $1 
		AND batch_num_final = $2
	`
	e := d.getExecQuerier(dbTx)
	now := time.Now().UTC().Round(time.Microsecond)

	var (
		generatingSince *uint64
		updatedAt       *uint64
	)

	if proof.GeneratingSince != nil {
		generatingSince = new(uint64)
		*generatingSince = uint64(proof.GeneratingSince.Unix())
	}

	if !proof.UpdatedAt.IsZero() {
		updatedAt = new(uint64)
		*updatedAt = uint64(proof.UpdatedAt.Unix())
	} else {
		updatedAt = new(uint64)
		*updatedAt = uint64(now.Unix())
	}
	_, err := e.Exec(
		updateGeneratedProofSQL, proof.Proof, proof.ProofID, proof.InputProver,
		proof.Prover, proof.ProverID, generatingSince, updatedAt, proof.BatchNumber, proof.BatchNumberFinal,
	)
	return err
}

// DeleteGeneratedProofs deletes from the storage the generated proofs falling
// inside the batch numbers range.
func (d *DBStorage) DeleteGeneratedProofs(
	ctx context.Context, batchNumber uint64, batchNumberFinal uint64, dbTx db.Txer,
) error {
	const deleteGeneratedProofSQL = "DELETE FROM proof WHERE batch_num >= $1 AND batch_num_final <= $2"
	e := d.getExecQuerier(dbTx)
	_, err := e.Exec(deleteGeneratedProofSQL, batchNumber, batchNumberFinal)
	return err
}

// CleanupGeneratedProofs deletes from the storage the generated proofs up to
// the specified batch number included.
func (d *DBStorage) CleanupGeneratedProofs(ctx context.Context, batchNumber uint64, dbTx db.Txer) error {
	const deleteGeneratedProofSQL = "DELETE FROM proof WHERE batch_num_final <= $1"
	e := d.getExecQuerier(dbTx)
	_, err := e.Exec(deleteGeneratedProofSQL, batchNumber)
	return err
}

// CleanupLockedProofs deletes from the storage the proofs locked in generating
// state for more than the provided threshold.
func (d *DBStorage) CleanupLockedProofs(ctx context.Context, duration string, dbTx db.Txer) (int64, error) {
	seconds, err := convertDurationToSeconds(duration)
	if err != nil {
		return 0, err
	}

	difference := time.Now().Unix() - seconds

	sql := fmt.Sprintf("DELETE FROM proof WHERE generating_since is not null and generating_since < %d", difference)
	e := d.getExecQuerier(dbTx)
	ct, err := e.Exec(sql)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected()
}

// DeleteUngeneratedProofs deletes ungenerated proofs.
// This method is meant to be use during aggregator boot-up sequence
func (d *DBStorage) DeleteUngeneratedProofs(ctx context.Context, dbTx db.Txer) error {
	const deleteUngeneratedProofsSQL = "DELETE FROM proof WHERE generating_since IS NOT NULL"
	e := d.getExecQuerier(dbTx)
	_, err := e.Exec(deleteUngeneratedProofsSQL)
	return err
}

func convertDurationToSeconds(duration string) (int64, error) {
	// Parse the duration using time.ParseDuration
	parsedDuration, err := time.ParseDuration(duration)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format: %w", err)
	}

	// Return the duration in seconds
	return int64(parsedDuration.Seconds()), nil
}
