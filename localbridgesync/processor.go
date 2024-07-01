package localbridgesync

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type processor struct {
	db *leveldb.DB
}

func newProcessor() (*processor, error) {
	return &processor{}, errors.New("not implemented")
}

func (p *processor) GetClaimsAndBridges(
	ctx context.Context, fromBlock, toBlock uint64,
) ([]Claim, []Bridge, error) {
	// TODO: if toBlock is not yet synced, return error, however we do not store blocks if they have no events :(?
	claims := []Claim{}
	bridges := []Bridge{}
	from := []byte{}
	to := []byte{}
	binary.LittleEndian.PutUint64(from, fromBlock)
	binary.LittleEndian.PutUint64(to, toBlock)
	iter := p.db.NewIterator(&util.Range{Start: from, Limit: to}, nil)
	for iter.Next() {
		v := iter.Value()
		block := bridgeEvents{}
		err := json.Unmarshal(v, &block)
		if err != nil {
			iter.Release()
			return nil, nil, err
		}
		bridges = append(bridges, block.Bridges...)
		claims = append(claims, block.Claims...)
	}
	iter.Release()
	return claims, bridges, iter.Error()
}

func (p *processor) getLastProcessedBlock(ctx context.Context) (uint64, error) {
	return 0, errors.New("not implemented")
}

func (p *processor) reorg(lastValidBlock uint64) error {
	return errors.New("not implemented")
}

func (p *processor) storeBridgeEvents(blockNum uint64, block bridgeEvents) error {
	value, err := json.Marshal(block)
	if err != nil {
		return err
	}
	key := []byte{}
	binary.LittleEndian.PutUint64(key, blockNum)
	return p.db.Put(key, value, nil)
}
