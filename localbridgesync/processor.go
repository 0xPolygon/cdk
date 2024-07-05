package localbridgesync

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

const (
	eventsTable = "localbridgesync-events"
)

var (
	ErrBlockNotProcessed = errors.New("given block(s) have not been processed yet")
)

type processor struct {
	db kv.RwDB
}

func tableCfgFunc(defaultBuckets kv.TableCfg) kv.TableCfg {
	return kv.TableCfg{
		eventsTable: {},
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

// GetClaimsAndBridges returns the claims and bridges occurred between fromBlock, toBlock both included
func (p *processor) GetClaimsAndBridges(
	ctx context.Context, fromBlock, toBlock uint64,
) ([]Claim, []Bridge, error) {
	// TODO: if toBlock is not yet synced, return error, however we do not store blocks if they have no events :(?
	claims := []Claim{}
	bridges := []Bridge{}

	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()
	c, err := tx.Cursor(eventsTable)
	if err != nil {
		return nil, nil, err
	}
	defer c.Close()

	for k, v, err := c.Seek(blockNum2Key(fromBlock)); k != nil; k, v, err = c.Next() {
		if err != nil {
			return nil, nil, err
		}
		if key2BlockNum(k) > toBlock {
			break
		}
		block := bridgeEvents{}
		err := json.Unmarshal(v, &block)
		if err != nil {
			return nil, nil, err
		}
		bridges = append(bridges, block.Bridges...)
		claims = append(claims, block.Claims...)
	}

	return claims, bridges, nil
}

func (p *processor) getLastProcessedBlock(ctx context.Context) (uint64, error) {
	return 0, errors.New("not implemented")
}

func (p *processor) reorg(firstReorgedBlock uint64) error {
	tx, err := p.db.BeginRw(context.Background())
	if err != nil {
		return err
	}
	c, err := tx.Cursor(eventsTable)
	if err != nil {
		return err
	}
	defer c.Close()
	firstKey := blockNum2Key(firstReorgedBlock)
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
	return tx.Commit()
}

func (p *processor) storeBridgeEvents(blockNum uint64, block bridgeEvents) error {
	// TODO: add logic to store last processed block even if there are no events
	value, err := json.Marshal(block)
	if err != nil {
		return err
	}
	tx, err := p.db.BeginRw(context.Background())
	if err != nil {
		return err
	}
	if err := tx.Put(eventsTable, blockNum2Key(blockNum), value); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func blockNum2Key(blockNum uint64) []byte {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, blockNum)
	return key
}

func key2BlockNum(key []byte) uint64 {
	return binary.LittleEndian.Uint64(key)
}
