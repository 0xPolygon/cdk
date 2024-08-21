package reorgdetector

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/0xPolygon/cdk/log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
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
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
}

type ReorgDetector struct {
	client EthClient
	db     kv.RwDB

	canonicalBlocks *headersList

	trackedBlocksLock sync.RWMutex
	trackedBlocks     map[string]*headersList

	subscriptionsLock sync.RWMutex
	subscriptions     map[string]*Subscription
}

func New(client EthClient, dbPath string) (*ReorgDetector, error) {
	db, err := mdbx.NewMDBX(nil).
		Path(dbPath).
		WithTableCfg(tableCfgFunc).
		Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	return &ReorgDetector{
		client:          client,
		db:              db,
		canonicalBlocks: newHeadersList(),
		trackedBlocks:   make(map[string]*headersList),
		subscriptions:   make(map[string]*Subscription),
	}, nil
}

func (rd *ReorgDetector) Start(ctx context.Context) (err error) {
	// Initially load a full canonical chain
	/*if err := rd.loadCanonicalChain(ctx); err != nil {
		log.Errorf("failed to load canonical chain: %v", err)
	}*/

	// Load tracked blocks
	if rd.trackedBlocks, err = rd.getTrackedBlocks(ctx); err != nil {
		return fmt.Errorf("failed to get tracked blocks: %w", err)
	}

	// Continuously check reorgs in tracked by subscribers blocks
	// TODO: Optimize this process
	go func() {
		ticker := time.NewTicker(time.Second * 2)
		for range ticker.C {
			if err = rd.detectReorgInTrackedList(ctx); err != nil {
				log.Errorf("failed to detect reorgs in tracked blocks: %v", err)
			}
		}
	}()

	// Load and process tracked headers from the DB
	/*if err := rd.loadAndProcessTrackedHeaders(ctx); err != nil {
		return fmt.Errorf("failed to load and process tracked headers: %w", err)
	}*/

	return nil
}

func (rd *ReorgDetector) AddBlockToTrack(ctx context.Context, id string, blockNum uint64, blockHash, parentHash common.Hash) error {
	rd.subscriptionsLock.RLock()
	if sub, ok := rd.subscriptions[id]; !ok {
		rd.subscriptionsLock.RUnlock()
		return ErrNotSubscribed
	} else {
		// In case there are reorgs being processed, wait
		// Note that this also makes any addition to trackedBlocks[id] safe
		sub.pendingReorgsToBeProcessed.Wait()
	}
	rd.subscriptionsLock.RUnlock()

	hdr := newHeader(blockNum, blockHash, parentHash)
	if err := rd.saveTrackedBlock(ctx, id, hdr); err != nil {
		return fmt.Errorf("failed to save tracked block: %w", err)
	}

	return nil
}

// 1. Get finalized block: 10 (latest block 15)
// 2. Get tracked blocks: 4, 7, 9, 12, 14
// 3. Go from 10 till the lowest block in tracked blocks
// 4. Notify about reorg if exists
func (rd *ReorgDetector) detectReorgInTrackedList(ctx context.Context) error {
	// Get the latest finalized block
	lastFinalisedBlock, err := rd.client.HeaderByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
	if err != nil {
		return err
	}

	// Notify subscribers about reorgs
	// TODO: Optimize it
	for id, hdrs := range rd.trackedBlocks {
		// Get the sorted headers
		sorted := hdrs.getSorted()
		if len(sorted) == 0 {
			continue
		}

		// Do not check blocks that are higher than the last finalized block
		if sorted[0].Num > lastFinalisedBlock.Number.Uint64() {
			continue
		}

		// Go from the lowest tracked block till the latest finalized block
		for i := sorted[0].Num; i <= lastFinalisedBlock.Number.Uint64(); i++ {
			// Get the actual header hash from the client
			header, err := rd.client.HeaderByNumber(ctx, big.NewInt(int64(i)))
			if err != nil {
				return err
			}

			// Check if the block hash matches with the actual block hash
			if sorted[0].Hash != header.Hash() {
				// Reorg detected, notify subscriber
				go rd.notifySubscriber(id, sorted[0])
				break
			}
		}
	}

	return nil
}

// loadCanonicalChain loads canonical chain from the latest finalized block till the latest one
func (rd *ReorgDetector) loadCanonicalChain(ctx context.Context) error {
	// Get the latest finalized block
	lastFinalisedBlock, err := rd.client.HeaderByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
	if err != nil {
		return err
	}

	// Get the latest block
	latestBlock, err := rd.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return err
	}

	startFromBlock := lastFinalisedBlock.Number.Uint64()
	if sortedBlocks := rd.canonicalBlocks.getSorted(); len(sortedBlocks) > 0 {
		lastTrackedBlock := sortedBlocks[rd.canonicalBlocks.len()-1].Num
		if lastTrackedBlock < startFromBlock {
			startFromBlock = lastTrackedBlock
		}
	}

	// Load the canonical chain from the last finalized block till the latest block
	for i := startFromBlock; i <= latestBlock.Number.Uint64(); i++ {
		blockHeader, err := rd.client.HeaderByNumber(ctx, big.NewInt(int64(i)))
		if err != nil {
			return fmt.Errorf("failed to fetch block header for block number %d: %w", i, err)
		}

		if err = rd.onNewHeader(ctx, blockHeader); err != nil {
			return fmt.Errorf("failed to process new header: %w", err)
		}
	}

	return nil
}

// onNewHeader processes a new header and checks for reorgs
func (rd *ReorgDetector) onNewHeader(ctx context.Context, header *types.Header) error {
	hdr := newHeader(header.Number.Uint64(), header.Hash(), header.ParentHash)

	// No canonical chain yet
	if rd.canonicalBlocks.isEmpty() {
		// TODO: Fill canonical chain from the last finalized block
		rd.canonicalBlocks.add(hdr)
		return nil
	}

	closestHigherBlock, ok := rd.canonicalBlocks.getClosestHigherBlock(hdr.Num)
	if !ok {
		// No same or higher blocks, only lower blocks exist. Check hashes.
		// Current tracked blocks: N-i, N-i+1, ..., N-1
		sortedBlocks := rd.canonicalBlocks.getSorted()
		closestBlock := sortedBlocks[len(sortedBlocks)-1]
		if closestBlock.Num < hdr.Num-1 {
			// There is a gap between the last block and the given block
			// Current tracked blocks: N-i, <gap>, N
		} else if closestBlock.Num == hdr.Num-1 {
			if closestBlock.Hash != hdr.ParentHash {
				// Block hashes do not match, reorg happened
				rd.processReorg(hdr)
			} else {
				// All good, add the block to the map
				rd.canonicalBlocks.add(hdr)
			}
		} else {
			// This should not happen
			log.Fatal("Unexpected block number comparison")
		}
	} else {
		if closestHigherBlock.Num == hdr.Num {
			if closestHigherBlock.Hash != hdr.Hash {
				// Block has already been tracked and added to the map but with different hash.
				// Current tracked blocks: N-2, N-1, N (given block)
				rd.processReorg(hdr)
			}
		} else if closestHigherBlock.Num >= hdr.Num+1 {
			// The given block is lower than the closest higher block:
			//  N-2, N-1, N (given block), N+1, N+2
			//  OR
			// There is a gap between the current block and the closest higher block
			//  N-2, N-1, N (given block), <gap>, N+i
			rd.processReorg(hdr)
		} else {
			// This should not happen
			log.Fatal("Unexpected block number comparison")
		}
	}

	return nil
}

// processReorg processes a reorg and notifies subscribers
func (rd *ReorgDetector) processReorg(hdr header) {
	hdrs := rd.canonicalBlocks.copy()
	hdrs.add(hdr)

	if reorgedBlock := hdrs.detectReorg(); reorgedBlock != nil {
		// Notify subscribers about the reorg
		rd.notifySubscribers(*reorgedBlock)
	} else {
		// Should not happen, check the logic below
		// log.Fatal("Unexpected reorg detection")
	}
}

// loadAndProcessTrackedHeaders loads tracked headers from the DB and checks for reorgs. Loads in memory.
func (rd *ReorgDetector) loadAndProcessTrackedHeaders(ctx context.Context) error {
	// Load tracked blocks for all subscribers from the DB
	trackedBlocks, err := rd.getTrackedBlocks(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tracked blocks: %w", err)
	}

	rd.trackedBlocks = trackedBlocks

	lastFinalisedBlock, err := rd.client.HeaderByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
	if err != nil {
		return err
	}

	var errGrp errgroup.Group
	for id, blocks := range rd.trackedBlocks {
		id := id
		blocks := blocks

		errGrp.Go(func() error {
			return rd.processTrackedHeaders(ctx, id, blocks, lastFinalisedBlock.Number.Uint64())
		})
	}

	return errGrp.Wait()
}

// processTrackedHeaders processes tracked headers for a subscriber and checks for reorgs
func (rd *ReorgDetector) processTrackedHeaders(ctx context.Context, id string, headers *headersList, finalized uint64) error {
	rd.subscriptionsLock.Lock()
	rd.subscriptions[id] = &Subscription{
		FirstReorgedBlock: make(chan uint64),
		ReorgProcessed:    make(chan bool),
	}
	rd.subscriptionsLock.Unlock()

	// Nothing to process for this subscriber
	if headers.isEmpty() {
		return nil
	}

	var (
		lastTrackedBlock uint64
		actualBlockHash  common.Hash
	)

	sortedBlocks := headers.getSorted()
	lastTrackedBlock = sortedBlocks[headers.len()-1].Num

	for _, block := range sortedBlocks {
		// Fetch an actual header hash from the client if it does not exist in the canonical chain
		if actualHeader := rd.canonicalBlocks.get(block.Num); actualHeader == nil {
			header, err := rd.client.HeaderByNumber(ctx, big.NewInt(int64(block.Num)))
			if err != nil {
				return err
			}

			actualBlockHash = header.Hash()
		} else {
			actualBlockHash = actualHeader.Hash
		}

		// Check if the block hash matches with the actual block hash
		if actualBlockHash != block.Hash {
			// Reorg detected, notify subscriber
			go rd.notifySubscriber(id, block)

			// Remove the reorged blocks from the tracked blocks
			headers.removeRange(block.Num, lastTrackedBlock)

			break
		} else if block.Num <= finalized {
			headers.removeRange(block.Num, block.Num)
		}
	}

	// If we processed finalized or reorged blocks, update the tracked blocks in memory and db
	return rd.updateTrackedBlocksNoLock(ctx, id, headers)
}
