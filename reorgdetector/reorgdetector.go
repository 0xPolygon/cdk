package reorgdetector

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// TODO: consider the case where blocks can disappear, current implementation assumes that if there is a reorg,
// the client will have at least as many blocks as it had before the reorg, however this may not be the case for L2

const (
	waitPeriodBlockRemover = time.Second * 20
	waitPeriodBlockAdder   = time.Second * 2 // should be smaller than block time of the tracked chain
)

var (
	ErrNotSubscribed    = errors.New("id not found in subscriptions")
	ErrInavlidBlockHash = errors.New("the block hash does not match with the expected block hash")
)

type Subscription struct {
	FirstReorgedBlock          chan uint64
	ReorgProcessed             chan bool
	mu                         sync.Mutex
	pendingReorgsToBeProcessed *sync.WaitGroup
}

type ReorgDetector struct {
	ethClient         *ethclient.Client
	mu                sync.Mutex
	unfinalisedBlocks map[uint64]common.Hash
	trackedBlocks     map[string]map[uint64]common.Hash // TODO: needs persistance! needs to be able to iterate in order!
	// the channel is used to notify first invalid block
	subscriptions map[string]*Subscription
}

func New(ctx context.Context) (*ReorgDetector, error) {
	r := &ReorgDetector{}
	lastFinalisedBlock, err := r.ethClient.HeaderByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
	if err != nil {
		return nil, err
	}
	r.unfinalisedBlocks[lastFinalisedBlock.Number.Uint64()] = lastFinalisedBlock.Hash()
	err = r.cleanStoredSubsBeforeStart(ctx, lastFinalisedBlock.Number.Uint64())
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *ReorgDetector) Start(ctx context.Context) {
	var lastFinalisedBlock uint64
	for lastFinalisedBlock = range r.unfinalisedBlocks {
	}
	go r.removeFinalisedBlocks(ctx)
	go r.addUnfinalisedBlocks(ctx, lastFinalisedBlock+1)
}

func (r *ReorgDetector) Subscribe(id string) *Subscription {
	if sub, ok := r.subscriptions[id]; ok {
		return sub
	}
	sub := &Subscription{
		FirstReorgedBlock: make(chan uint64),
		ReorgProcessed:    make(chan bool),
	}
	r.subscriptions[id] = sub
	r.trackedBlocks[id] = make(map[uint64]common.Hash)
	return sub
}

func (r *ReorgDetector) AddBlockToTrack(ctx context.Context, id string, blockNum uint64, blockHash common.Hash) error {
	if sub, ok := r.subscriptions[id]; !ok {
		return ErrNotSubscribed
	} else {
		// In case there are reorgs being processed, wait
		// Note that this also makes any addition to trackedBlocks[id] safe
		sub.mu.Lock()
		defer sub.mu.Unlock()
		sub.pendingReorgsToBeProcessed.Wait()
	}
	if actualHash, ok := r.unfinalisedBlocks[blockNum]; ok {
		if actualHash == blockHash {
			r.trackedBlocks[id][blockNum] = blockHash
			return nil
		} else {
			return ErrInavlidBlockHash
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
			r.trackedBlocks[id][blockNum] = blockHash
			return nil
		}
	}
}

func (r *ReorgDetector) cleanStoredSubsBeforeStart(ctx context.Context, latestFinalisedBlock uint64) error {
	for id := range r.trackedBlocks {
		sub := &Subscription{
			FirstReorgedBlock: make(chan uint64),
			ReorgProcessed:    make(chan bool),
		}
		r.subscriptions[id] = sub
		if err := r.cleanStoredSubBeforeStart(ctx, id, latestFinalisedBlock); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReorgDetector) cleanStoredSubBeforeStart(ctx context.Context, id string, latestFinalisedBlock uint64) error {
	blocks := r.trackedBlocks[id]
	var lastTrackedBlock uint64                               // TODO: get the greatest block num tracked
	for expectedBlockNum, expectedBlockHash := range blocks { // should iterate in order
		actualBlock, err := r.ethClient.HeaderByNumber(ctx, big.NewInt(int64(expectedBlockNum)))
		if err != nil {
			return err
		}
		if actualBlock.Hash() != expectedBlockHash {
			r.subscriptions[id].pendingReorgsToBeProcessed.Add(1)
			go r.notifyReorgToSubscription(id, expectedBlockNum, lastTrackedBlock)
			return nil
		} else if expectedBlockNum < latestFinalisedBlock {
			delete(blocks, expectedBlockNum)
		}
	}
	return nil
}

func (r *ReorgDetector) removeFinalisedBlocks(ctx context.Context) {
	for {
		lastFinalisedBlock, err := r.ethClient.HeaderByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
		if err != nil {
			// TODO: handle error
			return
		}
		for i := lastFinalisedBlock.Number.Uint64(); i >= 0; i-- {
			if _, ok := r.unfinalisedBlocks[i]; ok {
				r.mu.Lock()
				delete(r.unfinalisedBlocks, i)
				r.mu.Unlock()
			} else {
				break
			}
			for id, blocks := range r.trackedBlocks {
				r.subscriptions[id].mu.Lock()
				delete(blocks, i)
				r.subscriptions[id].mu.Unlock()
			}
		}
		time.Sleep(waitPeriodBlockRemover)
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
				// TODO: handle error
				return
			}
			time.Sleep(waitPeriodBlockAdder)
		}
		block, err := r.ethClient.HeaderByNumber(ctx, big.NewInt(int64(currentBlock)))
		if err != nil {
			// TODO: handle error
			return
		}
		prevBlockHash, ok := r.unfinalisedBlocks[currentBlock-1]
		if !ok || block.ParentHash == prevBlockHash {
			// previous block is correct or there is no previous block, add current block
			r.mu.Lock()
			r.unfinalisedBlocks[currentBlock] = block.Hash()
			r.mu.Unlock()
			if firstBlockReorged > 0 {
				r.notifyReorg(currentBlock, firstBlockReorged)
			}
			currentBlock++
			firstBlockReorged = 0
		} else {
			// previous block is reorged:
			// 1. add a pending reorg to be processed for all the subscribers (so they don't add more blocks)
			// 2. remove block
			for _, sub := range r.subscriptions {
				sub.mu.Lock()
				sub.pendingReorgsToBeProcessed.Add(1)
				sub.mu.Unlock()
			}
			r.mu.Lock()
			delete(r.unfinalisedBlocks, currentBlock-1)
			r.mu.Unlock()
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
	blocks := r.trackedBlocks[id]
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
