package l1infotreesync

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
	"github.com/0xPolygon/cdk/tree"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/keccak256"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
	"golang.org/x/crypto/sha3"
)

const (
	dbPrefix             = "l1infotreesync"
	l1InfoTreeSuffix     = "-l1infotree"
	rollupExitTreeSuffix = "-rollupexittree"

	// infoTable stores the information of L1 Info Tree (the leaves)
	// Key: index (uint32 converted to bytes)
	// Value: JSON of storeLeaf struct
	infoTable = dbPrefix + "-info"
	// blockTable stores the first and last index of L1 Info Tree that have been updated on
	// Value: JSON of blockWithLeafs
	blockTable = dbPrefix + "-block"
	// lastBlockTable used to store the last block processed. This is needed to know the last processed blcok
	lastBlockTable = dbPrefix + "-lastBlock"

	treeHeight uint8 = 32
)

var (
	ErrBlockNotProcessed = errors.New("given block(s) have not been processed yet")
	ErrNotFound          = errors.New("not found")
	ErrNoBlock0          = errors.New("blockNum must be greater than 0")
	lastBlockKey         = []byte("lb")
)

type processor struct {
	db             kv.RwDB
	l1InfoTree     *tree.AppendOnlyTree
	rollupExitTree *tree.UpdatableTree
}

type UpdateL1InfoTree struct {
	MainnetExitRoot ethCommon.Hash
	RollupExitRoot  ethCommon.Hash
	ParentHash      ethCommon.Hash
	Timestamp       uint64
}

type VerifyBatches struct {
	RollupID   uint32
	NumBatch   uint64
	StateRoot  ethCommon.Hash
	ExitRoot   ethCommon.Hash
	Aggregator ethCommon.Address
}

type Event struct {
	UpdateL1InfoTree *UpdateL1InfoTree
	VerifyBatches    *VerifyBatches
}

type L1InfoTreeLeaf struct {
	L1InfoTreeIndex   uint32
	PreviousBlockHash ethCommon.Hash
	BlockNumber       uint64
	Timestamp         uint64
	MainnetExitRoot   ethCommon.Hash
	RollupExitRoot    ethCommon.Hash
	GlobalExitRoot    ethCommon.Hash
}

type storeLeaf struct {
	BlockNumber     uint64
	MainnetExitRoot ethCommon.Hash
	RollupExitRoot  ethCommon.Hash
	ParentHash      ethCommon.Hash
	Index           uint32
	Timestamp       uint64
}

func (l *storeLeaf) Hash() ethCommon.Hash {
	var res [32]byte
	t := make([]byte, 8) //nolint:gomnd
	binary.BigEndian.PutUint64(t, l.Timestamp)
	copy(res[:], keccak256.Hash(l.GlobalExitRoot().Bytes(), l.ParentHash.Bytes(), t))
	return res
}

type blockWithLeafs struct {
	// inclusive
	FirstIndex uint32
	// not inclusive
	LastIndex uint32
}

func (l *storeLeaf) GlobalExitRoot() ethCommon.Hash {
	var gerBytes [32]byte
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(l.MainnetExitRoot[:])
	hasher.Write(l.RollupExitRoot[:])
	copy(gerBytes[:], hasher.Sum(nil))
	return gerBytes
}

func newProcessor(ctx context.Context, dbPath string) (*processor, error) {
	tableCfgFunc := func(defaultBuckets kv.TableCfg) kv.TableCfg {
		cfg := kv.TableCfg{
			infoTable:      {},
			blockTable:     {},
			lastBlockTable: {},
		}
		tree.AddTables(cfg, dbPrefix+rollupExitTreeSuffix)
		tree.AddTables(cfg, dbPrefix+l1InfoTreeSuffix)
		return cfg
	}
	db, err := mdbx.NewMDBX(nil).
		Path(dbPath).
		WithTableCfg(tableCfgFunc).
		Open()
	if err != nil {
		return nil, err
	}
	p := &processor{
		db: db,
	}

	l1InfoTree, err := tree.NewAppendOnly(ctx, db, dbPrefix+l1InfoTreeSuffix)
	if err != nil {
		return nil, err
	}
	p.l1InfoTree = l1InfoTree
	rollupExitTree := tree.NewUpdatable(ctx, db, dbPrefix+rollupExitTreeSuffix)
	p.rollupExitTree = rollupExitTree
	return p, nil
}

func (p *processor) ComputeMerkleProofByIndex(ctx context.Context, index uint32) ([]ethCommon.Hash, ethCommon.Hash, error) {
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return nil, ethCommon.Hash{}, err
	}
	defer tx.Rollback()

	root, err := p.l1InfoTree.GetRootByIndex(tx, index)
	if err != nil {
		return nil, ethCommon.Hash{}, err
	}

	proof, err := p.l1InfoTree.GetProof(ctx, index, root)
	if err != nil {
		return nil, ethCommon.Hash{}, err
	}

	// TODO: check if we need to return root or wat
	return proof, root, nil
}

// GetLatestInfoUntilBlock returns the most recent L1InfoTreeLeaf that occurred before or at blockNum.
// If the blockNum has not been processed yet the error ErrBlockNotProcessed will be returned
func (p *processor) GetLatestInfoUntilBlock(ctx context.Context, blockNum uint64) (*L1InfoTreeLeaf, error) {
	if blockNum == 0 {
		return nil, ErrNoBlock0
	}
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	lpb, err := p.getLastProcessedBlockWithTx(tx)
	if err != nil {
		return nil, err
	}
	if lpb < blockNum {
		return nil, ErrBlockNotProcessed
	}
	iter, err := tx.RangeDescend(blockTable, common.Uint64ToBytes(blockNum), common.Uint64ToBytes(0), 1)
	if err != nil {
		return nil, fmt.Errorf(
			"error calling RangeDescend(blockTable, %d, 0, 1): %w", blockNum, err,
		)
	}
	k, v, err := iter.Next()
	if err != nil {
		return nil, err
	}
	if k == nil {
		return nil, ErrNotFound
	}
	blk := blockWithLeafs{}
	if err := json.Unmarshal(v, &blk); err != nil {
		return nil, err
	}
	return p.getInfoByIndexWithTx(tx, blk.LastIndex-1)
}

func (p *processor) GetInfoByIndex(ctx context.Context, index uint32) (*L1InfoTreeLeaf, error) {
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	return p.getInfoByIndexWithTx(tx, index)
}

func (p *processor) getInfoByIndexWithTx(tx kv.Tx, index uint32) (*L1InfoTreeLeaf, error) {
	infoBytes, err := tx.GetOne(infoTable, common.Uint32ToBytes(index))
	if err != nil {
		return nil, err
	}
	if infoBytes == nil {
		return nil, ErrNotFound
	}

	var info storeLeaf
	if err := json.Unmarshal(infoBytes, &info); err != nil {
		return nil, err
	}
	return &L1InfoTreeLeaf{
		L1InfoTreeIndex:   info.Index,
		PreviousBlockHash: info.ParentHash,
		BlockNumber:       info.BlockNumber,
		Timestamp:         info.Timestamp,
		MainnetExitRoot:   info.MainnetExitRoot,
		RollupExitRoot:    info.RollupExitRoot,
		GlobalExitRoot:    info.GlobalExitRoot(),
	}, nil
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
	blockNumBytes, err := tx.GetOne(lastBlockTable, lastBlockKey)
	if err != nil {
		return 0, err
	} else if blockNumBytes == nil {
		return 0, nil
	}
	return common.BytesToUint64(blockNumBytes), nil
}

func (p *processor) Reorg(ctx context.Context, firstReorgedBlock uint64) error {
	tx, err := p.db.BeginRw(ctx)
	if err != nil {
		return err
	}
	c, err := tx.Cursor(blockTable)
	if err != nil {
		return err
	}
	defer c.Close()
	firstKey := common.Uint64ToBytes(firstReorgedBlock)
	firstReorgedL1InfoTreeIndex := int64(-1)
	for blkKey, blkValue, err := c.Seek(firstKey); blkKey != nil; blkKey, blkValue, err = c.Next() {
		if err != nil {
			tx.Rollback()
			return err
		}
		var blk blockWithLeafs
		if err := json.Unmarshal(blkValue, &blk); err != nil {
			tx.Rollback()
			return err
		}
		for i := blk.FirstIndex; i < blk.LastIndex; i++ {
			if firstReorgedL1InfoTreeIndex == -1 {
				firstReorgedL1InfoTreeIndex = int64(i)
			}
			if err := p.deleteLeaf(tx, i); err != nil {
				tx.Rollback()
				return err
			}
		}
		if err := tx.Delete(blockTable, blkKey); err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := p.updateLastProcessedBlock(tx, firstReorgedBlock-1); err != nil {
		tx.Rollback()
		return err
	}
	var rollbackL1InfoTree func()
	if firstReorgedL1InfoTreeIndex != -1 {
		rollbackL1InfoTree, err = p.l1InfoTree.Reorg(tx, uint32(firstReorgedL1InfoTreeIndex))
		if err != nil {
			tx.Rollback()
			rollbackL1InfoTree()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		rollbackL1InfoTree()
		return err
	}
	return nil
}

func (p *processor) deleteLeaf(tx kv.RwTx, index uint32) error {
	if err := tx.Delete(infoTable, common.Uint32ToBytes(index)); err != nil {
		return err
	}
	return nil
}

// ProcessBlock process the leafs of the L1 info tree found on a block
// this function can be called without leafs with the intention to track the last processed block
func (p *processor) ProcessBlock(ctx context.Context, b sync.Block) error {
	tx, err := p.db.BeginRw(ctx)
	if err != nil {
		return err
	}
	events := make([]Event, len(b.Events))
	rollupExitTreeRollback := func() {}
	l1InfoTreeRollback := func() {}
	rollback := func() {
		tx.Rollback()
		rollupExitTreeRollback()
		l1InfoTreeRollback()
	}
	l1InfoTreeLeavesToAdd := []tree.Leaf{}
	if len(b.Events) > 0 {
		var initialIndex uint32
		lastIndex, err := p.getLastIndex(tx)
		if err == ErrNotFound {
			initialIndex = 0
		} else if err != nil {
			rollback()
			return err
		} else {
			initialIndex = lastIndex + 1
		}
		var nextExpectedRollupExitTreeRoot *ethCommon.Hash
		for i, e := range b.Events {
			event := e.(Event)
			events = append(events, event)
			if event.UpdateL1InfoTree != nil {
				leafToStore := storeLeaf{
					BlockNumber:     b.Num,
					Index:           initialIndex + uint32(i),
					MainnetExitRoot: event.UpdateL1InfoTree.MainnetExitRoot,
					RollupExitRoot:  event.UpdateL1InfoTree.RollupExitRoot,
					ParentHash:      event.UpdateL1InfoTree.ParentHash,
					Timestamp:       event.UpdateL1InfoTree.Timestamp,
				}
				if err := p.storeLeafInfo(tx, leafToStore); err != nil {
					tx.Rollback()
					return err
				}
				l1InfoTreeLeavesToAdd = append(l1InfoTreeLeavesToAdd, tree.Leaf{
					Index: initialIndex + uint32(i),
					Hash:  leafToStore.Hash(),
				})
				nextExpectedRollupExitTreeRoot = &leafToStore.RollupExitRoot
			}

			if event.VerifyBatches != nil {
				// before the verify batches event happens, the updateExitRoot event is emitted.
				// Since the previous event include the rollup exit root, this can use it to assert
				// that the computation of the tree is correct. However, there are some execution paths
				// on the contract that don't follow this (verifyBatches + pendingStateTimeout != 0)
				rollupExitTreeRollback, err = p.rollupExitTree.UpsertLeaf(
					tx,
					event.VerifyBatches.RollupID,
					event.VerifyBatches.ExitRoot,
					nextExpectedRollupExitTreeRoot,
				)
				if err != nil {
					rollback()
					return err
				}
				nextExpectedRollupExitTreeRoot = nil
			}
		}
		bwl := blockWithLeafs{
			FirstIndex: initialIndex,
			LastIndex:  initialIndex + uint32(len(b.Events)),
		}
		blockValue, err := json.Marshal(bwl)
		if err != nil {
			rollback()
			return err
		}
		if err := tx.Put(blockTable, common.Uint64ToBytes(b.Num), blockValue); err != nil {
			rollback()
			return err
		}
	}
	if err := p.updateLastProcessedBlock(tx, b.Num); err != nil {
		rollback()
		return err
	}

	l1InfoTreeRollback, err = p.l1InfoTree.AddLeaves(tx, l1InfoTreeLeavesToAdd)
	if err != nil {
		rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		rollback()
		return err
	}
	log.Debugf("block %d processed with events: %+v", b.Num, events)
	return nil
}

func (p *processor) getLastIndex(tx kv.Tx) (uint32, error) {
	bNum, err := p.getLastProcessedBlockWithTx(tx)
	if err != nil {
		return 0, err
	}
	if bNum == 0 {
		return 0, nil
	}
	iter, err := tx.RangeDescend(blockTable, common.Uint64ToBytes(bNum), common.Uint64ToBytes(0), 1)
	if err != nil {
		return 0, err
	}
	_, blkBytes, err := iter.Next()
	if err != nil {
		return 0, err
	}
	if blkBytes == nil {
		return 0, ErrNotFound
	}
	var blk blockWithLeafs
	if err := json.Unmarshal(blkBytes, &blk); err != nil {
		return 0, err
	}
	return blk.LastIndex - 1, nil
}

func (p *processor) storeLeafInfo(tx kv.RwTx, leaf storeLeaf) error {
	leafValue, err := json.Marshal(leaf)
	if err != nil {
		return err
	}
	return tx.Put(infoTable, common.Uint32ToBytes(leaf.Index), leafValue)
}

func (p *processor) updateLastProcessedBlock(tx kv.RwTx, blockNum uint64) error {
	blockNumBytes := common.Uint64ToBytes(blockNum)
	return tx.Put(lastBlockTable, lastBlockKey, blockNumBytes)
}
