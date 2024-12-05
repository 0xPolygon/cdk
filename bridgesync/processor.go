package bridgesync

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk/bridgesync/migrations"
	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
	"github.com/0xPolygon/cdk/tree"
	"github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/keccak256"
	"github.com/russross/meddler"
	_ "modernc.org/sqlite"
)

const (
	globalIndexPartSize = 4
	globalIndexMaxSize  = 9
)

var (
	// errBlockNotProcessedFormat indicates that the given block(s) have not been processed yet.
	errBlockNotProcessedFormat = fmt.Sprintf("block %%d not processed, last processed: %%d")
)

// Bridge is the representation of a bridge event
type Bridge struct {
	BlockNum           uint64         `meddler:"block_num"`
	BlockPos           uint64         `meddler:"block_pos"`
	LeafType           uint8          `meddler:"leaf_type"`
	OriginNetwork      uint32         `meddler:"origin_network"`
	OriginAddress      common.Address `meddler:"origin_address"`
	DestinationNetwork uint32         `meddler:"destination_network"`
	DestinationAddress common.Address `meddler:"destination_address"`
	Amount             *big.Int       `meddler:"amount,bigint"`
	Metadata           []byte         `meddler:"metadata"`
	DepositCount       uint32         `meddler:"deposit_count"`
}

// Hash returns the hash of the bridge event as expected by the exit tree
func (b *Bridge) Hash() common.Hash {
	const (
		uint32ByteSize = 4
		bigIntSize     = 32
	)
	origNet := make([]byte, uint32ByteSize)
	binary.BigEndian.PutUint32(origNet, b.OriginNetwork)
	destNet := make([]byte, uint32ByteSize)
	binary.BigEndian.PutUint32(destNet, b.DestinationNetwork)

	metaHash := keccak256.Hash(b.Metadata)
	var buf [bigIntSize]byte
	if b.Amount == nil {
		b.Amount = big.NewInt(0)
	}

	return common.BytesToHash(keccak256.Hash(
		[]byte{b.LeafType},
		origNet,
		b.OriginAddress[:],
		destNet,
		b.DestinationAddress[:],
		b.Amount.FillBytes(buf[:]),
		metaHash,
	))
}

// Claim representation of a claim event
type Claim struct {
	BlockNum            uint64         `meddler:"block_num"`
	BlockPos            uint64         `meddler:"block_pos"`
	GlobalIndex         *big.Int       `meddler:"global_index,bigint"`
	OriginNetwork       uint32         `meddler:"origin_network"`
	OriginAddress       common.Address `meddler:"origin_address"`
	DestinationAddress  common.Address `meddler:"destination_address"`
	Amount              *big.Int       `meddler:"amount,bigint"`
	ProofLocalExitRoot  types.Proof    `meddler:"proof_local_exit_root,merkleproof"`
	ProofRollupExitRoot types.Proof    `meddler:"proof_rollup_exit_root,merkleproof"`
	MainnetExitRoot     common.Hash    `meddler:"mainnet_exit_root,hash"`
	RollupExitRoot      common.Hash    `meddler:"rollup_exit_root,hash"`
	GlobalExitRoot      common.Hash    `meddler:"global_exit_root,hash"`
	DestinationNetwork  uint32         `meddler:"destination_network"`
	Metadata            []byte         `meddler:"metadata"`
	IsMessage           bool           `meddler:"is_message"`
}

// Event combination of bridge and claim events
type Event struct {
	Pos    uint64
	Bridge *Bridge
	Claim  *Claim
}

type BridgeContractor interface {
	LastUpdatedDepositCount(ctx context.Context, BlockNumber uint64) (uint32, error)
}

type processor struct {
	db             *sql.DB
	exitTree       *tree.AppendOnlyTree
	log            *log.Logger
	bridgeContract BridgeContractor
}

func newProcessor(dbPath, loggerPrefix string) (*processor, error) {
	err := migrations.RunMigrations(dbPath)
	if err != nil {
		return nil, err
	}
	db, err := db.NewSQLiteDB(dbPath)
	if err != nil {
		return nil, err
	}
	logger := log.WithFields("bridge-syncer", loggerPrefix)
	exitTree := tree.NewAppendOnlyTree(db, "")
	return &processor{
		db:       db,
		exitTree: exitTree,
		log:      logger,
	}, nil
}
func (p *processor) GetBridgesPublished(
	ctx context.Context, fromBlock, toBlock uint64,
) ([]Bridge, error) {
	return p.GetBridges(ctx, fromBlock, toBlock)
}

func (p *processor) GetBridges(
	ctx context.Context, fromBlock, toBlock uint64,
) ([]Bridge, error) {
	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Warnf("error rolling back tx: %v", err)
		}
	}()
	rows, err := p.queryBlockRange(tx, fromBlock, toBlock, "bridge")
	if err != nil {
		return nil, err
	}
	bridgePtrs := []*Bridge{}
	if err = meddler.ScanAll(rows, &bridgePtrs); err != nil {
		return nil, err
	}
	bridgesIface := db.SlicePtrsToSlice(bridgePtrs)
	bridges, ok := bridgesIface.([]Bridge)
	if !ok {
		return nil, errors.New("failed to convert from []*Bridge to []Bridge")
	}
	return bridges, nil
}

func (p *processor) GetClaims(
	ctx context.Context, fromBlock, toBlock uint64,
) ([]Claim, error) {
	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Warnf("error rolling back tx: %v", err)
		}
	}()
	rows, err := p.queryBlockRange(tx, fromBlock, toBlock, "claim")
	if err != nil {
		return nil, err
	}
	claimPtrs := []*Claim{}
	if err = meddler.ScanAll(rows, &claimPtrs); err != nil {
		return nil, err
	}
	claimsIface := db.SlicePtrsToSlice(claimPtrs)
	claims, ok := claimsIface.([]Claim)
	if !ok {
		return nil, errors.New("failed to convert from []*Claim to []Claim")
	}
	return claims, nil
}

func (p *processor) queryBlockRange(tx db.Querier, fromBlock, toBlock uint64, table string) (*sql.Rows, error) {
	if err := p.isBlockProcessed(tx, toBlock); err != nil {
		return nil, err
	}
	rows, err := tx.Query(fmt.Sprintf(`
		SELECT * FROM %s
		WHERE block_num >= $1 AND block_num <= $2;
	`, table), fromBlock, toBlock)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, db.ErrNotFound
		}
		return nil, err
	}
	return rows, nil
}

func (p *processor) isBlockProcessed(tx db.Querier, blockNum uint64) error {
	lpb, err := p.getLastProcessedBlockWithTx(tx)
	if err != nil {
		return err
	}
	if lpb < blockNum {
		return fmt.Errorf(errBlockNotProcessedFormat, blockNum, lpb)
	}
	return nil
}

// GetLastProcessedBlock returns the last processed block by the processor, including blocks
// that don't have events
func (p *processor) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	return p.getLastProcessedBlockWithTx(p.db)
}

func (p *processor) getLastProcessedBlockWithTx(tx db.Querier) (uint64, error) {
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
	tx, err := db.NewTx(ctx, p.db)
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

	if err = p.exitTree.Reorg(tx, firstReorgedBlock); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// ProcessBlock process the events of the block to build the exit tree
// and updates the last processed block (can be called without events for that purpose)
func (p *processor) ProcessBlock(ctx context.Context, block sync.Block) error {
	tx, err := db.NewTx(ctx, p.db)
	if err != nil {
		return err
	}
	shouldRollback := true
	defer func() {
		if shouldRollback {
			if errRllbck := tx.Rollback(); errRllbck != nil {
				log.Errorf("error while rolling back tx %v", errRllbck)
			}
		}
	}()

	if _, err := tx.Exec(`INSERT INTO block (num) VALUES ($1)`, block.Num); err != nil {
		return err
	}
	for _, e := range block.Events {
		event, ok := e.(Event)
		if !ok {
			return errors.New("failed to convert sync.Block.Event to Event")
		}
		if event.Bridge != nil {
			if err = p.exitTree.AddLeaf(tx, block.Num, event.Pos, types.Leaf{
				Index: event.Bridge.DepositCount,
				Hash:  event.Bridge.Hash(),
			}); err != nil {
				return err
			}
			if err = meddler.Insert(tx, "bridge", event.Bridge); err != nil {
				return err
			}
		}
		if event.Claim != nil {
			if err = meddler.Insert(tx, "claim", event.Claim); err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	shouldRollback = false

	p.log.Debugf("processed %d events until block %d", len(block.Events), block.Num)
	return nil
}

func GenerateGlobalIndex(mainnetFlag bool, rollupIndex uint32, localExitRootIndex uint32) *big.Int {
	var (
		globalIndexBytes []byte
		buf              [globalIndexPartSize]byte
	)
	if mainnetFlag {
		globalIndexBytes = append(globalIndexBytes, big.NewInt(1).Bytes()...)
		ri := big.NewInt(0).FillBytes(buf[:])
		globalIndexBytes = append(globalIndexBytes, ri...)
	} else {
		ri := big.NewInt(0).SetUint64(uint64(rollupIndex)).FillBytes(buf[:])
		globalIndexBytes = append(globalIndexBytes, ri...)
	}
	leri := big.NewInt(0).SetUint64(uint64(localExitRootIndex)).FillBytes(buf[:])
	globalIndexBytes = append(globalIndexBytes, leri...)

	result := big.NewInt(0).SetBytes(globalIndexBytes)

	return result
}

// Decodes global index to its three parts:
// 1. mainnetFlag - first byte
// 2. rollupIndex - next 4 bytes
// 3. localExitRootIndex - last 4 bytes
// NOTE - mainnet flag is not in the global index bytes if it is false
// NOTE - rollup index is 0 if mainnet flag is true
// NOTE - rollup index is not in the global index bytes if mainnet flag is false and rollup index is 0
func DecodeGlobalIndex(globalIndex *big.Int) (mainnetFlag bool,
	rollupIndex uint32, localExitRootIndex uint32, err error) {
	globalIndexBytes := globalIndex.Bytes()
	l := len(globalIndexBytes)
	if l > globalIndexMaxSize {
		return false, 0, 0, errors.New("invalid global index length")
	}

	if l == 0 {
		// false, 0, 0
		return
	}

	if l == globalIndexMaxSize {
		// true, rollupIndex, localExitRootIndex
		mainnetFlag = true
	}

	localExitRootFromIdx := l - globalIndexPartSize
	if localExitRootFromIdx < 0 {
		localExitRootFromIdx = 0
	}

	rollupIndexFromIdx := localExitRootFromIdx - globalIndexPartSize
	if rollupIndexFromIdx < 0 {
		rollupIndexFromIdx = 0
	}

	rollupIndex = convertBytesToUint32(globalIndexBytes[rollupIndexFromIdx:localExitRootFromIdx])
	localExitRootIndex = convertBytesToUint32(globalIndexBytes[localExitRootFromIdx:])

	return
}

func convertBytesToUint32(bytes []byte) uint32 {
	return uint32(big.NewInt(0).SetBytes(bytes).Uint64())
}
