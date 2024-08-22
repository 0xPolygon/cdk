package reorgdetector

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

const (
	defaultWaitPeriodBlockRemover = time.Second * 20
	defaultWaitPeriodBlockAdder   = time.Second * 2 // should be smaller than block time of the tracked chain

	subscriberBlocks = "reorgdetector-subscriberBlocks"
)

func tableCfgFunc(_ kv.TableCfg) kv.TableCfg {
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

	notifiedReorgsLock sync.RWMutex
	notifiedReorgs     map[string]map[uint64]struct{}
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
		notifiedReorgs:  make(map[string]map[uint64]struct{}),
	}, nil
}

func (rd *ReorgDetector) Start(ctx context.Context) (err error) {
	// Load tracked blocks from the DB
	if err = rd.loadTrackedHeaders(ctx); err != nil {
		return fmt.Errorf("failed to load tracked headers: %w", err)
	}

	// Initially load a full canonical chain
	if err = rd.loadCanonicalChain(ctx); err != nil {
		log.Errorf("failed to load canonical chain: %v", err)
	}

	// Continuously load canonical chain
	go func() {
		ticker := time.NewTicker(time.Second * 2) // TODO: Configure it
		for range ticker.C {
			if err = rd.loadCanonicalChain(ctx); err != nil {
				log.Errorf("failed to load canonical chain: %v", err)
			}
		}
	}()

	// Continuously check reorgs in tracked by subscribers blocks
	go func() {
		ticker := time.NewTicker(time.Second) // TODO: Configure it
		for range ticker.C {
			rd.detectReorgInTrackedList()
		}
	}()

	return nil
}

// AddBlockToTrack adds a block to the tracked list for a subscriber
func (rd *ReorgDetector) AddBlockToTrack(ctx context.Context, id string, num uint64, hash common.Hash) error {
	// Skip if the given block has already been stored
	rd.trackedBlocksLock.RLock()
	existingHeader := rd.trackedBlocks[id].get(num)
	rd.trackedBlocksLock.RUnlock()

	if existingHeader != nil && existingHeader.Hash == hash {
		return nil
	}

	// Store the given header to the tracked list
	hdr := newHeader(num, hash)
	if err := rd.saveTrackedBlock(ctx, id, hdr); err != nil {
		return fmt.Errorf("failed to save tracked block: %w", err)
	}

	return nil
}

// detectReorgInTrackedList detects reorgs in the tracked blocks.
// Notifies subscribers if reorg has happened
func (rd *ReorgDetector) detectReorgInTrackedList() {
	rd.trackedBlocksLock.RLock()
	var wg sync.WaitGroup
	for id, hdrs := range rd.trackedBlocks {
		wg.Add(1)
		go func(id string, hdrs *headersList) {
			for _, hdr := range hdrs.headers {
				currentHeader := rd.canonicalBlocks.get(hdr.Num)
				if currentHeader == nil {
					break
				}

				// Check if the block hash matches with the actual block hash
				if hdr.Hash == currentHeader.Hash {
					continue
				}

				go rd.notifySubscriber(id, hdr)
			}

			wg.Done()
		}(id, hdrs)

	}
	wg.Wait()
	rd.trackedBlocksLock.RUnlock()
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

	// Start from the last stored block if it less than the last finalized one
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

		// Add the block to the canonical chain
		rd.canonicalBlocks.add(newHeader(blockHeader.Number.Uint64(), blockHeader.Hash()))
	}

	return nil
}

// loadTrackedHeaders loads tracked headers from the DB and stores them in memory
func (rd *ReorgDetector) loadTrackedHeaders(ctx context.Context) (err error) {
	rd.trackedBlocksLock.Lock()
	defer rd.trackedBlocksLock.Unlock()

	// Load tracked blocks for all subscribers from the DB
	if rd.trackedBlocks, err = rd.getTrackedBlocks(ctx); err != nil {
		return fmt.Errorf("failed to get tracked blocks: %w", err)
	}

	// Go over tracked blocks and create subscription for each tracker
	for id := range rd.trackedBlocks {
		_, _ = rd.Subscribe(id)
	}

	return nil
}
