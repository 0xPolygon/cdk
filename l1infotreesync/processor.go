package l1infotreesync

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/l1infotreesync/migrations"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
	"github.com/0xPolygon/cdk/tree"
	treeTypes "github.com/0xPolygon/cdk/tree/types"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/keccak256"
	"github.com/russross/meddler"
	"golang.org/x/crypto/sha3"
)

var (
	ErrBlockNotProcessed = errors.New("given block(s) have not been processed yet")
	ErrNotFound          = errors.New("not found")
	ErrNoBlock0          = errors.New("blockNum must be greater than 0")
)

type processor struct {
	db             *sql.DB
	l1InfoTree     *tree.AppendOnlyTree
	rollupExitTree *tree.UpdatableTree
}

// UpdateL1InfoTree representation of the UpdateL1InfoTree event
type UpdateL1InfoTree struct {
	BlockPosition   uint64
	MainnetExitRoot ethCommon.Hash
	RollupExitRoot  ethCommon.Hash
	ParentHash      ethCommon.Hash
	Timestamp       uint64
}

// VerifyBatches representation of the VerifyBatches and VerifyBatchesTrustedAggregator events
type VerifyBatches struct {
	BlockPosition uint64
	RollupID      uint32
	NumBatch      uint64
	StateRoot     ethCommon.Hash
	ExitRoot      ethCommon.Hash
	Aggregator    ethCommon.Address
}

type InitL1InfoRootMap struct {
	LeafCount         uint32
	CurrentL1InfoRoot ethCommon.Hash
}

type Event struct {
	UpdateL1InfoTree  *UpdateL1InfoTree
	VerifyBatches     *VerifyBatches
	InitL1InfoRootMap *InitL1InfoRootMap
}

// L1InfoTreeLeaf representation of a leaf of the L1 Info tree
type L1InfoTreeLeaf struct {
	BlockNumber       uint64         `meddler:"block_num"`
	BlockPosition     uint64         `meddler:"block_pos"`
	L1InfoTreeIndex   uint32         `meddler:"position"`
	PreviousBlockHash ethCommon.Hash `meddler:"previous_block_hash,hash"`
	Timestamp         uint64         `meddler:"timestamp"`
	MainnetExitRoot   ethCommon.Hash `meddler:"mainnet_exit_root,hash"`
	RollupExitRoot    ethCommon.Hash `meddler:"rollup_exit_root,hash"`
	GlobalExitRoot    ethCommon.Hash `meddler:"global_exit_root,hash"`
	Hash              ethCommon.Hash `meddler:"hash,hash"`
}

// Hash as expected by the tree
func (l *L1InfoTreeLeaf) hash() ethCommon.Hash {
	var res [32]byte
	t := make([]byte, 8) //nolint:gomnd
	binary.BigEndian.PutUint64(t, l.Timestamp)
	copy(res[:], keccak256.Hash(l.globalExitRoot().Bytes(), l.PreviousBlockHash.Bytes(), t))
	return res
}

// GlobalExitRoot returns the GER
func (l *L1InfoTreeLeaf) globalExitRoot() ethCommon.Hash {
	var gerBytes [32]byte
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(l.MainnetExitRoot[:])
	hasher.Write(l.RollupExitRoot[:])
	copy(gerBytes[:], hasher.Sum(nil))
	return gerBytes
}

func newProcessor(dbPath string) (*processor, error) {
	err := migrations.RunMigrations(dbPath)
	if err != nil {
		return nil, err
	}
	db, err := db.NewSQLiteDB(dbPath)
	if err != nil {
		return nil, err
	}
	return &processor{
		db:             db,
		l1InfoTree:     tree.NewAppendOnlyTree(db, migrations.L1InfoTreePrefix),
		rollupExitTree: tree.NewUpdatableTree(db, migrations.RollupExitTreePrefix),
	}, nil
}

// GetL1InfoTreeMerkleProof creates a merkle proof for the L1 Info tree
func (p *processor) GetL1InfoTreeMerkleProof(ctx context.Context, index uint32) (treeTypes.Proof, treeTypes.Root, error) {
	root, err := p.l1InfoTree.GetRootByIndex(ctx, index)
	if err != nil {
		return treeTypes.Proof{}, treeTypes.Root{}, err
	}
	proof, err := p.l1InfoTree.GetProof(ctx, root.Index, root.Hash)
	return proof, root, err
}

// GetLatestInfoUntilBlock returns the most recent L1InfoTreeLeaf that occurred before or at blockNum.
// If the blockNum has not been processed yet the error ErrBlockNotProcessed will be returned
func (p *processor) GetLatestInfoUntilBlock(ctx context.Context, blockNum uint64) (*L1InfoTreeLeaf, error) {
	if blockNum == 0 {
		return nil, ErrNoBlock0
	}
	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
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

	info := &L1InfoTreeLeaf{}
	return info, meddler.QueryRow(
		tx, info,
		`SELECT * FROM l1info_leaf ORDER BY block_num DESC, block_pos DESC LIMIT 1;`,
	)
}

// GetInfoByIndex returns the value of a leaf (not the hash) of the L1 info tree
func (p *processor) GetInfoByIndex(ctx context.Context, index uint32) (*L1InfoTreeLeaf, error) {
	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	return p.getInfoByIndexWithTx(tx, index)
}

func (p *processor) getInfoByIndexWithTx(tx *sql.Tx, index uint32) (*L1InfoTreeLeaf, error) {
	info := &L1InfoTreeLeaf{}
	return info, meddler.QueryRow(
		tx, info,
		`SELECT * FROM l1info_leaf WHERE position = $1;`, index,
	)
}

// GetLastProcessedBlock returns the last processed block
func (p *processor) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	return p.getLastProcessedBlockWithTx(tx)
}

func (p *processor) getLastProcessedBlockWithTx(tx *sql.Tx) (uint64, error) {
	var lastProcessedBlock uint64
	row := tx.QueryRow("SELECT num FROM BLOCK ORDER BY num DESC LIMIT 1;")
	err := row.Scan(&lastProcessedBlock)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return lastProcessedBlock, err
}

// Reorg triggers a purge and reset process on the processor to leaf it on a state
// as if the last block processed was firstReorgedBlock-1
func (p *processor) Reorg(ctx context.Context, firstReorgedBlock uint64) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if errRllbck := tx.Rollback(); errRllbck != nil {
				log.Errorf("error while rolling back tx %v", errRllbck)
			}
		}
	}()

	_, err = tx.Exec(`DELETE FROM block WHERE num >= $1;`, firstReorgedBlock)
	if err != nil {
		return err
	}

	if err = p.l1InfoTree.Reorg(tx, firstReorgedBlock); err != nil {
		return err
	}

	if err = p.rollupExitTree.Reorg(tx, firstReorgedBlock); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// ProcessBlock process the events of the block to build the rollup exit tree and the l1 info tree
// and updates the last processed block (can be called without events for that purpose)
func (p *processor) ProcessBlock(ctx context.Context, b sync.Block) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if errRllbck := tx.Rollback(); errRllbck != nil {
				log.Errorf("error while rolling back tx %v", errRllbck)
			}
		}
	}()

	if _, err := tx.Exec(`INSERT INTO block (num) VALUES ($1)`, b.Num); err != nil {
		return err
	}

	var initialL1InfoIndex uint32
	var l1InfoLeavesAdded uint32
	lastIndex, err := p.getLastIndex(tx)
	if err == ErrNotFound {
		initialL1InfoIndex = 0
	} else if err != nil {
		return err
	} else {
		initialL1InfoIndex = lastIndex + 1
	}
	for _, e := range b.Events {
		event := e.(Event)
		if event.UpdateL1InfoTree != nil {
			index := initialL1InfoIndex + l1InfoLeavesAdded
			info := &L1InfoTreeLeaf{
				BlockNumber:       b.Num,
				BlockPosition:     event.UpdateL1InfoTree.BlockPosition,
				L1InfoTreeIndex:   index,
				PreviousBlockHash: event.UpdateL1InfoTree.ParentHash,
				Timestamp:         event.UpdateL1InfoTree.Timestamp,
				MainnetExitRoot:   event.UpdateL1InfoTree.MainnetExitRoot,
				RollupExitRoot:    event.UpdateL1InfoTree.RollupExitRoot,
			}
			info.GlobalExitRoot = info.globalExitRoot()
			info.Hash = info.hash()
			err = meddler.Insert(tx, "l1info_leaf", info)
			if err != nil {
				return err
			}
			err = p.l1InfoTree.AddLeaf(tx, info.BlockNumber, info.BlockPosition, treeTypes.Leaf{
				Index: info.L1InfoTreeIndex,
				Hash:  info.Hash,
			})
			if err != nil {
				return err
			}
			l1InfoLeavesAdded++
		}

		if event.VerifyBatches != nil {
			err = p.rollupExitTree.UpsertLeaf(tx, b.Num, event.VerifyBatches.BlockPosition, treeTypes.Leaf{
				Index: event.VerifyBatches.RollupID - 1,
				Hash:  event.VerifyBatches.ExitRoot,
			})
			if err != nil {
				return err
			}
		}

		if event.InitL1InfoRootMap != nil {
			// TODO: indicate that l1 Info tree indexes before the one on this
			// event are not safe to use
			log.Debugf("TODO: handle InitL1InfoRootMap event")
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	log.Infof("block %d processed with %d events", b.Num, len(b.Events))
	return nil
}

func (p *processor) getLastIndex(tx *sql.Tx) (uint32, error) {
	var lastProcessedIndex uint32
	row := tx.QueryRow("SELECT position FROM l1info_leaf ORDER BY block_num DESC, block_pos DESC LIMIT 1;")
	err := row.Scan(&lastProcessedIndex)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return lastProcessedIndex, err
}
