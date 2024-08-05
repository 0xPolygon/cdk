package localbridgesync

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/sync"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

const (
	eventsTable    = "localbridgesync-events"
	lastBlockTable = "localbridgesync-lastBlock"
)

var (
	ErrBlockNotProcessed = errors.New("given block(s) have not been processed yet")
	lastBlokcKey         = []byte("lb")
)

type Bridge struct {
	LeafType           uint8
	OriginNetwork      uint32
	OriginAddress      ethCommon.Address
	DestinationNetwork uint32
	DestinationAddress ethCommon.Address
	Amount             *big.Int
	Metadata           []byte
	DepositCount       uint32
}

type Claim struct {
	GlobalIndex        *big.Int
	OriginNetwork      uint32
	OriginAddress      ethCommon.Address
	DestinationAddress ethCommon.Address
	Amount             *big.Int
}

type Event struct {
	Bridge *Bridge
	Claim  *Claim
}

type processor struct {
	db kv.RwDB
}

func tableCfgFunc(defaultBuckets kv.TableCfg) kv.TableCfg {
	return kv.TableCfg{
		eventsTable:    {},
		lastBlockTable: {},
	}
}

func newProcessor(dbPath string) (*processor, error) {
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

// GetClaimsAndBridges returns the claims and bridges occurred between fromBlock, toBlock both included.
// If toBlock has not been porcessed yet, ErrBlockNotProcessed will be returned
func (p *processor) GetClaimsAndBridges(
	ctx context.Context, fromBlock, toBlock uint64,
) ([]Event, error) {
	events := []Event{}

	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	lpb, err := p.getLastProcessedBlockWithTx(tx)
	if err != nil {
		return nil, err
	}
	if lpb < toBlock {
		return nil, ErrBlockNotProcessed
	}
	c, err := tx.Cursor(eventsTable)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	for k, v, err := c.Seek(common.Uint64ToBytes(fromBlock)); k != nil; k, v, err = c.Next() {
		if err != nil {
			return nil, err
		}
		if common.BytesToUint64(k) > toBlock {
			break
		}
		blockEvents := []Event{}
		err := json.Unmarshal(v, &blockEvents)
		if err != nil {
			return nil, err
		}
		events = append(events, blockEvents...)
	}

	return events, nil
}

func (p *processor) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	return p.getLastProcessedBlockWithTx(tx)
}

func (p *processor) getLastProcessedBlockWithTx(tx kv.Tx) (uint64, error) {
	if blockNumBytes, err := tx.GetOne(lastBlockTable, lastBlokcKey); err != nil {
		return 0, err
	} else if blockNumBytes == nil {
		return 0, nil
	} else {
		return common.BytesToUint64(blockNumBytes), nil
	}
}

func (p *processor) Reorg(firstReorgedBlock uint64) error {
	tx, err := p.db.BeginRw(context.Background())
	if err != nil {
		return err
	}
	c, err := tx.Cursor(eventsTable)
	if err != nil {
		return err
	}
	defer c.Close()
	firstKey := common.Uint64ToBytes(firstReorgedBlock)
	for k, _, err := c.Seek(firstKey); k != nil; k, _, err = c.Next() {
		if err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Delete(eventsTable, k); err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := p.updateLastProcessedBlock(tx, firstReorgedBlock-1); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (p *processor) ProcessBlock(block sync.Block) error {
	tx, err := p.db.BeginRw(context.Background())
	if err != nil {
		return err
	}
	if len(block.Events) > 0 {
		events := []Event{}
		for _, e := range block.Events {
			events = append(events, e.(Event))
		}
		value, err := json.Marshal(events)
		if err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Put(eventsTable, common.Uint64ToBytes(block.Num), value); err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := p.updateLastProcessedBlock(tx, block.Num); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (p *processor) updateLastProcessedBlock(tx kv.RwTx, blockNum uint64) error {
	blockNumBytes := common.Uint64ToBytes(blockNum)
	return tx.Put(lastBlockTable, lastBlokcKey, blockNumBytes)
}
