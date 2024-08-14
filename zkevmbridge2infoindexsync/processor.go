package zkevmbridge2infoindexsync

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

const (
	// block -> batch
	batchTable = "batchsync-batch"
)

var ErrNotFound = errors.New("not found")

type processor struct {
	db kv.RwDB
}

func newProcessor(dbPath string) (*processor, error) {
	tableCfgFunc := func(defaultBuckets kv.TableCfg) kv.TableCfg {
		cfg := kv.TableCfg{
			batchTable: {},
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

func (p *processor) ProcessBlock(ctx context.Context, block sync.Block) error {
	if len(block.Events) != 1 {
		return fmt.Errorf("unexpected len. Expected 1, actual %d", len(block.Events))
	}
	tx, err := p.db.BeginRw(ctx)
	if err != nil {
		return err
	}
	batchNum := block.Events[0].(uint64)
	err = tx.Put(batchTable, common.Uint64ToBytes(block.Num), common.Uint64ToBytes(batchNum))
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (p *processor) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	iter, err := tx.RangeDescend(batchTable, common.Uint64ToBytes(math.MaxUint16), common.Uint64ToBytes(0), 1)
	if err != nil {
		return 0, err
	}
	k, _, err := iter.Next()
	if err != nil {
		return 0, err
	}
	if k == nil {
		return 0, ErrNotFound
	}

	return common.BytesToUint64(k), nil
}

func (p *processor) Reorg(ctx context.Context, firstReorgedBlock uint64) error {
	tx, err := p.db.BeginRw(ctx)
	if err != nil {
		return err
	}

	iter, err := tx.Range(batchTable, common.Uint64ToBytes(firstReorgedBlock), nil)
	if err != nil {
		tx.Rollback()
		return err
	}
	for k, _, err := iter.Next(); k != nil; k, _, err = iter.Next() {
		if err != nil {
			tx.Rollback()
			return err
		}
		err = tx.Delete(batchTable, k)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (p *processor) getBatchNum(ctx context.Context, blockNum uint64) (uint64, error) {
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	v, err := tx.GetOne(batchTable, common.Uint64ToBytes(blockNum))
	if err != nil {
		return 0, err
	}
	if v == nil {
		return 0, ErrNotFound
	}
	return common.BytesToUint64(v), nil
}
