package l1infotreesync

import (
	"errors"
	"fmt"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/log"
	treeTypes "github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/russross/meddler"
)

func (p *processor) processVerifyBatches(tx db.Txer, blockNumber uint64, event *VerifyBatches) error {
	if event == nil {
		return fmt.Errorf("processVerifyBatches: event is nil")
	}
	if tx == nil {
		return fmt.Errorf("processVerifyBatches: tx is nil, is mandatory to pass a tx")
	}
	log.Debugf("VerifyBatches: rollupExitTree.UpsertLeaf (blockNumber=%d, event=%s)", blockNumber, event.String())
	// If ExitRoot is zero if the leaf doesnt exists doesnt change the root of tree.
	//  	if leaf already exists doesn't make sense to 'empty' the leaf, so we keep previous value
	if event.ExitRoot == (common.Hash{}) {
		log.Infof("skipping VerifyBatches event with empty ExitRoot (blockNumber=%d, event=%s)", blockNumber, event.String())
		return nil
	}
	isNewLeaf, err := p.isNewValueForRollupExitTree(tx, event)
	if err != nil {
		return fmt.Errorf("isNewValueForrollupExitTree. err: %w", err)
	}
	if !isNewLeaf {
		log.Infof("skipping VerifyBatches event with same ExitRoot (blockNumber=%d, event=%s)", blockNumber, event.String())
		return nil
	}
	log.Infof("UpsertLeaf VerifyBatches event (blockNumber=%d, event=%s)", blockNumber, event.String())
	newRoot, err := p.rollupExitTree.UpsertLeaf(tx, blockNumber, event.BlockPosition, treeTypes.Leaf{
		Index: event.RollupID - 1,
		Hash:  event.ExitRoot,
	})
	if err != nil {
		return fmt.Errorf("error rollupExitTree.UpsertLeaf. err: %w", err)
	}
	verifyBatches := event
	verifyBatches.BlockNumber = blockNumber
	verifyBatches.RollupExitRoot = newRoot
	if err = meddler.Insert(tx, "verify_batches", verifyBatches); err != nil {
		return fmt.Errorf("error inserting verify_batches. err: %w", err)
	}
	return nil
}

func (p *processor) isNewValueForRollupExitTree(tx db.Querier, event *VerifyBatches) (bool, error) {
	currentRoot, err := p.rollupExitTree.GetLastRoot(tx)
	if err != nil && errors.Is(err, db.ErrNotFound) {
		// The tree is empty, so is a new value for sure
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("error rollupExitTree.GetLastRoot. err: %w", err)
	}
	leaf, err := p.rollupExitTree.GetLeaf(tx, event.RollupID-1, currentRoot.Hash)
	if err != nil && errors.Is(err, db.ErrNotFound) {
		// The leaf doesn't exist, so is a new value
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("error rollupExitTree.GetLeaf. err: %w", err)
	}
	return leaf != event.ExitRoot, nil
}

func (p *processor) GetLastVerifiedBatches(rollupID uint32) (*VerifyBatches, error) {
	verified := &VerifyBatches{}
	err := meddler.QueryRow(p.db, verified, `
		SELECT * FROM verify_batches
		WHERE rollup_id = $1
		ORDER BY block_num DESC, block_pos DESC
		LIMIT 1;
	`, rollupID)
	return verified, db.ReturnErrNotFound(err)
}

func (p *processor) GetFirstVerifiedBatches(rollupID uint32) (*VerifyBatches, error) {
	verified := &VerifyBatches{}
	err := meddler.QueryRow(p.db, verified, `
		SELECT * FROM verify_batches
		WHERE rollup_id = $1
		ORDER BY block_num ASC, block_pos ASC
		LIMIT 1;
	`, rollupID)
	return verified, db.ReturnErrNotFound(err)
}

func (p *processor) GetFirstVerifiedBatchesAfterBlock(rollupID uint32, blockNum uint64) (*VerifyBatches, error) {
	verified := &VerifyBatches{}
	err := meddler.QueryRow(p.db, verified, `
		SELECT * FROM verify_batches
		WHERE rollup_id = $1 AND block_num >= $2
		ORDER BY block_num ASC, block_pos ASC
		LIMIT 1;
	`, rollupID, blockNum)
	return verified, db.ReturnErrNotFound(err)
}
