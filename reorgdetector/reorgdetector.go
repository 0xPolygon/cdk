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
	waitPeriodBlockRemover = time.Second * 20
	waitPeriodBlockAdder   = time.Second * 2 // should be smaller than block time of the tracked chain

	subscriberBlocks = "reorgdetector-subscriberBlocks"

	unfalisedBlocksID = "unfinalisedBlocks"
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
	numOfBlocks := len(sortedBlocks)
	lastBlock := sortedBlocks[numOfBlocks-1].Num

	if blockNum < lastBlock {
		numOfBlocksToLeave := int(lastBlock - blockNum)

		if numOfBlocksToLeave > numOfBlocks {
			numOfBlocksToLeave %= numOfBlocks
		}

		newBlocks := make([]block, 0, numOfBlocksToLeave)
		for i := numOfBlocks - numOfBlocksToLeave; i < numOfBlocks; i++ {
			if sortedBlocks[i].Num < lastBlock {
				// skip blocks that are finalised
				continue
			}

			newBlocks = append(newBlocks, sortedBlocks[i])
		}

		sortedBlocks = newBlocks
	} else {
		sortedBlocks = []block{}
	}

	return sortedBlocks
}

type Subscription struct {
	FirstReorgedBlock          chan uint64
	ReorgProcessed             chan bool
	pendingReorgsToBeProcessed *sync.WaitGroup
}

type ReorgDetector struct {
	ethClient     EthClient
	subscriptions map[string]*Subscription

	trackedBlocksLock sync.RWMutex
	trackedBlocks     map[string]blockMap

	db kv.RwDB
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
	r := &ReorgDetector{
		ethClient:     client,
		db:            db,
		subscriptions: make(map[string]*Subscription, 0),
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
	unfinalisedBlocks, ok := r.trackedBlocks[unfalisedBlocksID]
	if !ok {
		unfinalisedBlocks = newBlockMap()
	}

	var lastUnfinalisedBlock uint64
	if len(unfinalisedBlocks) > 0 {
		lastUnfinalisedBlock = unfinalisedBlocks.getSorted()[len(unfinalisedBlocks)-1].Num
	}

	go r.removeFinalisedBlocks(ctx)
	go r.addUnfinalisedBlocks(ctx, lastUnfinalisedBlock+1)
}

func (r *ReorgDetector) Subscribe(id string) (*Subscription, error) {
	if id == unfalisedBlocksID {
		return nil, ErrIDReserverd
	}

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
	if sub, ok := r.subscriptions[id]; !ok {
		return ErrNotSubscribed
	} else {
		// In case there are reorgs being processed, wait
		// Note that this also makes any addition to trackedBlocks[id] safe
		sub.pendingReorgsToBeProcessed.Wait()
	}

	if actualHash, ok := r.getUnfinalisedBlocksMap()[blockNum]; ok {
		if actualHash.Hash == blockHash {
			return r.saveTrackedBlock(ctx, id, block{Num: blockNum, Hash: blockHash})
		} else {
			return ErrInvalidBlockHash
		}
	} else {
		// block not found in local storage
		lastFinalisedBlock, err := r.ethClient.HeaderByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
		if err != nil {
			return err
		}
		if lastFinalisedBlock.Number.Uint64() >= blockNum {
			// block already finalised, no need to track
			return nil
		} else {
			// ReorgDetector has not added the requested block yet, adding it
			return r.saveTrackedBlock(ctx, id, block{Num: blockNum, Hash: blockHash})
		}
	}
}

func (r *ReorgDetector) cleanStoredSubsBeforeStart(ctx context.Context, latestFinalisedBlock uint64) error {
	blocksGotten := make(map[uint64]common.Hash, 0)

	r.trackedBlocksLock.Lock()
	defer r.trackedBlocksLock.Unlock()

	for id, blocks := range r.trackedBlocks {
		r.subscriptions[id] = &Subscription{
			FirstReorgedBlock: make(chan uint64),
			ReorgProcessed:    make(chan bool),
		}

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

		for _, block = range blocks {
			if actualBlockHash, ok = blocksGotten[block.Num]; !ok {
				actualBlock, err := r.ethClient.HeaderByNumber(ctx, big.NewInt(int64(block.Num)))
				if err != nil {
					return err
				}

				actualBlockHash = actualBlock.Hash()
			}

			if actualBlockHash != block.Hash {
				// reorg detected, notify subscriber
				if id != unfalisedBlocksID {
					r.subscriptions[id].pendingReorgsToBeProcessed.Add(1)
					go r.notifyReorgToSubscription(id, block.Num, lastTrackedBlock)
				}

				break
			} else if block.Num <= latestFinalisedBlock {
				delete(blocks, block.Num)
			}
		}

		if err := r.updateTrackedBlocksNoLock(ctx, id, blocks); err != nil {
			return err
		}
	}

	return nil
}

func (r *ReorgDetector) removeFinalisedBlocks(ctx context.Context) {
	ticker := time.NewTicker(waitPeriodBlockRemover)

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

func (r *ReorgDetector) addUnfinalisedBlocks(ctx context.Context, initBlock uint64) {
	var (
		firstBlockReorged uint64
		err               error
	)
	currentBlock := initBlock
	lastBlockFromClient := currentBlock - 1
	for {
		for currentBlock > lastBlockFromClient {
			lastBlockFromClient, err = r.ethClient.BlockNumber(ctx)
			if err != nil {
				log.Error("reorg detector - error getting last block number from client", "err", err)
				time.Sleep(waitPeriodBlockAdder)

				continue
			}

			time.Sleep(waitPeriodBlockAdder)
		}

		header, err := r.ethClient.HeaderByNumber(ctx, big.NewInt(int64(currentBlock)))
		if err != nil {
			log.Error("reorg detector - error getting block header from client", "err", err)
		}

		unfinalisedBlocks := r.getUnfinalisedBlocksMap()

		prevBlock, ok := unfinalisedBlocks[currentBlock-1]
		if !ok || header.ParentHash == prevBlock.Hash {
			r.saveTrackedBlock(ctx, unfalisedBlocksID, block{header.Number.Uint64(), header.Hash()})
			if firstBlockReorged > 0 {
				r.notifyReorg(currentBlock, firstBlockReorged)
			}
			currentBlock++
			firstBlockReorged = 0
		} else if header.ParentHash != prevBlock.Hash {
			// previous block is reorged:
			// 1. add a pending reorg to be processed for all the subscribers (so they don't add more blocks)
			// 2. remove block
			for _, sub := range r.subscriptions {
				sub.pendingReorgsToBeProcessed.Add(1)
			}

			r.removeTrackedBlock(ctx, unfalisedBlocksID, currentBlock-1)

			currentBlock--
			if firstBlockReorged == 0 {
				firstBlockReorged = currentBlock
			}
		}
	}
}

func (r *ReorgDetector) notifyReorg(fromBlock, toBlock uint64) {
	for id := range r.subscriptions {
		go r.notifyReorgToSubscription(id, fromBlock, toBlock)
	}
}

func (r *ReorgDetector) notifyReorgToSubscription(id string, fromBlock, toBlock uint64) {
	r.trackedBlocksLock.RLock()
	blocks := r.trackedBlocks[id]
	r.trackedBlocksLock.RUnlock()

	sub := r.subscriptions[id]
	var found bool
	for i := fromBlock; i <= toBlock; i++ {
		if _, ok := blocks[i]; ok {
			if !found {
				// notify about the first reorged block that was tracked
				// and wait for the receiver to process
				found = true
				sub.FirstReorgedBlock <- i
				<-sub.ReorgProcessed
			}
			delete(blocks, i)
		}
	}

	sub.pendingReorgsToBeProcessed.Done()
}

// getUnfinalisedBlocksMap returns the map of unfinalised blocks
func (r *ReorgDetector) getUnfinalisedBlocksMap() blockMap {
	r.trackedBlocksLock.RLock()
	defer r.trackedBlocksLock.RUnlock()

	return r.trackedBlocks[unfalisedBlocksID]
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

	if _, ok := trackedBlocks[unfalisedBlocksID]; !ok {
		// add unfinalised blocks to tracked blocks map if not present in db
		trackedBlocks[unfalisedBlocksID] = newBlockMap()
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

// removeTrackedBlock removes the tracked block for a subscriber in db and in memory
func (r *ReorgDetector) removeTrackedBlock(ctx context.Context, id string, blockNum uint64) error {
	r.trackedBlocksLock.Lock()
	defer r.trackedBlocksLock.Unlock()

	subscriberBlockMap, ok := r.trackedBlocks[id]
	if !ok {
		return nil
	}

	delete(subscriberBlockMap, blockNum)

	return r.updateTrackedBlocksNoLock(ctx, id, subscriberBlockMap)
}

// removeTrackedBlocks removes the tracked blocks for a subscriber in db and in memory
func (r *ReorgDetector) removeTrackedBlocks(ctx context.Context, lastFinalizedBlock uint64) error {
	r.trackedBlocksLock.Lock()
	defer r.trackedBlocksLock.Unlock()

	for id := range r.trackedBlocks {
		newTrackedBlocks := r.trackedBlocks[id].getFromBlockSorted(lastFinalizedBlock)

		if err := r.updateTrackedBlocksNoLock(ctx, id, newBlockMap(newTrackedBlocks...)); err != nil {
			return err
		}
	}

	return nil
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
