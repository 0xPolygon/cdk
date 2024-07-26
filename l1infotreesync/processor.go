package l1infotreesync

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/l1infotree"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
	"golang.org/x/crypto/sha3"
)

const (
	rootTable      = "l1infotreesync-root"
	indexTable     = "l1infotreesync-index"
	infoTable      = "l1infotreesync-info"
	blockTable     = "l1infotreesync-block"
	lastBlockTable = "l1infotreesync-lastBlock"

	treeHeight uint8 = 32
)

var (
	ErrBlockNotProcessed = errors.New("given block(s) have not been processed yet")
	ErrNotFound          = errors.New("not found")
	lastBlokcKey         = []byte("lb")
)

type processor struct {
	db   kv.RwDB
	tree *l1infotree.L1InfoTree
}

type Event struct {
	MainnetExitRoot ethCommon.Hash
	RollupExitRoot  ethCommon.Hash
	ParentHash      ethCommon.Hash
	Timestamp       uint64
}

type L1InfoTreeLeaf struct {
	L1InfoTreeRoot    ethCommon.Hash
	L1InfoTreeIndex   uint32
	PreviousBlockHash ethCommon.Hash
	BlockNumber       uint64
	Timestamp         uint64
	MainnetExitRoot   ethCommon.Hash
	RollupExitRoot    ethCommon.Hash
	GlobalExitRoot    ethCommon.Hash
}

type storeLeaf struct {
	MainnetExitRoot ethCommon.Hash
	RollupExitRoot  ethCommon.Hash
	ParentHash      ethCommon.Hash
	InfoRoot        ethCommon.Hash
	Index           uint32
	Timestamp       uint64
	BlockNumber     uint64
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

func tableCfgFunc(defaultBuckets kv.TableCfg) kv.TableCfg {
	return kv.TableCfg{
		rootTable:      {},
		indexTable:     {},
		infoTable:      {},
		blockTable:     {},
		lastBlockTable: {},
	}
}

func newProcessor(ctx context.Context, dbPath string) (*processor, error) {
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
	leaves, err := p.getAllLeavesHashed(ctx)
	if err != nil {
		return nil, err
	}
	tree, err := l1infotree.NewL1InfoTree(treeHeight, leaves)
	if err != nil {
		return nil, err
	}
	p.tree = tree
	return p, nil
}

func (p *processor) getAllLeavesHashed(ctx context.Context) ([][32]byte, error) {
	// TODO: same coment about refactor that appears at ComputeMerkleProofByIndex
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	index, err := p.getLastIndex(tx)
	if err == ErrNotFound || index == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return p.getHasedLeaves(tx, index)
}

func (p *processor) ComputeMerkleProofByIndex(ctx context.Context, index uint32) ([][32]byte, ethCommon.Hash, error) {
	// TODO: refactor the tree to store the nodes so it's not neede to load all the leaves and compute the tree
	// every time this function is called. Since it's not a sparse MT, an alternative could be to store the proofs
	// as part of the info
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return nil, ethCommon.Hash{}, err
	}
	defer tx.Rollback()

	leaves, err := p.getHasedLeaves(tx, index)
	if err != nil {
		return nil, ethCommon.Hash{}, err
	}
	return p.tree.ComputeMerkleProof(index, leaves)
}

func (p *processor) getHasedLeaves(tx kv.Tx, untilIndex uint32) ([][32]byte, error) {
	leaves := [][32]byte{}
	for i := uint32(0); i <= untilIndex; i++ {
		info, err := p.getInfoByIndexWithTx(tx, i)
		if err != nil {
			return nil, err
		}
		h := l1infotree.HashLeafData(info.GlobalExitRoot, info.PreviousBlockHash, info.Timestamp)
		leaves = append(leaves, h)
	}
	return leaves, nil
}

func (p *processor) ComputeMerkleProofByRoot(ctx context.Context, root ethCommon.Hash) ([][32]byte, ethCommon.Hash, error) {
	info, err := p.GetInfoByRoot(ctx, root)
	if err != nil {
		return nil, ethCommon.Hash{}, err
	}
	return p.ComputeMerkleProofByIndex(ctx, info.L1InfoTreeIndex)
}

func (p *processor) GetInfoByRoot(ctx context.Context, root ethCommon.Hash) (*L1InfoTreeLeaf, error) {
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	hash, err := tx.GetOne(rootTable, root[:])
	if err != nil {
		return nil, err
	}
	if hash == nil {
		return nil, ErrNotFound
	}
	return p.getInfoByHashWithTx(tx, hash)
}

// GetLatestInfoUntilBlock returns the most recent L1InfoTreeLeaf that occured before or at blockNum.
// If the blockNum has not been processed yet the error ErrBlockNotProcessed will be returned
func (p *processor) GetLatestInfoUntilBlock(ctx context.Context, blockNum uint64) (*L1InfoTreeLeaf, error) {
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	lpb, err := p.getLastProcessedBlockWithTx(tx)
	if lpb < blockNum {
		return nil, ErrBlockNotProcessed
	}
	iter, err := tx.RangeDescend(blockTable, uint64ToBytes(blockNum), uint64ToBytes(0), 1)
	if err != nil {
		return nil, err
	}
	if !iter.HasNext() {
		return nil, ErrNotFound
	}
	_, v, err := iter.Next()
	if err != nil {
		return nil, err
	}
	blk := blockWithLeafs{}
	if err := json.Unmarshal(v, &blk); err != nil {
		return nil, err
	}
	hash, err := tx.GetOne(indexTable, common.Uint32ToBytes(blk.LastIndex-1))
	if err != nil {
		return nil, err
	}
	return p.getInfoByHashWithTx(tx, hash)
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
	hash, err := tx.GetOne(indexTable, common.Uint32ToBytes(index))
	if err != nil {
		return nil, err
	}
	if hash == nil {
		return nil, ErrNotFound
	}
	return p.getInfoByHashWithTx(tx, hash)
}

func (p *processor) GetInfoByHash(ctx context.Context, hash []byte) (*L1InfoTreeLeaf, error) {
	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	return p.getInfoByHashWithTx(tx, hash)
}

func (p *processor) getInfoByHashWithTx(tx kv.Tx, hash []byte) (*L1InfoTreeLeaf, error) {
	infoBytes, err := tx.GetOne(infoTable, hash)
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
		L1InfoTreeRoot:    info.InfoRoot,
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
	if blockNumBytes, err := tx.GetOne(lastBlockTable, lastBlokcKey); err != nil {
		return 0, err
	} else if blockNumBytes == nil {
		return 0, nil
	} else {
		return bytes2Uint64(blockNumBytes), nil
	}
}

func (p *processor) Reorg(firstReorgedBlock uint64) error {
	// TODO: Does tree need to be reorged?
	tx, err := p.db.BeginRw(context.Background())
	if err != nil {
		return err
	}
	c, err := tx.Cursor(blockTable)
	if err != nil {
		return err
	}
	defer c.Close()
	firstKey := uint64ToBytes(firstReorgedBlock)
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
	return tx.Commit()
}

func (p *processor) deleteLeaf(tx kv.RwTx, index uint32) error {
	// TODO: do we need to do something with p.tree here?
	// Get leaf info to delete all relations
	hash, err := tx.GetOne(indexTable, common.Uint32ToBytes(index))
	if err != nil {
		return err
	}
	if hash == nil {
		return ErrNotFound
	}
	infoBytes, err := tx.GetOne(infoTable, hash)
	if err != nil {
		return err
	}
	if infoBytes == nil {
		return ErrNotFound
	}
	var info storeLeaf
	if err := json.Unmarshal(infoBytes, &info); err != nil {
		return err
	}

	// Delete
	if err := tx.Delete(rootTable, info.InfoRoot[:]); err != nil {
		return err
	}
	if err := tx.Delete(indexTable, common.Uint32ToBytes(index)); err != nil {
		return err
	}
	if err := tx.Delete(infoTable, hash); err != nil {
		return err
	}
	return nil
}

// ProcessBlock process the leafs of the L1 info tree found on a block
// this function can be called without leafs with the intention to track the last processed block
func (p *processor) ProcessBlock(b sync.Block) error {
	tx, err := p.db.BeginRw(context.Background())
	if err != nil {
		return err
	}
	if len(b.Events) > 0 {
		var initialIndex uint32
		lastIndex, err := p.getLastIndex(tx)
		if err == ErrNotFound {
			initialIndex = 0
		} else if err != nil {
			tx.Rollback()
			return err
		} else {
			initialIndex = lastIndex + 1
		}
		for i, e := range b.Events {
			event := e.(Event)
			leafToStore := storeLeaf{
				Index:           initialIndex + uint32(i),
				MainnetExitRoot: event.MainnetExitRoot,
				RollupExitRoot:  event.RollupExitRoot,
				ParentHash:      event.ParentHash,
				Timestamp:       event.Timestamp,
				BlockNumber:     b.Num,
			}
			if err := p.addLeaf(tx, leafToStore); err != nil {
				tx.Rollback()
				return err
			}
		}
		bwl := blockWithLeafs{
			FirstIndex: initialIndex,
			LastIndex:  initialIndex + uint32(len(b.Events)),
		}
		blockValue, err := json.Marshal(bwl)
		if err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Put(blockTable, uint64ToBytes(b.Num), blockValue); err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := p.updateLastProcessedBlock(tx, b.Num); err != nil {
		tx.Rollback()
		return err
	}
	log.Debugf("block %d processed with events: %+v", b.Num, b.Events)
	return tx.Commit()
}

func (p *processor) getLastIndex(tx kv.Tx) (uint32, error) {
	bNum, err := p.getLastProcessedBlockWithTx(tx)
	if err != nil {
		return 0, err
	}
	if bNum == 0 {
		return 0, nil
	}
	iter, err := tx.RangeDescend(blockTable, uint64ToBytes(bNum), uint64ToBytes(0), 1)
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

func (p *processor) addLeaf(tx kv.RwTx, leaf storeLeaf) error {
	// Update tree
	hash := l1infotree.HashLeafData(leaf.GlobalExitRoot(), leaf.ParentHash, leaf.Timestamp)
	root, err := p.tree.AddLeaf(leaf.Index, hash)
	if err != nil {
		return err
	}
	leaf.InfoRoot = root
	// store info
	leafValue, err := json.Marshal(leaf)
	if err != nil {
		return err
	}
	if err := tx.Put(infoTable, hash[:], leafValue); err != nil {
		return err
	}
	// store index relation
	if err := tx.Put(indexTable, common.Uint32ToBytes(leaf.Index), hash[:]); err != nil {
		return err
	}
	// store root relation
	if err := tx.Put(rootTable, root.Bytes(), hash[:]); err != nil {
		return err
	}
	return nil
}

func (p *processor) updateLastProcessedBlock(tx kv.RwTx, blockNum uint64) error {
	blockNumBytes := uint64ToBytes(blockNum)
	return tx.Put(lastBlockTable, lastBlokcKey, blockNumBytes)
}

func uint64ToBytes(num uint64) []byte {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, num)
	return key
}

func bytes2Uint64(key []byte) uint64 {
	return binary.LittleEndian.Uint64(key)
}
