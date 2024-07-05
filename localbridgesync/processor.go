package localbridgesync

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type processor struct {
	processCh chan batch
	db        *leveldb.DB
}

func newProcessor() (*processor, error) {
	return &processor{}, errors.New("not implemented")
}

func (p *processor) GetClaimsAndBridges(
	ctx context.Context, fromBatchNum, toBatchNum uint64,
) ([]Claim, []Bridge, error) {
	// TODO: if toBatchNum is not yet synced, return error
	claims := []Claim{}
	bridges := []Bridge{}
	from := []byte{}
	to := []byte{}
	binary.LittleEndian.PutUint64(from, fromBatchNum)
	binary.LittleEndian.PutUint64(to, toBatchNum)
	iter := p.db.NewIterator(&util.Range{Start: from, Limit: to}, nil)
	for iter.Next() {
		v := iter.Value()
		batch := bridgeEvents{}
		err := json.Unmarshal(v, &batch)
		if err != nil {
			iter.Release()
			return nil, nil, err
		}
		bridges = append(bridges, batch.Bridges...)
		claims = append(claims, batch.Claims...)
	}
	iter.Release()
	return claims, bridges, iter.Error()
}

func (p *processor) process() {
	for {
		batch := <-p.processCh
		err := p.storeBridgeEvents(batch.BatchNum, batch.Events)
		for err != nil {
			// TODO: replace with log.Errorf once the repo have it
			fmt.Println(err)
			time.Sleep(time.Second)
			err = p.storeBridgeEvents(batch.BatchNum, batch.Events)
		}
	}
}

func (p *processor) reorg(lastValidBatch uint64) error {
	return errors.New("not implemented")
}

func (p *processor) storeBridgeEvents(batchNum uint64, batch bridgeEvents) error {
	value, err := json.Marshal(batch)
	if err != nil {
		return err
	}
	key := []byte{}
	binary.LittleEndian.PutUint64(key, batchNum)
	return p.db.Put(key, value, nil)
}
