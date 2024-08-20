package reorgdetector

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	types "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

type ReorgMonitor struct {
	lock sync.Mutex

	client EthClient

	maxBlocksInCache int
	lastBlockHeight  uint64

	knownReorgs       map[string]uint64 // key: reorgId, value: endBlockNumber
	subscriptionsLock sync.RWMutex
	subscriptions     map[string]*Subscription

	blockByHash    map[common.Hash]*Block
	blocksByHeight map[uint64]map[common.Hash]*Block

	EarliestBlockNumber uint64
	LatestBlockNumber   uint64
}

func NewReorgMonitor(client EthClient, maxBlocks int) *ReorgMonitor {
	return &ReorgMonitor{
		client:           client,
		maxBlocksInCache: maxBlocks,
		blockByHash:      make(map[common.Hash]*Block),
		blocksByHeight:   make(map[uint64]map[common.Hash]*Block),
		knownReorgs:      make(map[string]uint64),
		subscriptions:    make(map[string]*Subscription),
	}
}

func (mon *ReorgMonitor) Start(ctx context.Context) error {
	// Add head tracker
	ch := make(chan *types.Header, 100)
	sub, err := mon.client.SubscribeNewHead(ctx, ch)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-sub.Err():
				return
			case header := <-ch:
				if err = mon.onNewHeader(header); err != nil {
					log.Println(err)
					continue
				}
			}
		}
	}()

	return nil
}

func (mon *ReorgMonitor) Subscribe(id string) (*Subscription, error) {
	mon.subscriptionsLock.Lock()
	defer mon.subscriptionsLock.Unlock()

	if sub, ok := mon.subscriptions[id]; ok {
		return sub, nil
	}

	sub := &Subscription{
		FirstReorgedBlock: make(chan uint64),
		ReorgProcessed:    make(chan bool),
	}
	mon.subscriptions[id] = sub

	return sub, nil
}

func (mon *ReorgMonitor) AddBlockToTrack(ctx context.Context, id string, blockNum uint64, blockHash common.Hash) error {
	return nil
}

func (mon *ReorgMonitor) notifySubscribers(reorg *Reorg) {
	// TODO: Add wg group
	mon.subscriptionsLock.Lock()
	for _, sub := range mon.subscriptions {
		sub.FirstReorgedBlock <- reorg.StartBlockHeight
		<-sub.ReorgProcessed
	}
	mon.subscriptionsLock.Unlock()
}

func (mon *ReorgMonitor) onNewHeader(header *types.Header) error {
	mon.lock.Lock()
	defer mon.lock.Unlock()

	mon.addBlock(NewBlock(header, OriginSubscription))

	// Do nothing if block is at previous height
	if header.Number.Uint64() == mon.lastBlockHeight {
		return nil
	}

	if len(mon.blocksByHeight) < 3 {
		return nil
	}

	// Analyze blocks once a new height has been reached
	mon.lastBlockHeight = header.Number.Uint64()

	anal, err := mon.AnalyzeTree(0, 2)
	if err != nil {
		return err
	}

	for _, reorg := range anal.Reorgs {
		if !reorg.IsFinished { // don't care about unfinished reorgs
			continue
		}

		// Send new finished reorgs to channel
		if _, isKnownReorg := mon.knownReorgs[reorg.Id()]; !isKnownReorg {
			mon.knownReorgs[reorg.Id()] = reorg.EndBlockHeight
			go mon.notifySubscribers(reorg)
		}
	}

	return nil
}

// addBlock adds a block to history if it hasn't been seen before, and download unknown referenced blocks (parent, uncles).
func (mon *ReorgMonitor) addBlock(block *Block) bool {
	defer mon.trimCache()

	// If known, then only overwrite if known was by uncle
	_, isKnown := mon.blockByHash[block.Header.Hash()]
	if isKnown {
		return false
	}

	// Only accept blocks that are after the earliest known (some nodes might be further back)
	if block.Number() < mon.EarliestBlockNumber {
		return false
	}

	// Add for access by hash
	mon.blockByHash[block.Header.Hash()] = block

	// Create array of blocks at this height, if necessary
	if _, found := mon.blocksByHeight[block.Number()]; !found {
		mon.blocksByHeight[block.Number()] = make(map[common.Hash]*Block)
	}

	// Add to map of blocks at this height
	mon.blocksByHeight[block.Number()][block.Header.Hash()] = block

	// Set earliest block
	if mon.EarliestBlockNumber == 0 || block.Number() < mon.EarliestBlockNumber {
		mon.EarliestBlockNumber = block.Number()
	}

	// Set latest block
	if block.Number() > mon.LatestBlockNumber {
		mon.LatestBlockNumber = block.Number()
	}

	// Check if further blocks can be downloaded from this one
	if block.Number() > mon.EarliestBlockNumber { // check backhistory only if we are past the earliest block
		err := mon.checkBlockForReferences(block)
		if err != nil {
			log.Println(err)
		}
	}

	return true
}

func (mon *ReorgMonitor) trimCache() {
	// Trim reorg history
	for reorgId, reorgEndBlockheight := range mon.knownReorgs {
		if reorgEndBlockheight < mon.EarliestBlockNumber {
			delete(mon.knownReorgs, reorgId)
		}
	}

	for currentHeight := mon.EarliestBlockNumber; currentHeight < mon.LatestBlockNumber; currentHeight++ {
		blocks, heightExists := mon.blocksByHeight[currentHeight]
		if !heightExists {
			continue
		}

		// Set new lowest block number
		mon.EarliestBlockNumber = currentHeight

		// Stop if trimmed enough
		if len(mon.blockByHash) <= mon.maxBlocksInCache {
			return
		}

		// Trim
		for hash := range blocks {
			delete(mon.blocksByHeight[currentHeight], hash)
			delete(mon.blockByHash, hash)
		}
		delete(mon.blocksByHeight, currentHeight)
	}
}

func (mon *ReorgMonitor) checkBlockForReferences(block *Block) error {
	// Check parent
	_, found := mon.blockByHash[block.Header.ParentHash]
	if !found {
		// fmt.Printf("- parent of %d %s not found (%s), downloading...\n", block.Number, block.Hash, block.ParentHash)
		_, _, err := mon.ensureBlock(block.Header.ParentHash, OriginGetParent)
		if err != nil {
			return errors.Wrap(err, "get-parent error")
		}
	}

	// Check uncles
	/*for _, uncleHeader := range block.Block.Uncles() {
		// fmt.Printf("- block %d %s has uncle: %s\n", block.Number, block.Hash, uncleHeader.Hash())
		_, _, err := mon.ensureBlock(uncleHeader.Hash(), OriginUncle)
		if err != nil {
			return errors.Wrap(err, "get-uncle error")
		}
	}*/

	// ro.DebugPrintln(fmt.Sprintf("- added block %d %s", block.NumberU64(), block.Hash()))
	return nil
}

func (mon *ReorgMonitor) ensureBlock(blockHash common.Hash, origin BlockOrigin) (block *Block, alreadyExisted bool, err error) {
	// Check and potentially download block
	var found bool
	block, found = mon.blockByHash[blockHash]
	if found {
		return block, true, nil
	}

	fmt.Printf("- block %s (%s) not found, downloading from...\n", blockHash, origin)
	ethBlock, err := mon.client.HeaderByHash(context.Background(), blockHash)
	if err != nil {
		fmt.Println("- err block not found:", blockHash, err) // todo: try other clients
		msg := fmt.Sprintf("EnsureBlock error for hash %s", blockHash)
		return nil, false, errors.Wrap(err, msg)
	}

	block = NewBlock(ethBlock, origin)

	// Add a new block without sending to channel, because that makes reorg.AddBlock() asynchronous,
	// but we want reorg.AddBlock() to wait until all references are added.
	mon.addBlock(block)

	return block, false, nil
}

func (mon *ReorgMonitor) AnalyzeTree(maxBlocks, distanceToLastBlockHeight uint64) (*TreeAnalysis, error) {
	// Set end height of search
	endBlockNumber := mon.LatestBlockNumber - distanceToLastBlockHeight

	// Set start height of search
	startBlockNumber := mon.EarliestBlockNumber
	if maxBlocks > 0 && endBlockNumber-maxBlocks > mon.EarliestBlockNumber {
		startBlockNumber = endBlockNumber - maxBlocks
	}

	// Build tree datastructure
	tree := NewBlockTree()
	for height := startBlockNumber; height <= endBlockNumber; height++ {
		numBlocksAtHeight := len(mon.blocksByHeight[height])
		if numBlocksAtHeight == 0 {
			err := fmt.Errorf("error in monitor.AnalyzeTree: no blocks at height %d", height)
			return nil, err
		}

		// Start tree only when 1 block at this height. If more blocks then skip.
		if tree.FirstNode == nil && numBlocksAtHeight > 1 {
			continue
		}

		// Add all blocks at this height to the tree
		for _, currentBlock := range mon.blocksByHeight[height] {
			err := tree.AddBlock(currentBlock)
			if err != nil {
				return nil, errors.Wrap(err, "monitor.AnalyzeTree->tree.AddBlock error")
			}
		}
	}

	// Get analysis of tree
	anal, err := NewTreeAnalysis(tree)
	if err != nil {
		return nil, errors.Wrap(err, "monitor.AnalyzeTree->NewTreeAnalysis error")
	}

	return anal, nil
}

func (mon *ReorgMonitor) String() string {
	return fmt.Sprintf("ReorgMonitor: %d - %d, %d / %d blocks, %d reorgcache", mon.EarliestBlockNumber, mon.LatestBlockNumber, len(mon.blockByHash), len(mon.blocksByHeight), len(mon.knownReorgs))
}
