package lastgersync

import (
	"context"
	"errors"
	"fmt"

	"github.com/0xPolygon/cdk/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

const (
	lastProcessedTable = "lastgersync-lastProcessed"
	gerTable           = "lastgersync-relation"
)

var (
	lastProcessedKey = []byte("lp")
	ErrNotFound      = errors.New("not found")
)

type processor struct {
	db kv.RwDB
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
		cfg := kv.TableCfg{
			lastProcessedTable: {},
			gerTable:           {},
		}
		return cfg
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
