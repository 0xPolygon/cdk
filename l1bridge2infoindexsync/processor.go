package l1bridge2infoindexsync

import (
	"context"
	"errors"

	"github.com/0xPolygon/cdk/common"
	"github.com/ledgerwatch/erigon-lib/kv"
)

var (
	lastBlokcKey = []byte("lb")
)

type processor struct {
	db             kv.RwDB
	lastBlockTable string
}

type bridge2L1InfoRelation struct {
	bridgeIndex     uint32
	l1InfoTreeIndex uint32
}

// GetLastProcessedBlock returns the last processed block oby the processor, including blocks
// that don't have events
func (p *processor) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	return p.getLastProcessedBlockWithTx(tx)
}

func (p *processor) getLastProcessedBlockWithTx(tx kv.Tx) (uint64, error) {
	if blockNumBytes, err := tx.GetOne(p.lastBlockTable, lastBlokcKey); err != nil {
		return 0, err
	} else if blockNumBytes == nil {
		return 0, nil
	} else {
		return common.BytesToUint64(blockNumBytes), nil
	}
}

func (p *processor) getLastL1InfoTreeIndexProcessed(ctx context.Context) (uint32, error) {
	return 0, errors.New("not implemented")
}

func (p *processor) updateLastProcessedBlock(ctx context.Context, blockNum uint64) error {
	return errors.New("not implemented")
}

func (p *processor) addBridge2L1InfoRelations(ctx context.Context, lastProcessedBlcok uint64, relations []bridge2L1InfoRelation) error {
	// Note that indexes could be repeated as the L1 Info tree update can be produced by a rollup and not mainnet.
	// Hence if the index already exist, do not update as it's better to have the lowest indes possible for the relation
	return errors.New("not implemented")
}
