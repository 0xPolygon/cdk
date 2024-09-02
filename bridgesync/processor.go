package bridgesync

import (
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"math/big"

	dbCommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
	"github.com/0xPolygon/cdk/tree"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/keccak256"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
	_ "modernc.org/sqlite"
)

const (
	eventsTableSufix    = "-events"
	lastBlockTableSufix = "-lastBlock"
)

var (
	ErrBlockNotProcessed = errors.New("given block(s) have not been processed yet")
	ErrNotFound          = errors.New("not found")
	lastBlockKey         = []byte("lb")
)

// Bridge is the representation of a bridge event
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
	BlockNum uint64
	BlockPos uint64
	// From claim event
	GlobalIndex        *big.Int
	OriginNetwork      uint32
	OriginAddress      common.Address
	DestinationAddress common.Address
	Amount             *big.Int
	// From call data
	ProofLocalExitRoot  [tree.DefaultHeight]common.Hash
	ProofRollupExitRoot [tree.DefaultHeight]common.Hash
	MainnetExitRoot     common.Hash
	RollupExitRoot      common.Hash
	GlobalExitRoot      common.Hash
	DestinationNetwork  uint32
	Metadata            []byte
	// Meta
	IsMessage bool
}

// Event combination of bridge and claim events
type Event struct {
	Bridge *Bridge
	Claim  *Claim
}

type processor struct {
	db             *sql.DB
	eventsTable    string
	lastBlockTable string
	exitTree       *tree.AppendOnlyTree
	log            *log.Logger
}

func newProcessor(ctx context.Context, dbPath, dbPrefix string) (*processor, error) {
	db, err := db.NewSQLiteDB(dbPath)
	if err != nil {
		return nil, err
	}
	eventsTable := dbPrefix + eventsTableSufix
	lastBlockTable := dbPrefix + lastBlockTableSufix
	logger := log.WithFields("bridge-syncer", dbPrefix)
	tableCfgFunc := func(defaultBuckets kv.TableCfg) kv.TableCfg {
		cfg := kv.TableCfg{}
		tree.AddTables(cfg, dbPrefix)
		return cfg
	}
	treeDB, err := mdbx.NewMDBX(nil).
		Path(dbPath).
		WithTableCfg(tableCfgFunc).
		Open()
	exitTree, err := tree.NewAppendOnlyTree(ctx, treeDB, dbPrefix)
	if err != nil {
		return nil, err
	}
	return &processor{
		db:             db,
		eventsTable:    eventsTable,
		lastBlockTable: lastBlockTable,
		exitTree:       exitTree,
		log:            logger,
	}, nil
}

// GetClaimsAndBridges returns the claims and bridges occurred between fromBlock, toBlock both included.
// If toBlock has not been porcessed yet, ErrBlockNotProcessed will be returned
func (p *processor) GetBridges(
	ctx context.Context, fromBlock, toBlock uint64,
) ([]Bridge, error) {
	return nil, nil
}

func (p *processor) GetClaims(
	ctx context.Context, fromBlock, toBlock uint64,
) ([]Claim, error) {
	tx, err := p.db.BeginTx(ctx, nil)
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

	rows, err := tx.Query(`
		SELECT
			block_num,
			block_pos,
			global_index,
			origin_network,
			origin_address,
			destination_address,
			amount,
			proof_local_exit_root,
			proof_rollup_exit_root,
			mainnet_exit_root,
			rollup_exit_root,
			global_exit_root,
			destination_network,
			is_message
			metadata,
		FROM bridge
		WHERE block_num >= $1 AND block_num <= $2;
	 `)
	if err != nil {
		return nil, err
	}
	claims := []Claim{}
	for rows.Next() {
		b, err := scanClaim(rows)
		if err != nil {
			return nil, err
		}
		claims = append(claims, b)
	}

	return claims, nil
}

func scanClaim(rows *sql.Rows) (Claim, error) {
	var (
		block_num              uint64
		block_pos              uint64
		global_index           *big.Int
		origin_network         uint32
		origin_address         common.Address
		destination_address    common.Address
		amount                 *big.Int
		proof_local_exit_root  common.Hash
		proof_rollup_exit_root common.Hash
		mainnet_exit_root      common.Hash
		rollup_exit_root       common.Hash
		global_exit_root       common.Hash
		destination_network    uint32
		is_message             bool
		metadata               []byte
	)
	if err := rows.Scan(
		&block_num,
		&block_pos,
		&global_index,
		&origin_network,
		&origin_address,
		&destination_address,
		&amount,
		&proof_local_exit_root,
		&proof_rollup_exit_root,
		&mainnet_exit_root,
		&rollup_exit_root,
		&global_exit_root,
		&destination_network,
		&metadata,
		&is_message,
	); err != nil {
		return Claim{}, err
	}
	return Claim{
		BlockNum:            block_num,
		BlockPos:            block_pos,
		GlobalIndex:         global_index,
		OriginNetwork:       origin_network,
		OriginAddress:       origin_address,
		DestinationAddress:  destination_address,
		Amount:              amount,
		ProofLocalExitRoot:  proof_local_exit_root,
		ProofRollupExitRoot: proof_rollup_exit_root,
		MainnetExitRoot:     mainnet_exit_root,
		GlobalExitRoot:      rollup_exit_root,
		RollupExitRoot:      global_exit_root,
		DestinationNetwork:  destination_network,
		Metadata:            metadata,
		IsMessage:           is_message,
	}, nil
}

// GetLastProcessedBlock returns the last processed block by the processor, including blocks
// that don't have events
func (p *processor) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	tx, err := p.db.BeginTx(ctx, nil)
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

	row := tx.QueryRow(`
		SELECT deposit_count
		FROM bridge
		WHERE block_num >= $1
		ORDER BY (block_num, block_pos) ASC
		LIMIT 1;
	`)
	var firstDepositCountReorged int
	err = row.Scan(&firstDepositCountReorged)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			firstDepositCountReorged = -1
		} else {
			return err
		}
	}

	_, err = tx.Exec(`DELETE FROM block WHERE block >= $1;`, firstReorgedBlock)
	if err != nil {
		if errRllbck := tx.Rollback(); errRllbck != nil {
			log.Errorf("error while rolling back tx %v", errRllbck)
		}
		return err
	}

	exitTreeRollback := func() {}
	if firstDepositCountReorged != -1 {
		if exitTreeRollback, err = p.exitTree.Reorg(tx, uint32(firstDepositCountReorged)); err != nil {
			tx.Rollback()
			exitTreeRollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		exitTreeRollback()
		return err
	}
	return nil
}

// ProcessBlock process the events of the block to build the exit tree
// and updates the last processed block (can be called without events for that purpose)
func (p *processor) ProcessBlock(ctx context.Context, block sync.Block) error {
	tx, err := p.db.BeginRw(ctx)
	if err != nil {
		return err
	}
	leaves := []tree.Leaf{}
	if len(block.Events) > 0 {
		events := []Event{}
		for _, e := range block.Events {
			event := e.(Event)
			events = append(events, event)
			if event.Bridge != nil {
				leaves = append(leaves, tree.Leaf{
					Index: event.Bridge.DepositCount,
					Hash:  event.Bridge.Hash(),
				})
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

	exitTreeRollback, err := p.exitTree.AddLeaves(tx, leaves)
	if err != nil {
		tx.Rollback()
		exitTreeRollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		exitTreeRollback()
		return err
	}
	p.log.Debugf("processed %d events until block %d", len(block.Events), block.Num)
	return nil
}

func (p *processor) updateLastProcessedBlock(tx kv.RwTx, blockNum uint64) error {
	blockNumBytes := dbCommon.Uint64ToBytes(blockNum)
	return tx.Put(p.lastBlockTable, lastBlockKey, blockNumBytes)
}

func GenerateGlobalIndex(mainnetFlag bool, rollupIndex uint32, localExitRootIndex uint32) *big.Int {
	var (
		globalIndexBytes []byte
		buf              [4]byte
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
	return big.NewInt(0).SetBytes(globalIndexBytes)
}
