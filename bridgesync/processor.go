package bridgesync

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"math/big"
	"path"

	dbCommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/sync"
	"github.com/0xPolygon/cdk/tree"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/keccak256"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

const (
	eventsTableSufix    = "-events"
	lastBlockTableSufix = "-lastBlock"
)

var (
	ErrBlockNotProcessed = errors.New("given block(s) have not been processed yet")
	ErrNotFound          = errors.New("not found")
	lastBlokcKey         = []byte("lb")
)

type Bridge struct {
	LeafType           uint8
	OriginNetwork      uint32
	OriginAddress      common.Address
	DestinationNetwork uint32
	DestinationAddress common.Address
	Amount             *big.Int
	Metadata           []byte
	DepositCount       uint32
}

func (b *Bridge) Hash() common.Hash {
	origNet := make([]byte, 4) //nolint:gomnd
	binary.BigEndian.PutUint32(origNet, uint32(b.OriginNetwork))
	destNet := make([]byte, 4) //nolint:gomnd
	binary.BigEndian.PutUint32(destNet, uint32(b.DestinationNetwork))

	metaHash := keccak256.Hash(b.Metadata)
	hash := common.Hash{}
	var buf [32]byte //nolint:gomnd
	if b.Amount == nil {
		b.Amount = big.NewInt(0)
	}
	copy(
		hash[:],
		keccak256.Hash(
			[]byte{b.LeafType},
			origNet,
			b.OriginAddress[:],
			destNet,
			b.DestinationAddress[:],
			b.Amount.FillBytes(buf[:]),
			metaHash,
		),
	)
	return hash
}

type Claim struct {
	GlobalIndex        *big.Int
	OriginNetwork      uint32
	OriginAddress      common.Address
	DestinationAddress common.Address
	Amount             *big.Int
}

type Event struct {
	Bridge *Bridge
	Claim  *Claim
}

type processor struct {
	db             kv.RwDB
	eventsTable    string
	lastBlockTable string
	exitTree       *tree.AppendOnlyTree
}

func newProcessor(ctx context.Context, dbPath, dbPrefix string) (*processor, error) {
	eventsTable := dbPrefix + eventsTableSufix
	lastBlockTable := dbPrefix + lastBlockTableSufix
	db, err := mdbx.NewMDBX(nil).
		Path(dbPath).
		WithTableCfg(func(defaultBuckets kv.TableCfg) kv.TableCfg {
			return kv.TableCfg{
				eventsTable:    {},
				lastBlockTable: {},
			}
		}).
		Open()
	if err != nil {
		return nil, err
	}
	exitTreeDBPath := path.Join(dbPath, "exittree")
	exitTree, err := tree.NewAppendOnly(ctx, exitTreeDBPath, dbPrefix)
	if err != nil {
		return nil, err
	}
	return &processor{
		db:             db,
		eventsTable:    eventsTable,
		lastBlockTable: lastBlockTable,
		exitTree:       exitTree,
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
	c, err := tx.Cursor(p.eventsTable)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	for k, v, err := c.Seek(dbCommon.Uint64ToBytes(fromBlock)); k != nil; k, v, err = c.Next() {
		if err != nil {
			return nil, err
		}
		if dbCommon.BytesToUint64(k) > toBlock {
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
	if blockNumBytes, err := tx.GetOne(p.lastBlockTable, lastBlokcKey); err != nil {
		return 0, err
	} else if blockNumBytes == nil {
		return 0, nil
	} else {
		return dbCommon.BytesToUint64(blockNumBytes), nil
	}
}

func (p *processor) Reorg(ctx context.Context, firstReorgedBlock uint64) error {
	tx, err := p.db.BeginRw(ctx)
	if err != nil {
		return err
	}
	c, err := tx.Cursor(p.eventsTable)
	if err != nil {
		return err
	}
	defer c.Close()
	firstKey := dbCommon.Uint64ToBytes(firstReorgedBlock)
	firstDepositCountReorged := int64(-1)
	for k, v, err := c.Seek(firstKey); k != nil; k, _, err = c.Next() {
		if err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Delete(p.eventsTable, k); err != nil {
			tx.Rollback()
			return err
		}
		if firstDepositCountReorged == -1 {
			events := []Event{}
			if err := json.Unmarshal(v, &events); err != nil {
				tx.Rollback()
				return err
			}
			for _, event := range events {
				if event.Bridge != nil {
					firstDepositCountReorged = int64(event.Bridge.DepositCount)
					break
				}
			}
		}
	}
	if err := p.updateLastProcessedBlock(tx, firstReorgedBlock-1); err != nil {
		tx.Rollback()
		return err
	}
	if firstDepositCountReorged != -1 {
		if err := p.exitTree.Reorg(ctx, uint32(firstDepositCountReorged)); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (p *processor) ProcessBlock(ctx context.Context, block sync.Block) error {
	tx, err := p.db.BeginRw(ctx)
	if err != nil {
		return err
	}
	bridges := []Bridge{}
	if len(block.Events) > 0 {
		events := []Event{}
		for _, e := range block.Events {
			event := e.(Event)
			events = append(events, event)
			if event.Bridge != nil {
				bridges = append(bridges, *event.Bridge)
			}
		}
		value, err := json.Marshal(events)
		if err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Put(p.eventsTable, dbCommon.Uint64ToBytes(block.Num), value); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := p.updateLastProcessedBlock(tx, block.Num); err != nil {
		tx.Rollback()
		return err
	}

	leaves := []tree.Leaf{}
	for _, bridge := range bridges {
		leaves = append(leaves, tree.Leaf{
			Index: bridge.DepositCount,
			Hash:  bridge.Hash(),
		})
	}
	if err := p.exitTree.AddLeaves(ctx, leaves); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (p *processor) updateLastProcessedBlock(tx kv.RwTx, blockNum uint64) error {
	blockNumBytes := dbCommon.Uint64ToBytes(blockNum)
	return tx.Put(p.lastBlockTable, lastBlokcKey, blockNumBytes)
}
