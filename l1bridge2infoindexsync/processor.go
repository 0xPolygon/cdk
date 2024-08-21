package l1bridge2infoindexsync

import (
	"context"
	"errors"
	"fmt"

	"github.com/0xPolygon/cdk/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

const (
	lastProcessedTable = "l1bridge2infoindexsync-lastProcessed"
	relationTable      = "l1bridge2infoindexsync-relation"
)

var (
	lastProcessedKey = []byte("lp")
	ErrNotFound      = errors.New("not found")
)

type processor struct {
	db kv.RwDB
}

type bridge2L1InfoRelation struct {
	bridgeIndex     uint32
	l1InfoTreeIndex uint32
}

type lastProcessed struct {
	block uint64
	index uint32
}

func (lp *lastProcessed) MarshalBinary() ([]byte, error) {
	return append(common.Uint64ToBytes(lp.block), common.Uint32ToBytes(lp.index)...), nil
}

func (lp *lastProcessed) UnmarshalBinary(data []byte) error {
	if len(data) != 12 {
		return fmt.Errorf("expected len %d, actual len %d", 12, len(data))
	}
	lp.block = common.BytesToUint64(data[:8])
	lp.index = common.BytesToUint32(data[8:])
	return nil
}

func newProcessor(dbPath string) (*processor, error) {
	tableCfgFunc := func(defaultBuckets kv.TableCfg) kv.TableCfg {
		return kv.TableCfg{
			lastProcessedTable: {},
			relationTable:      {},
		}
	}
	db, err := mdbx.NewMDBX(nil).
		Path(dbPath).
		WithTableCfg(tableCfgFunc).
		Open()
	if err != nil {
		return nil, err
	}
	return &processor{
		db: db,
	}, nil
}

// GetLastProcessedBlockAndL1InfoTreeIndex returns the last processed block oby the processor, including blocks
// that don't have events
func (p *processor) GetLastProcessedBlockAndL1InfoTreeIndex(ctx context.Context) (uint64, uint32, error) {
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()
	return p.getLastProcessedBlockAndL1InfoTreeIndexWithTx(tx)
}

func (p *processor) getLastProcessedBlockAndL1InfoTreeIndexWithTx(tx kv.Tx) (uint64, uint32, error) {
	if lastProcessedBytes, err := tx.GetOne(lastProcessedTable, lastProcessedKey); err != nil {
		return 0, 0, err
	} else if lastProcessedBytes == nil {
		return 0, 0, nil
	} else {
		lp := &lastProcessed{}
		if err := lp.UnmarshalBinary(lastProcessedBytes); err != nil {
			return 0, 0, err
		}
		return lp.block, lp.index, nil
	}
}

func (p *processor) updateLastProcessedBlockAndL1InfoTreeIndex(ctx context.Context, blockNum uint64, index uint32) error {
	tx, err := p.db.BeginRw(ctx)
	if err != nil {
		return err
	}
	if err := p.updateLastProcessedBlockAndL1InfoTreeIndexWithTx(tx, blockNum, index); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (p *processor) updateLastProcessedBlockAndL1InfoTreeIndexWithTx(tx kv.RwTx, blockNum uint64, index uint32) error {
	lp := &lastProcessed{
		block: blockNum,
		index: index,
	}
	value, err := lp.MarshalBinary()
	if err != nil {
		return err
	}
	return tx.Put(lastProcessedTable, lastProcessedKey, value)
}

func (p *processor) processUntilBlock(ctx context.Context, lastProcessedBlock uint64, relations []bridge2L1InfoRelation) error {
	tx, err := p.db.BeginRw(ctx)
	if err != nil {
		return err
	}

	if len(relations) == 0 {
		_, lastIndex, err := p.getLastProcessedBlockAndL1InfoTreeIndexWithTx(tx)
		if err != nil {
			tx.Rollback()
			return err
		}
		if err := p.updateLastProcessedBlockAndL1InfoTreeIndexWithTx(
			tx,
			lastProcessedBlock,
			lastIndex,
		); err != nil {
			tx.Rollback()
			return err
		}
		return tx.Commit()
	}

	for _, relation := range relations {
		if _, err := p.getL1InfoTreeIndexByBridgeIndexWithTx(tx, relation.bridgeIndex); err != ErrNotFound {
			// Note that indexes could be repeated as the L1 Info tree update can be produced by a rollup and not mainnet.
			// Hence if the index already exist, do not update as it's better to have the lowest index possible for the relation
			continue
		}
		if err := tx.Put(
			relationTable,
			common.Uint32ToBytes(relation.bridgeIndex),
			common.Uint32ToBytes(relation.l1InfoTreeIndex),
		); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := p.updateLastProcessedBlockAndL1InfoTreeIndexWithTx(
		tx,
		lastProcessedBlock,
		relations[len(relations)-1].l1InfoTreeIndex,
	); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (p *processor) getL1InfoTreeIndexByBridgeIndex(ctx context.Context, depositCount uint32) (uint32, error) {
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	return p.getL1InfoTreeIndexByBridgeIndexWithTx(tx, depositCount)
}

func (p *processor) getL1InfoTreeIndexByBridgeIndexWithTx(tx kv.Tx, depositCount uint32) (uint32, error) {
	indexBytes, err := tx.GetOne(relationTable, common.Uint32ToBytes(depositCount))
	if err != nil {
		return 0, err
	}
	if indexBytes == nil {
		return 0, ErrNotFound
	}
	return common.BytesToUint32(indexBytes), nil
}
