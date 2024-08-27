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
	"golang.org/x/sync/errgroup"
)

type EthClient interface {
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
}

type ReorgDetector struct {
	client             EthClient
	db                 kv.RwDB
	checkReorgInterval time.Duration

	trackedBlocksLock sync.RWMutex
	trackedBlocks     map[string]*headersList

	subscriptionsLock sync.RWMutex
	subscriptions     map[string]*Subscription
}

func New(client EthClient, cfg Config) (*ReorgDetector, error) {
	db, err := mdbx.NewMDBX(nil).
		Path(cfg.DBPath).
		WithTableCfg(tableCfgFunc).
		Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	return &ReorgDetector{
		client:             client,
		db:                 db,
		checkReorgInterval: cfg.GetCheckReorgsInterval(),
		trackedBlocks:      make(map[string]*headersList),
		subscriptions:      make(map[string]*Subscription),
	}, nil
}

// Start starts the reorg detector
func (rd *ReorgDetector) Start(ctx context.Context) (err error) {
	// Load tracked blocks from the DB
	if err = rd.loadTrackedHeaders(ctx); err != nil {
		return fmt.Errorf("failed to load tracked headers: %w", err)
	}

	// Continuously check reorgs in tracked by subscribers blocks
	go func() {
		ticker := time.NewTicker(rd.checkReorgInterval)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err = rd.detectReorgInTrackedList(ctx); err != nil {
					log.Errorf("failed to detect reorg in tracked list: %v", err)
				}
			}
		}
	}()

	return nil
}

// AddBlockToTrack adds a block to the tracked list for a subscriber
func (rd *ReorgDetector) AddBlockToTrack(ctx context.Context, id string, num uint64, hash common.Hash) error {
	// Skip if the given block has already been stored
	rd.trackedBlocksLock.RLock()
	trackedBlocks, ok := rd.trackedBlocks[id]
	if !ok {
		rd.trackedBlocksLock.RUnlock()
		return fmt.Errorf("subscriber %s is not subscribed", id)
	}
	rd.trackedBlocksLock.RUnlock()

	if existingHeader := trackedBlocks.get(num); existingHeader != nil && existingHeader.Hash == hash {
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
func (rd *ReorgDetector) detectReorgInTrackedList(ctx context.Context) error {
	// Get the latest finalized block
	lastFinalisedBlock, err := rd.client.HeaderByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
	if err != nil {
		return fmt.Errorf("failed to get the latest finalized block: %w", err)
	}

	var (
		headersCacheLock sync.Mutex
		headersCache     = map[uint64]*types.Header{
			lastFinalisedBlock.Number.Uint64(): lastFinalisedBlock,
		}
		errGroup errgroup.Group
	)

	rd.trackedBlocksLock.Lock()
	defer rd.trackedBlocksLock.Unlock()

	for id, hdrs := range rd.trackedBlocks {
		id := id
		hdrs := hdrs

		errGroup.Go(func() error {
			headers := hdrs.getSorted()
			for _, hdr := range headers {
				// Get the actual header from the network or from the cache
				headersCacheLock.Lock()
				currentHeader, ok := headersCache[hdr.Num]
				if !ok || currentHeader == nil {
					if currentHeader, err = rd.client.HeaderByNumber(ctx, big.NewInt(int64(hdr.Num))); err != nil {
						headersCacheLock.Unlock()
						return fmt.Errorf("failed to get the header: %w", err)
					}
					headersCache[hdr.Num] = currentHeader
				}
				headersCacheLock.Unlock()

				// Check if the block hash matches with the actual block hash
				if hdr.Hash == currentHeader.Hash() {
					// Delete block from the tracked blocks list if it is less than or equal to the last finalized block
					// and hashes matches. If higher than finalized block, we assume a reorg still might happen.
					if hdr.Num <= lastFinalisedBlock.Number.Uint64() {
						hdrs.removeRange(hdr.Num, hdr.Num)
					}

					continue
				}

				// Notify the subscriber about the reorg
				rd.notifySubscriber(id, hdr)

				// Remove the reorged block and all the following blocks
				hdrs.removeRange(hdr.Num, headers[len(headers)-1].Num)

				break
			}

			// Update the tracked blocks in the DB
			if err := rd.updateTrackedBlocksDB(ctx, id, hdrs); err != nil {
				return fmt.Errorf("failed to update tracked blocks for subscriber %s: %w", id, err)
			}

			return nil
		})
	}

	return errGroup.Wait()
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
		rd.subscriptions[id] = &Subscription{
			ReorgedBlock:   make(chan uint64),
			ReorgProcessed: make(chan bool),
		}
	}

	return nil
}
