package reorgdetector

import (
	"context"
	"encoding/json"

	"github.com/ledgerwatch/erigon-lib/kv"
)

const (
	subscriberBlocks = "reorgdetector-subscriberBlocks"
)

func tableCfgFunc(_ kv.TableCfg) kv.TableCfg {
	return kv.TableCfg{
		subscriberBlocks: {},
	}
}

// getTrackedBlocks returns a list of tracked blocks for each subscriber from db
func (rd *ReorgDetector) getTrackedBlocks(ctx context.Context) (map[string]*headersList, error) {
	tx, err := rd.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	cursor, err := tx.Cursor(subscriberBlocks)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	trackedBlocks := make(map[string]*headersList, 0)

	for k, v, err := cursor.First(); k != nil; k, v, err = cursor.Next() {
		if err != nil {
			return nil, err
		}

		var headers []header
		if err := json.Unmarshal(v, &headers); err != nil {
			return nil, err
		}

		trackedBlocks[string(k)] = newHeadersList(headers...)
	}

	return trackedBlocks, nil
}

// saveTrackedBlock saves the tracked block for a subscriber in db and in memory
func (rd *ReorgDetector) saveTrackedBlock(ctx context.Context, id string, b header) error {
	rd.trackedBlocksLock.Lock()

	// this has to go after the lock, because of a possible deadlock
	// between AddBlocksToTrack and detectReorgInTrackedList
	tx, err := rd.db.BeginRw(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	hdrs, ok := rd.trackedBlocks[id]
	if !ok || hdrs.isEmpty() {
		hdrs = newHeadersList(b)
		rd.trackedBlocks[id] = hdrs
	} else {
		hdrs.add(b)
	}
	rd.trackedBlocksLock.Unlock()

	raw, err := json.Marshal(hdrs.getSorted())
	if err != nil {
		return err
	}

	return tx.Put(subscriberBlocks, []byte(id), raw)
}

// updateTrackedBlocksDB updates the tracked blocks for a subscriber in db
func (rd *ReorgDetector) updateTrackedBlocksDB(ctx context.Context, id string, blocks *headersList) error {
	tx, err := rd.db.BeginRw(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	raw, err := json.Marshal(blocks.getSorted())
	if err != nil {
		return err
	}

	if err = tx.Put(subscriberBlocks, []byte(id), raw); err != nil {
		return err
	}

	return nil
}
