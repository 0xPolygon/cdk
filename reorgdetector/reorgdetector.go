package reorgdetector

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/0xPolygon/cdk/log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

// TODO: consider the case where blocks can disappear, current implementation assumes that if there is a reorg,
// the client will have at least as many blocks as it had before the reorg, however this may not be the case for L2

var (
	ErrNotSubscribed = errors.New("id not found in subscriptions")
)

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

	trackedBlocksLock sync.RWMutex
	trackedBlocks     map[string]blockMap

	subscriptionsLock sync.RWMutex
	subscriptions     map[string]*Subscription
}

func New(client EthClient) *ReorgDetector {
	return &ReorgDetector{
		client:        client,
		trackedBlocks: make(map[string]blockMap),
		subscriptions: make(map[string]*Subscription),
	}
}

func (rd *ReorgDetector) Subscribe(id string) (*Subscription, error) {
	rd.subscriptionsLock.Lock()
	defer rd.subscriptionsLock.Unlock()

	if sub, ok := rd.subscriptions[id]; ok {
		return sub, nil
	}

	sub := &Subscription{
		FirstReorgedBlock: make(chan uint64),
		ReorgProcessed:    make(chan bool),
	}
	rd.subscriptions[id] = sub

	return sub, nil
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

	newBlock := block{
		Num:        blockNum,
		Hash:       blockHash,
		ParentHash: parentHash,
	}

	rd.trackedBlocksLock.Lock()
	defer rd.trackedBlocksLock.Unlock()

	// Get the last block from the list for the given subscription ID
	trackedBlocks, ok := rd.trackedBlocks[id]
	if !ok || len(trackedBlocks) == 0 {
		// No blocks for the given subscription
		trackedBlocks = newBlockMap(newBlock)
		rd.trackedBlocks[id] = trackedBlocks
	}

	findStartReorgBlock := func(blocks blockMap) *block {
		// Find the highest block number
		maxBlockNum := uint64(0)
		for blockNum := range blocks {
			if blockNum > maxBlockNum {
				maxBlockNum = blockNum
			}
		}

		// Iterate from the highest block number to the lowest
		reorgDetected := false
		for i := maxBlockNum; i > 1; i-- {
			currentBlock, currentExists := blocks[i]
			previousBlock, previousExists := blocks[i-1]

			// Check if both blocks exist (sanity check)
			if !currentExists || !previousExists {
				continue
			}

			// Check if the current block's parent hash matches the previous block's hash
			if currentBlock.ParentHash != previousBlock.Hash {
				reorgDetected = true
			} else if reorgDetected {
				// When reorg is detected, and we find the first match, return the previous block
				return &previousBlock
			}
		}

		return nil // No reorg detected
	}

	rebuildBlocksMap := func(blocks blockMap, from, to uint64) (blockMap, error) {
		for i := from; i <= to; i++ {
			blockHeader, err := rd.client.HeaderByNumber(ctx, big.NewInt(int64(i)))
			if err != nil {
				return nil, fmt.Errorf("failed to fetch block header for block number %d: %w", i, err)
			}

			blocks[blockHeader.Number.Uint64()] = block{
				Num:        blockHeader.Number.Uint64(),
				Hash:       blockHeader.Hash(),
				ParentHash: blockHeader.ParentHash,
			}
		}

		return blocks, nil
	}

	processReorg := func() error {
		trackedBlocks[newBlock.Num] = newBlock
		reorgedBlock := findStartReorgBlock(trackedBlocks)
		if reorgedBlock != nil {
			rd.notifySubscribers(*reorgedBlock)

			newBlocksMap, err := rebuildBlocksMap(trackedBlocks, reorgedBlock.Num, newBlock.Num)
			if err != nil {
				return err
			}
			rd.trackedBlocks[id] = newBlocksMap
		} else {
			// Should not happen
		}

		return nil
	}

	closestHigherBlock, ok := trackedBlocks.getClosestHigherBlock(newBlock.Num)
	if !ok {
		// No same or higher blocks, only lower blocks exist. Check hashes.
		// Current tracked blocks: N-i, N-i+1, ..., N-1
		sortedBlocks := trackedBlocks.getSorted()
		closestBlock := sortedBlocks[len(sortedBlocks)-1]
		if closestBlock.Num < newBlock.Num-1 {
			// There is a gap between the last block and the given block
			// Current tracked blocks: N-i, <gap>, N
		} else if closestBlock.Num == newBlock.Num-1 {
			if closestBlock.Hash != newBlock.ParentHash {
				// Block hashes do not match, reorg happened
				// TODO: Reorg happened
				return processReorg()
			} else {
				// All good, add the block to the map
				rd.trackedBlocks[id][newBlock.Num] = newBlock
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
				// TODO: Handle happened
				return processReorg()
			}
		} else if closestHigherBlock.Num == newBlock.Num+1 {
			// The given block is lower than the closest higher block:
			// Current tracked blocks: N-2, N-1, N (given block), N+1, N+2
			// TODO: Reorg happened
			return processReorg()
		} else if closestHigherBlock.Num > newBlock.Num+1 {
			// There is a gap between the current block and the closest higher block
			// Current tracked blocks: N-2, N-1, N (given block), <gap>, N+i
			// TODO: Reorg happened
			return processReorg()
		} else {
			// This should not happen
			log.Fatal("Unexpected block number comparison")
		}
	}

	return nil
}

func (rd *ReorgDetector) notifySubscribers(startingBlock block) {
	rd.subscriptionsLock.RLock()
	for _, sub := range rd.subscriptions {
		sub.pendingReorgsToBeProcessed.Add(1)
		go func(sub *Subscription) {
			sub.FirstReorgedBlock <- startingBlock.Num
			<-sub.ReorgProcessed
			sub.pendingReorgsToBeProcessed.Done()
		}(sub)
	}
	rd.subscriptionsLock.RUnlock()
}
