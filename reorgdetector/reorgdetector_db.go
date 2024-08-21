package reorgdetector

import (
	"context"
	"encoding/json"
)

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

	if _, ok := trackedBlocks[unfinalisedBlocksID]; !ok {
		// add unfinalised blocks to tracked blocks map if not present in db
		trackedBlocks[unfinalisedBlocksID] = newHeadersList()
	}

	return trackedBlocks, nil
}

// saveTrackedBlock saves the tracked block for a subscriber in db and in memory
func (rd *ReorgDetector) saveTrackedBlock(ctx context.Context, id string, b header) error {
	tx, err := rd.db.BeginRw(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	rd.trackedBlocksLock.Lock()
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

// removeTrackedBlocks removes the tracked blocks for a subscriber in db and in memory
func (rd *ReorgDetector) removeTrackedBlocks(ctx context.Context, lastFinalizedBlock uint64) error {
	rd.subscriptionsLock.RLock()
	defer rd.subscriptionsLock.RUnlock()

	for id := range rd.subscriptions {
		rd.trackedBlocksLock.RLock()
		newTrackedBlocks := rd.trackedBlocks[id].getFromBlockSorted(lastFinalizedBlock)
		rd.trackedBlocksLock.RUnlock()

		if err := rd.updateTrackedBlocks(ctx, id, newHeadersList(newTrackedBlocks...)); err != nil {
			return err
		}
	}

	return nil
}

// updateTrackedBlocks updates the tracked blocks for a subscriber in db and in memory
func (rd *ReorgDetector) updateTrackedBlocks(ctx context.Context, id string, blocks *headersList) error {
	rd.trackedBlocksLock.Lock()
	defer rd.trackedBlocksLock.Unlock()

	return rd.updateTrackedBlocksNoLock(ctx, id, blocks)
}

// updateTrackedBlocksNoLock updates the tracked blocks for a subscriber in db and in memory
func (rd *ReorgDetector) updateTrackedBlocksNoLock(ctx context.Context, id string, blocks *headersList) error {
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

	rd.trackedBlocks[id] = blocks

	return nil
}
