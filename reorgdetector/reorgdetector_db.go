package reorgdetector

import (
	context "context"
	"encoding/json"
)

// getUnfinalisedBlocksMap returns the map of unfinalised blocks
func (rd *ReorgDetector) getUnfinalisedBlocksMap() blockMap {
	rd.trackedBlocksLock.RLock()
	defer rd.trackedBlocksLock.RUnlock()

	return rd.trackedBlocks[unfinalisedBlocksID]
}

// getTrackedBlocks returns a list of tracked blocks for each subscriber from db
func (rd *ReorgDetector) getTrackedBlocks(ctx context.Context) (map[string]blockMap, error) {
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

	trackedBlocks := make(map[string]blockMap, 0)

	for k, v, err := cursor.First(); k != nil; k, v, err = cursor.Next() {
		if err != nil {
			return nil, err
		}

		var blocks []block
		if err := json.Unmarshal(v, &blocks); err != nil {
			return nil, err
		}

		trackedBlocks[string(k)] = newBlockMap(blocks...)
	}

	if _, ok := trackedBlocks[unfinalisedBlocksID]; !ok {
		// add unfinalised blocks to tracked blocks map if not present in db
		trackedBlocks[unfinalisedBlocksID] = newBlockMap()
	}

	return trackedBlocks, nil
}

// saveTrackedBlock saves the tracked block for a subscriber in db and in memory
func (rd *ReorgDetector) saveTrackedBlock(ctx context.Context, id string, b block) error {
	tx, err := rd.db.BeginRw(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	rd.trackedBlocksLock.Lock()

	subscriberBlockMap, ok := rd.trackedBlocks[id]
	if !ok || len(subscriberBlockMap) == 0 {
		subscriberBlockMap = newBlockMap(b)
		rd.trackedBlocks[id] = subscriberBlockMap
	} else {
		subscriberBlockMap[b.Num] = b
	}

	rd.trackedBlocksLock.Unlock()

	raw, err := json.Marshal(subscriberBlockMap.getSorted())
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

		if err := rd.updateTrackedBlocks(ctx, id, newBlockMap(newTrackedBlocks...)); err != nil {
			return err
		}
	}

	return nil
}

// updateTrackedBlocks updates the tracked blocks for a subscriber in db and in memory
func (rd *ReorgDetector) updateTrackedBlocks(ctx context.Context, id string, blocks blockMap) error {
	rd.trackedBlocksLock.Lock()
	defer rd.trackedBlocksLock.Unlock()

	return rd.updateTrackedBlocksNoLock(ctx, id, blocks)
}

// updateTrackedBlocksNoLock updates the tracked blocks for a subscriber in db and in memory
func (rd *ReorgDetector) updateTrackedBlocksNoLock(ctx context.Context, id string, blocks blockMap) error {
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
