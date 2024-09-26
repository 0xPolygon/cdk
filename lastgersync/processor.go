package lastgersync

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

const (
	lastProcessedTable = "lastgersync-lastProcessed"
	gerTable           = "lastgersync-ger"
	blockTable         = "lastgersync-block"
)

var (
	lastProcessedKey = []byte("lp")
)

type Event struct {
	GlobalExitRoot  ethCommon.Hash
	L1InfoTreeIndex uint32
}

type blockWithGERs struct {
	// inclusive
	FirstIndex uint32
	// not inclusive
	LastIndex uint32
}

func (b *blockWithGERs) MarshalBinary() ([]byte, error) {
	return append(common.Uint32ToBytes(b.FirstIndex), common.Uint32ToBytes(b.LastIndex)...), nil
}

func (b *blockWithGERs) UnmarshalBinary(data []byte) error {
	const expectedDataLength = 8
	if len(data) != expectedDataLength {
		return fmt.Errorf("expected len %d, actual len %d", expectedDataLength, len(data))
	}
	b.FirstIndex = common.BytesToUint32(data[:4])
	b.LastIndex = common.BytesToUint32(data[4:])

	return nil
}

type processor struct {
	db kv.RwDB
}

func newProcessor(dbPath string) (*processor, error) {
	tableCfgFunc := func(defaultBuckets kv.TableCfg) kv.TableCfg {
		cfg := kv.TableCfg{
			lastProcessedTable: {},
			gerTable:           {},
			blockTable:         {},
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
func (p *processor) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	return p.getLastProcessedBlockWithTx(tx)
}

func (p *processor) getLastIndex(ctx context.Context) (uint32, error) {
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	return p.getLastIndexWithTx(tx)
}

func (p *processor) getLastIndexWithTx(tx kv.Tx) (uint32, error) {
	iter, err := tx.RangeDescend(gerTable, common.Uint32ToBytes(math.MaxUint32), common.Uint32ToBytes(0), 1)
	if err != nil {
		return 0, err
	}
	k, _, err := iter.Next()
	if err != nil {
		return 0, err
	}
	if k == nil {
		return 0, db.ErrNotFound
	}

	return common.BytesToUint32(k), nil
}

func (p *processor) getLastProcessedBlockWithTx(tx kv.Tx) (uint64, error) {
	if lastProcessedBytes, err := tx.GetOne(lastProcessedTable, lastProcessedKey); err != nil {
		return 0, err
	} else if lastProcessedBytes == nil {
		return 0, nil
	} else {
		return common.BytesToUint64(lastProcessedBytes), nil
	}
}

func (p *processor) updateLastProcessedBlockWithTx(tx kv.RwTx, blockNum uint64) error {
	return tx.Put(lastProcessedTable, lastProcessedKey, common.Uint64ToBytes(blockNum))
}

func (p *processor) ProcessBlock(ctx context.Context, block sync.Block) error {
	tx, err := p.db.BeginRw(ctx)
	if err != nil {
		return err
	}

	lenEvents := len(block.Events)
	var lastIndex int64
	if lenEvents > 0 {
		li, err := p.getLastIndexWithTx(tx)
		switch {
		case errors.Is(err, db.ErrNotFound):
			lastIndex = -1

		case err != nil:
			tx.Rollback()
			return err

		default:
			lastIndex = int64(li)
		}
	}

	for _, e := range block.Events {
		event, ok := e.(Event)
		if !ok {
			log.Errorf("unexpected type %T in events", e)
		}
		if int64(event.L1InfoTreeIndex) < lastIndex {
			continue
		}
		lastIndex = int64(event.L1InfoTreeIndex)
		if err := tx.Put(
			gerTable,
			common.Uint32ToBytes(event.L1InfoTreeIndex),
			event.GlobalExitRoot[:],
		); err != nil {
			tx.Rollback()

			return err
		}
	}

	if lenEvents > 0 {
		firstEvent, ok := block.Events[0].(Event)
		if !ok {
			log.Errorf("unexpected type %T in events", block.Events[0])
			tx.Rollback()

			return fmt.Errorf("unexpected type %T in events", block.Events[0])
		}

		lastEvent, ok := block.Events[lenEvents-1].(Event)
		if !ok {
			log.Errorf("unexpected type %T in events", block.Events[lenEvents-1])
			tx.Rollback()

			return fmt.Errorf("unexpected type %T in events", block.Events[lenEvents-1])
		}

		bwg := blockWithGERs{
			FirstIndex: firstEvent.L1InfoTreeIndex,
			LastIndex:  lastEvent.L1InfoTreeIndex + 1,
		}

		data, err := bwg.MarshalBinary()
		if err != nil {
			tx.Rollback()

			return err
		}
		if err = tx.Put(blockTable, common.Uint64ToBytes(block.Num), data); err != nil {
			tx.Rollback()

			return err
		}
	}

	if err := p.updateLastProcessedBlockWithTx(tx, block.Num); err != nil {
		tx.Rollback()

		return err
	}

	return tx.Commit()
}

func (p *processor) Reorg(ctx context.Context, firstReorgedBlock uint64) error {
	tx, err := p.db.BeginRw(ctx)
	if err != nil {
		return err
	}

	iter, err := tx.Range(blockTable, common.Uint64ToBytes(firstReorgedBlock), nil)
	if err != nil {
		tx.Rollback()

		return err
	}
	for bNumBytes, bWithGERBytes, err := iter.Next(); bNumBytes != nil; bNumBytes, bWithGERBytes, err = iter.Next() {
		if err != nil {
			tx.Rollback()

			return err
		}
		if err := tx.Delete(blockTable, bNumBytes); err != nil {
			tx.Rollback()

			return err
		}

		bWithGER := &blockWithGERs{}
		if err := bWithGER.UnmarshalBinary(bWithGERBytes); err != nil {
			tx.Rollback()

			return err
		}
		for i := bWithGER.FirstIndex; i < bWithGER.LastIndex; i++ {
			if err := tx.Delete(gerTable, common.Uint32ToBytes(i)); err != nil {
				tx.Rollback()

				return err
			}
		}
	}

	if err := p.updateLastProcessedBlockWithTx(tx, firstReorgedBlock-1); err != nil {
		tx.Rollback()

		return err
	}

	return tx.Commit()
}

// GetFirstGERAfterL1InfoTreeIndex returns the first GER injected on the chain that is related to l1InfoTreeIndex
// or greater
func (p *processor) GetFirstGERAfterL1InfoTreeIndex(
	ctx context.Context, l1InfoTreeIndex uint32,
) (uint32, ethCommon.Hash, error) {
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return 0, ethCommon.Hash{}, err
	}
	defer tx.Rollback()

	iter, err := tx.Range(gerTable, common.Uint32ToBytes(l1InfoTreeIndex), nil)
	if err != nil {
		return 0, ethCommon.Hash{}, err
	}
	l1InfoIndexBytes, ger, err := iter.Next()
	if err != nil {
		return 0, ethCommon.Hash{}, err
	}
	if l1InfoIndexBytes == nil {
		return 0, ethCommon.Hash{}, db.ErrNotFound
	}

	return common.BytesToUint32(l1InfoIndexBytes), ethCommon.BytesToHash(ger), nil
}
