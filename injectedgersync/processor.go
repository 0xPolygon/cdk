package injectedgersync

import (
	"context"
	"database/sql"
	"errors"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/injectedgersync/migrations"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/russross/meddler"
)

var (
	// TODO: use db.ErrNotFound instead
	ErrNotFound = errors.New("not found")
)

type InjectedGER struct {
	BlockNum        uint64         `meddler:"block_num"`
	BlockPos        uint64         `meddler:"block_pos"`
	L1InfoTreeIndex uint32         `meddler:"l1_info_tree_index"`
	GlobalExitRoot  ethCommon.Hash `meddler:"global_exit_root,hash"`
}

type blockWithGERs struct {
	// inclusive
	FirstIndex uint32
	// not inclusive
	LastIndex uint32
}

type processor struct {
	db *sql.DB
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
		db: db,
	}, nil
}

// GetLastProcessedBlockAndL1InfoTreeIndex returns the last processed block oby the processor, including blocks
// that don't have events
func (p *processor) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	return p.getLastProcessedBlockWithTx(p.db)
}

func (p *processor) getLastIndex() (uint32, error) {
	var lastIndex uint32
	row := p.db.QueryRow("SELECT l1_info_tree_index FROM injected_ger ORDER BY l1_info_tree_index DESC LIMIT 1;")
	err := row.Scan(&lastIndex)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return lastIndex, err
}

func (p *processor) getLastProcessedBlockWithTx(tx db.DBer) (uint64, error) {
	var lastProcessedBlock uint64
	row := tx.QueryRow("SELECT num FROM BLOCK ORDER BY num DESC LIMIT 1;")
	err := row.Scan(&lastProcessedBlock)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return lastProcessedBlock, err
}

func (p *processor) ProcessBlock(ctx context.Context, block sync.Block) error {
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

	if _, err := tx.Exec(`INSERT INTO block (num) VALUES ($1)`, block.Num); err != nil {
		return err
	}

	for _, e := range block.Events {
		event := e.(InjectedGER)
		if err = meddler.Insert(tx, "injected_ger", event); err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (p *processor) Reorg(ctx context.Context, firstReorgedBlock uint64) error {
	_, err := p.db.Exec(`DELETE FROM block WHERE num >= $1;`, firstReorgedBlock)
	return err
}

// GetFirstGERAfterL1InfoTreeIndex returns the first GER injected on the chain that is related to l1InfoTreeIndex
// or greater
func (p *processor) GetFirstGERAfterL1InfoTreeIndex(ctx context.Context, l1InfoTreeIndex uint32) (*InjectedGER, error) {
	injectedGER := &InjectedGER{}
	if err := meddler.QueryRow(p.db, injectedGER, `
		SELECT * FROM injected_ger
		WHERE l1_info_tree_index >= $1
		ORDER BY l1_info_tree_index ASC
		LIMIT 1;
	`, l1InfoTreeIndex); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return injectedGER, nil
}
