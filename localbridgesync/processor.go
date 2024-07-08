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
	eventsTable    = "localbridgesync-events"
	lastBlockTable = "localbridgesync-lastBlock"
)

var (
	ErrBlockNotProcessed = errors.New("given block(s) have not been processed yet")
	lastBlokcKey         = []byte("lb")
)

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
) ([]Claim, []Bridge, error) {
	claims := []Claim{}
	bridges := []Bridge{}

	tx, err := p.db.BeginRo(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()
	lpb, err := p.getLastProcessedBlockWithTx(tx)
	if lpb < toBlock {
		return nil, nil, ErrBlockNotProcessed
	}
	c, err := tx.Cursor(eventsTable)
	if err != nil {
		return nil, nil, err
	}
	defer c.Close()

	for k, v, err := c.Seek(blockNum2Bytes(fromBlock)); k != nil; k, v, err = c.Next() {
		if err != nil {
			return nil, nil, err
		}
		if bytes2BlockNum(k) > toBlock {
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
		return bytes2BlockNum(blockNumBytes), nil
	}
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
	firstKey := blockNum2Bytes(firstReorgedBlock)
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

func (p *processor) storeBridgeEvents(blockNum uint64, block bridgeEvents) error {
	tx, err := p.db.BeginRw(context.Background())
	if err != nil {
		return err
	}
	if len(block.Bridges) > 0 || len(block.Claims) > 0 {
		value, err := json.Marshal(block)
		if err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Put(eventsTable, blockNum2Bytes(blockNum), value); err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := p.updateLastProcessedBlock(tx, blockNum); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (p *processor) updateLastProcessedBlock(tx kv.RwTx, blockNum uint64) error {
	blockNumBytes := blockNum2Bytes(blockNum)
	return tx.Put(lastBlockTable, lastBlokcKey, blockNumBytes)
}

func blockNum2Bytes(blockNum uint64) []byte {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, blockNum)
	return key
}

func bytes2BlockNum(key []byte) uint64 {
	return binary.LittleEndian.Uint64(key)
}
