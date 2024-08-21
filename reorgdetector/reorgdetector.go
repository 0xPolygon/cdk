package reorgdetector

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

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

type Subscription struct {
	FirstReorgedBlock          chan uint64
	ReorgProcessed             chan bool
	pendingReorgsToBeProcessed sync.WaitGroup
}

type ReorgDetector struct {
	client EthClient
	db     kv.RwDB

	canonicalBlocksLock sync.RWMutex
	canonicalBlocks     blockMap

	trackedBlocksLock sync.RWMutex
	trackedBlocks     map[string]blockMap

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
		canonicalBlocks: make(blockMap),
		trackedBlocks:   make(map[string]blockMap),
		subscriptions:   make(map[string]*Subscription),
	}, nil
}

func (rd *ReorgDetector) Start(ctx context.Context) error {
	// Load canonical chain from the last finalized block
	if err := rd.loadCanonicalChain(ctx); err != nil {
		return fmt.Errorf("failed to load canonical chain: %w", err)
	}

	// Load and process tracked blocks from the DB
	if err := rd.loadAndProcessTrackedBlocks(ctx); err != nil {
		return fmt.Errorf("failed to load and process tracked blocks: %w", err)
	}

	// Start the reorg detector for the canonical chain
	go rd.monitorCanonicalChain(ctx)

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

	if err := rd.saveTrackedBlock(ctx, id, block{
		Num:        blockNum,
		Hash:       blockHash,
		ParentHash: parentHash,
	}); err != nil {
		return fmt.Errorf("failed to save tracked block: %w", err)
	}

	return nil
}

func (rd *ReorgDetector) monitorCanonicalChain(ctx context.Context) {
	// Add head tracker
	ch := make(chan *types.Header, 100)
	sub, err := rd.client.SubscribeNewHead(ctx, ch)
	if err != nil {
		log.Fatal("failed to subscribe to new head", "err", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-sub.Err():
				return
			case header := <-ch:
				if err = rd.onNewHeader(ctx, header); err != nil {
					log.Error("failed to process new header", "err", err)
					continue
				}
			}
		}
	}()
}

func (rd *ReorgDetector) onNewHeader(ctx context.Context, header *types.Header) error {
	rd.canonicalBlocksLock.Lock()
	defer rd.canonicalBlocksLock.Unlock()

	newBlock := block{
		Num:        header.Number.Uint64(),
		Hash:       header.Hash(),
		ParentHash: header.ParentHash,
	}

	// No canonical chain yet
	if len(rd.canonicalBlocks) == 0 {
		// TODO: Fill canonical chain from the last finalized block
		rd.canonicalBlocks = newBlockMap(newBlock)
		return nil
	}

	processReorg := func() error {
		rd.canonicalBlocks[newBlock.Num] = newBlock
		reorgedBlock := rd.canonicalBlocks.detectReorg()
		if reorgedBlock != nil {
			// Notify subscribers about the reorg
			rd.notifySubscribers(*reorgedBlock)

			// Rebuild the canonical chain
			if err := rd.rebuildCanonicalChain(ctx, reorgedBlock.Num, newBlock.Num); err != nil {
				return fmt.Errorf("failed to rebuild canonical chain: %w", err)
			}
		} else {
			// Should not happen, check the logic below
			// log.Fatal("Unexpected reorg detection")
		}

		return nil
	}

	closestHigherBlock, ok := rd.canonicalBlocks.getClosestHigherBlock(newBlock.Num)
	if !ok {
		// No same or higher blocks, only lower blocks exist. Check hashes.
		// Current tracked blocks: N-i, N-i+1, ..., N-1
		sortedBlocks := rd.canonicalBlocks.getSorted()
		closestBlock := sortedBlocks[len(sortedBlocks)-1]
		if closestBlock.Num < newBlock.Num-1 {
			// There is a gap between the last block and the given block
			// Current tracked blocks: N-i, <gap>, N
		} else if closestBlock.Num == newBlock.Num-1 {
			if closestBlock.Hash != newBlock.ParentHash {
				// Block hashes do not match, reorg happened
				return processReorg()
			} else {
				// All good, add the block to the map
				rd.canonicalBlocks[newBlock.Num] = newBlock
			}
		} else {
			// This should not happen
			log.Fatal("Unexpected block number comparison")
		}
	} else {
		if closestHigherBlock.Num == newBlock.Num {
			// Block has already been tracked and added to the map
			// Current tracked blocks: N-2, N-1, N (given block)
			if closestHigherBlock.Hash != newBlock.Hash {
				// Block hashes have changed, reorg happened
				return processReorg()
			}
		} else if closestHigherBlock.Num >= newBlock.Num+1 {
			// The given block is lower than the closest higher block:
			//  N-2, N-1, N (given block), N+1, N+2
			//  OR
			// There is a gap between the current block and the closest higher block
			//  N-2, N-1, N (given block), <gap>, N+i
			return processReorg()
		} else {
			// This should not happen
			log.Fatal("Unexpected block number comparison")
		}
	}

	return nil
}

// rebuildCanonicalChain rebuilds the canonical chain from the given block number to the given block number.
func (rd *ReorgDetector) rebuildCanonicalChain(ctx context.Context, from, to uint64) error {
	// TODO: Potentially rebuild from the latest finalized block

	for i := from; i <= to; i++ {
		blockHeader, err := rd.client.HeaderByNumber(ctx, big.NewInt(int64(i)))
		if err != nil {
			return fmt.Errorf("failed to fetch block header for block number %d: %w", i, err)
		}

		rd.canonicalBlocks[blockHeader.Number.Uint64()] = block{
			Num:        blockHeader.Number.Uint64(),
			Hash:       blockHeader.Hash(),
			ParentHash: blockHeader.ParentHash,
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

	rd.canonicalBlocksLock.Lock()
	defer rd.canonicalBlocksLock.Unlock()

	// Load the canonical chain from the last finalized block till the latest block
	for i := lastFinalisedBlock.Number.Uint64(); i <= latestBlock.Number.Uint64(); i++ {
		blockHeader, err := rd.client.HeaderByNumber(ctx, big.NewInt(int64(i)))
		if err != nil {
			return fmt.Errorf("failed to fetch block header for block number %d: %w", i, err)
		}

		rd.canonicalBlocks[blockHeader.Number.Uint64()] = block{
			Num:        blockHeader.Number.Uint64(),
			Hash:       blockHeader.Hash(),
			ParentHash: blockHeader.ParentHash,
		}
	}

	return nil
}

// loadAndProcessTrackedBlocks loads tracked blocks from the DB and checks for reorgs. Loads in memory.
func (rd *ReorgDetector) loadAndProcessTrackedBlocks(ctx context.Context) error {
	rd.trackedBlocksLock.Lock()
	defer rd.trackedBlocksLock.Unlock()

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

	for id, blocks := range rd.trackedBlocks {
		rd.subscriptionsLock.Lock()
		rd.subscriptions[id] = &Subscription{
			FirstReorgedBlock: make(chan uint64),
			ReorgProcessed:    make(chan bool),
		}
		rd.subscriptionsLock.Unlock()

		// Nothing to process for this subscriber
		if len(blocks) == 0 {
			continue
		}

		var (
			lastTrackedBlock uint64
			actualBlockHash  common.Hash
		)

		sortedBlocks := blocks.getSorted()
		lastTrackedBlock = sortedBlocks[len(blocks)-1].Num

		for _, block := range sortedBlocks {
			if actualBlock, ok := rd.canonicalBlocks[block.Num]; !ok {
				header, err := rd.client.HeaderByNumber(ctx, big.NewInt(int64(block.Num)))
				if err != nil {
					return err
				}

				actualBlockHash = header.Hash()
			} else {
				actualBlockHash = actualBlock.Hash
			}

			if actualBlockHash != block.Hash {
				// Reorg detected, notify subscriber
				go rd.notifySubscriber(id, block)

				// Remove the reorged blocks from the tracked blocks
				blocks.removeRange(block.Num, lastTrackedBlock)

				break
			} else if block.Num <= lastFinalisedBlock.Number.Uint64() {
				delete(blocks, block.Num)
			}
		}

		// If we processed finalized or reorged blocks, update the tracked blocks in memory and db
		if err = rd.updateTrackedBlocksNoLock(ctx, id, blocks); err != nil {
			return err
		}
	}

	return nil
}
