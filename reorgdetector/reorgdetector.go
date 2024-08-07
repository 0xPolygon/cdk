package reorgdetector

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

// TODO: consider the case where blocks can disappear, current implementation assumes that if there is a reorg,
// the client will have at least as many blocks as it had before the reorg, however this may not be the case for L2

const (
	defaultWaitPeriodBlockRemover = time.Second * 20
	defaultWaitPeriodBlockAdder   = time.Second * 2 // should be smaller than block time of the tracked chain

	subscriberBlocks = "reorgdetector-subscriberBlocks"

	unfinalisedBlocksID = "unfinalisedBlocks"
)

var (
	ErrNotSubscribed    = errors.New("id not found in subscriptions")
	ErrInvalidBlockHash = errors.New("the block hash does not match with the expected block hash")
	ErrIDReserverd      = errors.New("subscription id is reserved")
)

func tableCfgFunc(defaultBuckets kv.TableCfg) kv.TableCfg {
	return kv.TableCfg{
		subscriberBlocks: {},
	}
}

type EthClient interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	BlockNumber(ctx context.Context) (uint64, error)
}

type block struct {
	Num  uint64
	Hash common.Hash
}

type blockMap map[uint64]block

// newBlockMap returns a new instance of blockMap
func newBlockMap(blocks ...block) blockMap {
	blockMap := make(blockMap, len(blocks))

	for _, b := range blocks {
		blockMap[b.Num] = b
	}

	return blockMap
}

// getSorted returns blocks in sorted order
func (bm blockMap) getSorted() []block {
	sortedBlocks := make([]block, 0, len(bm))

	for _, b := range bm {
		sortedBlocks = append(sortedBlocks, b)
	}

	sort.Slice(sortedBlocks, func(i, j int) bool {
		return sortedBlocks[i].Num < sortedBlocks[j].Num
	})

	return sortedBlocks
}

// getFromBlockSorted returns blocks from blockNum in sorted order without including the blockNum
func (bm blockMap) getFromBlockSorted(blockNum uint64) []block {
	sortedBlocks := bm.getSorted()

	index := -1
	for i, b := range sortedBlocks {
		if b.Num > blockNum {
			index = i
			break
		}
	}

	if index == -1 {
		return []block{}
	}

	return sortedBlocks[index:]
}

// getClosestHigherBlock returns the closest higher block to the given blockNum
func (bm blockMap) getClosestHigherBlock(blockNum uint64) (block, bool) {
	if block, ok := bm[blockNum]; ok {
		return block, true
	}

	sorted := bm.getFromBlockSorted(blockNum)
	if len(sorted) == 0 {
		return block{}, false
	}

	return sorted[0], true
}

// removeRange removes blocks from "from" to "to"
func (bm blockMap) removeRange(from, to uint64) {
	for i := from; i <= to; i++ {
		delete(bm, i)
	}
}

type Subscription struct {
	FirstReorgedBlock          chan uint64
	ReorgProcessed             chan bool
	pendingReorgsToBeProcessed sync.WaitGroup
}

type ReorgDetector struct {
	ethClient EthClient

	subscriptionsLock sync.RWMutex
	subscriptions     map[string]*Subscription

	trackedBlocksLock sync.RWMutex
	trackedBlocks     map[string]blockMap

	db kv.RwDB

	waitPeriodBlockRemover time.Duration
	waitPeriodBlockAdder   time.Duration
}

type Config struct {
	DBPath string `mapstructure:"DBPath"`
}

// New creates a new instance of ReorgDetector
func New(ctx context.Context, client EthClient, dbPath string) (*ReorgDetector, error) {
	db, err := mdbx.NewMDBX(nil).
		Path(dbPath).
		WithTableCfg(tableCfgFunc).
		Open()
	if err != nil {
		return nil, err
	}

	return newReorgDetector(ctx, client, db)
}

// newReorgDetector creates a new instance of ReorgDetector
func newReorgDetector(ctx context.Context, client EthClient, db kv.RwDB) (*ReorgDetector, error) {
	return newReorgDetectorWithPeriods(ctx, client, db, defaultWaitPeriodBlockRemover, defaultWaitPeriodBlockAdder)
}

// newReorgDetectorWithPeriods creates a new instance of ReorgDetector with custom wait periods
func newReorgDetectorWithPeriods(ctx context.Context, client EthClient, db kv.RwDB,
	waitPeriodBlockRemover, waitPeriodBlockAdder time.Duration) (*ReorgDetector, error) {
	r := &ReorgDetector{
		ethClient:              client,
		db:                     db,
		subscriptions:          make(map[string]*Subscription, 0),
		waitPeriodBlockRemover: waitPeriodBlockRemover,
		waitPeriodBlockAdder:   waitPeriodBlockAdder,
	}

	trackedBlocks, err := r.getTrackedBlocks(ctx)
	if err != nil {
		return nil, err
	}

	r.trackedBlocks = trackedBlocks

	lastFinalisedBlock, err := r.ethClient.HeaderByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
	if err != nil {
		return nil, err
	}

	if err = r.cleanStoredSubsBeforeStart(ctx, lastFinalisedBlock.Number.Uint64()); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *ReorgDetector) Start(ctx context.Context) {
	go r.removeFinalisedBlocks(ctx)
	go r.addUnfinalisedBlocks(ctx)
}

func (r *ReorgDetector) Subscribe(id string) (*Subscription, error) {
	if id == unfinalisedBlocksID {
		return nil, ErrIDReserverd
	}

	r.subscriptionsLock.Lock()
	defer r.subscriptionsLock.Unlock()

	if sub, ok := r.subscriptions[id]; ok {
		return sub, nil
	}
	sub := &Subscription{
		FirstReorgedBlock: make(chan uint64),
		ReorgProcessed:    make(chan bool),
	}
	r.subscriptions[id] = sub

	return sub, nil
}

func (r *ReorgDetector) AddBlockToTrack(ctx context.Context, id string, blockNum uint64, blockHash common.Hash) error {
	r.subscriptionsLock.RLock()
	if sub, ok := r.subscriptions[id]; !ok {
		r.subscriptionsLock.RUnlock()
		return ErrNotSubscribed
	} else {
		// In case there are reorgs being processed, wait
		// Note that this also makes any addition to trackedBlocks[id] safe
		sub.pendingReorgsToBeProcessed.Wait()
	}

	r.subscriptionsLock.RUnlock()

	if actualHash, ok := r.getUnfinalisedBlocksMap()[blockNum]; ok {
		if actualHash.Hash == blockHash {
			return r.saveTrackedBlock(ctx, id, block{Num: blockNum, Hash: blockHash})
		} else {
			return ErrInvalidBlockHash
		}
	} else {
		// ReorgDetector has not added the requested block yet,
		// so we add it to the unfinalised blocks and then to the subscriber blocks as well
		block := block{Num: blockNum, Hash: blockHash}
		if err := r.saveTrackedBlock(ctx, unfinalisedBlocksID, block); err != nil {
			return err
		}

		return r.saveTrackedBlock(ctx, id, block)
	}
}

func (r *ReorgDetector) cleanStoredSubsBeforeStart(ctx context.Context, latestFinalisedBlock uint64) error {
	blocksGotten := make(map[uint64]common.Hash, 0)

	r.trackedBlocksLock.Lock()
	defer r.trackedBlocksLock.Unlock()

	for id, blocks := range r.trackedBlocks {
		r.subscriptionsLock.Lock()
		r.subscriptions[id] = &Subscription{
			FirstReorgedBlock: make(chan uint64),
			ReorgProcessed:    make(chan bool),
		}
		r.subscriptionsLock.Unlock()

		var (
			lastTrackedBlock uint64
			block            block
			actualBlockHash  common.Hash
			ok               bool
		)

		if len(blocks) == 0 {
			continue // nothing to process for this subscriber
		}

		sortedBlocks := blocks.getSorted()
		lastTrackedBlock = sortedBlocks[len(blocks)-1].Num

		for _, block = range sortedBlocks {
			if actualBlockHash, ok = blocksGotten[block.Num]; !ok {
				actualBlock, err := r.ethClient.HeaderByNumber(ctx, big.NewInt(int64(block.Num)))
				if err != nil {
					return err
				}

				actualBlockHash = actualBlock.Hash()
			}

			if actualBlockHash != block.Hash {
				// reorg detected, notify subscriber
				go r.notifyReorgToSubscription(id, block.Num)

				// remove the reorged blocks from the tracked blocks
				blocks.removeRange(block.Num, lastTrackedBlock)

				break
			} else if block.Num <= latestFinalisedBlock {
				delete(blocks, block.Num)
			}
		}

		// if we processed finalized or reorged blocks, update the tracked blocks in memory and db
		if err := r.updateTrackedBlocksNoLock(ctx, id, blocks); err != nil {
			return err
		}
	}

	return nil
}

func (r *ReorgDetector) removeFinalisedBlocks(ctx context.Context) {
	ticker := time.NewTicker(r.waitPeriodBlockRemover)

	for {
		select {
		case <-ticker.C:
			lastFinalisedBlock, err := r.ethClient.HeaderByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
			if err != nil {
				log.Error("reorg detector - error getting last finalised block", "err", err)

				continue
			}

			if err := r.removeTrackedBlocks(ctx, lastFinalisedBlock.Number.Uint64()); err != nil {
				log.Error("reorg detector - error removing tracked blocks", "err", err)

				continue
			}
		case <-ctx.Done():
			return
		}
	}
}

func (r *ReorgDetector) addUnfinalisedBlocks(ctx context.Context) {
	var (
		lastUnfinalisedBlock uint64
		ticker               = time.NewTicker(r.waitPeriodBlockAdder)
		unfinalisedBlocksMap = r.getUnfinalisedBlocksMap()
		prevBlock            *types.Header
		lastBlockFromClient  *types.Header
		err                  error
	)

	if len(unfinalisedBlocksMap) > 0 {
		lastUnfinalisedBlock = unfinalisedBlocksMap.getSorted()[len(unfinalisedBlocksMap)-1].Num
	}

	for {
		select {
		case <-ticker.C:
			lastBlockFromClient, err = r.ethClient.HeaderByNumber(ctx, big.NewInt(int64(rpc.LatestBlockNumber)))
			if err != nil {
				log.Error("reorg detector - error getting last block from client", "err", err)
				continue
			}

			if lastBlockFromClient.Number.Uint64() < lastUnfinalisedBlock {
				// a reorg probably happened, and the client has less blocks than we have
				// we should wait for the client to catch up so we can be sure
				continue
			}

			unfinalisedBlocksMap = r.getUnfinalisedBlocksMap()
			if len(unfinalisedBlocksMap) == 0 {
				// no unfinalised blocks, just add this block to the map
				if err := r.saveTrackedBlock(ctx, unfinalisedBlocksID, block{
					Num:  lastBlockFromClient.Number.Uint64(),
					Hash: lastBlockFromClient.Hash(),
				}); err != nil {
					log.Error("reorg detector - error saving unfinalised block", "block", lastBlockFromClient.Number.Uint64(), "err", err)
				}

				continue
			}

			startBlock := lastBlockFromClient
			unfinalisedBlocksSorted := unfinalisedBlocksMap.getSorted()
			reorgBlock := uint64(0)

			for i := startBlock.Number.Uint64(); i > unfinalisedBlocksSorted[0].Num; i-- {
				previousBlock, ok := unfinalisedBlocksMap[i-1]
				if !ok {
					prevBlock, err = r.ethClient.HeaderByNumber(ctx, big.NewInt(int64(i-1)))
					if err != nil {
						log.Error("reorg detector - error getting previous block", "block", i-1, "err", err)
						break // stop processing blocks, and we will try to detect it in the next iteration
					}

					previousBlock = block{Num: prevBlock.Number.Uint64(), Hash: prevBlock.Hash()}
				}

				if previousBlock.Hash == lastBlockFromClient.ParentHash {
					unfinalisedBlocksMap[i] = block{
						Num:  lastBlockFromClient.Number.Uint64(),
						Hash: lastBlockFromClient.Hash(),
					}
				} else {
					// reorg happened, we will find out from where exactly and report this to subscribers
					reorgBlock = i
				}

				lastBlockFromClient, err = r.ethClient.HeaderByNumber(ctx, big.NewInt(int64(i-1)))
				if err != nil {
					log.Error("reorg detector - error getting last block from client", "err", err)
					break // stop processing blocks, and we will try to detect it in the next iteration
				}
			}

			if err == nil {
				// if we noticed an error, do not notify or update tracked blocks
				if reorgBlock > 0 {
					r.notifyReorgToAllSubscriptions(reorgBlock)
				} else {
					r.updateTrackedBlocks(ctx, unfinalisedBlocksID, unfinalisedBlocksMap)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (r *ReorgDetector) notifyReorgToAllSubscriptions(reorgBlock uint64) {
	r.subscriptionsLock.RLock()
	defer r.subscriptionsLock.RUnlock()

	for id, sub := range r.subscriptions {
		r.trackedBlocksLock.RLock()
		subscriberBlocks := r.trackedBlocks[id]
		r.trackedBlocksLock.RUnlock()

		closestBlock, exists := subscriberBlocks.getClosestHigherBlock(reorgBlock)

		if exists {
			go r.notifyReorgToSub(sub, closestBlock.Num)

			// remove reorged blocks from tracked blocks
			sorted := subscriberBlocks.getSorted()
			subscriberBlocks.removeRange(closestBlock.Num, sorted[len(sorted)-1].Num)
			if err := r.updateTrackedBlocks(context.Background(), id, subscriberBlocks); err != nil {
				log.Error("reorg detector - error updating tracked blocks", "err", err)
			}
		}
	}
}

func (r *ReorgDetector) notifyReorgToSubscription(id string, reorgBlock uint64) {
	if id == unfinalisedBlocksID {
		// unfinalised blocks are not subscribers, and reorg block should be > 0
		return
	}

	r.subscriptionsLock.RLock()
	sub := r.subscriptions[id]
	r.subscriptionsLock.RUnlock()

	r.notifyReorgToSub(sub, reorgBlock)
}

func (r *ReorgDetector) notifyReorgToSub(sub *Subscription, reorgBlock uint64) {
	sub.pendingReorgsToBeProcessed.Add(1)

	// notify about the first reorged block that was tracked
	// and wait for the receiver to process
	sub.FirstReorgedBlock <- reorgBlock
	<-sub.ReorgProcessed

	sub.pendingReorgsToBeProcessed.Done()
}

// getUnfinalisedBlocksMap returns the map of unfinalised blocks
func (r *ReorgDetector) getUnfinalisedBlocksMap() blockMap {
	r.trackedBlocksLock.RLock()
	defer r.trackedBlocksLock.RUnlock()

	return r.trackedBlocks[unfinalisedBlocksID]
}

// getTrackedBlocks returns a list of tracked blocks for each subscriber from db
func (r *ReorgDetector) getTrackedBlocks(ctx context.Context) (map[string]blockMap, error) {
	tx, err := r.db.BeginRo(ctx)
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
func (r *ReorgDetector) saveTrackedBlock(ctx context.Context, id string, b block) error {
	tx, err := r.db.BeginRw(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	r.trackedBlocksLock.Lock()

	subscriberBlockMap, ok := r.trackedBlocks[id]
	if !ok {
		subscriberBlockMap = newBlockMap(b)
		r.trackedBlocks[id] = subscriberBlockMap
	}

	r.trackedBlocksLock.Unlock()

	raw, err := json.Marshal(subscriberBlockMap.getSorted())
	if err != nil {
		return err
	}

	return tx.Put(subscriberBlocks, []byte(id), raw)
}

// removeTrackedBlocks removes the tracked blocks for a subscriber in db and in memory
func (r *ReorgDetector) removeTrackedBlocks(ctx context.Context, lastFinalizedBlock uint64) error {
	r.subscriptionsLock.RLock()
	defer r.subscriptionsLock.RUnlock()

	for id := range r.subscriptions {
		r.trackedBlocksLock.RLock()
		newTrackedBlocks := r.trackedBlocks[id].getFromBlockSorted(lastFinalizedBlock)
		r.trackedBlocksLock.RUnlock()

		if err := r.updateTrackedBlocks(ctx, id, newBlockMap(newTrackedBlocks...)); err != nil {
			return err
		}
	}

	return nil
}

// updateTrackedBlocks updates the tracked blocks for a subscriber in db and in memory
func (r *ReorgDetector) updateTrackedBlocks(ctx context.Context, id string, blocks blockMap) error {
	r.trackedBlocksLock.Lock()
	defer r.trackedBlocksLock.Unlock()

	return r.updateTrackedBlocksNoLock(ctx, id, blocks)
}

// updateTrackedBlocksNoLock updates the tracked blocks for a subscriber in db and in memory
func (r *ReorgDetector) updateTrackedBlocksNoLock(ctx context.Context, id string, blocks blockMap) error {
	tx, err := r.db.BeginRw(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	raw, err := json.Marshal(blocks.getSorted())
	if err != nil {
		return err
	}

	if err := tx.Put(subscriberBlocks, []byte(id), raw); err != nil {
		return err
	}

	r.trackedBlocks[id] = blocks

	return nil
}
