package reorgmonitor

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

type EthClient interface {
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	BlockNumber(ctx context.Context) (uint64, error)
}

type ReorgMonitor struct {
	lock sync.Mutex

	client EthClient

	maxBlocksInCache int
	lastBlockHeight  uint64

	newReorgChan chan<- *Reorg
	knownReorgs  map[string]uint64 // key: reorgId, value: endBlockNumber

	blockByHash    map[common.Hash]*Block
	blocksByHeight map[uint64]map[common.Hash]*Block

	EarliestBlockNumber uint64
	LatestBlockNumber   uint64
}

func NewReorgMonitor(client EthClient, reorgChan chan<- *Reorg, maxBlocks int) *ReorgMonitor {
	return &ReorgMonitor{
		client:           client,
		maxBlocksInCache: maxBlocks,
		blockByHash:      make(map[common.Hash]*Block),
		blocksByHeight:   make(map[uint64]map[common.Hash]*Block),
		newReorgChan:     reorgChan,
		knownReorgs:      make(map[string]uint64),
	}
}

// AddBlockToTrack adds a block to the monitor, and sends it to the monitor's channel.
// After adding a new block, a reorg check takes place. If a new completed reorg is detected, it is sent to the channel.
func (mon *ReorgMonitor) AddBlockToTrack(block *Block) error {
	mon.lock.Lock()
	defer mon.lock.Unlock()

	mon.addBlock(block)

	// Do nothing if block is at previous height
	if block.Number == mon.lastBlockHeight {
		return nil
	}

	if len(mon.blocksByHeight) < 3 {
		return nil
	}

	// Analyze blocks once a new height has been reached
	mon.lastBlockHeight = block.Number

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
			mon.newReorgChan <- reorg
		}
	}

	return nil
}

// addBlock adds a block to history if it hasn't been seen before, and download unknown referenced blocks (parent, uncles).
func (mon *ReorgMonitor) addBlock(block *Block) bool {
	defer mon.trimCache()

	// If known, then only overwrite if known was by uncle
	knownBlock, isKnown := mon.blockByHash[block.Hash]
	if isKnown && knownBlock.Origin != OriginUncle {
		return false
	}

	// Only accept blocks that are after the earliest known (some nodes might be further back)
	if block.Number < mon.EarliestBlockNumber {
		return false
	}

	// Print
	blockInfo := fmt.Sprintf("Add%s \t %-12s \t %s", block.String(), block.Origin, mon)
	log.Println(blockInfo)

	// Add for access by hash
	mon.blockByHash[block.Hash] = block

	// Create array of blocks at this height, if necessary
	if _, found := mon.blocksByHeight[block.Number]; !found {
		mon.blocksByHeight[block.Number] = make(map[common.Hash]*Block)
	}

	// Add to map of blocks at this height
	mon.blocksByHeight[block.Number][block.Hash] = block

	// Set earliest block
	if mon.EarliestBlockNumber == 0 || block.Number < mon.EarliestBlockNumber {
		mon.EarliestBlockNumber = block.Number
	}

	// Set latest block
	if block.Number > mon.LatestBlockNumber {
		mon.LatestBlockNumber = block.Number
	}

	// Check if further blocks can be downloaded from this one
	if block.Number > mon.EarliestBlockNumber { // check backhistory only if we are past the earliest block
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
	_, found := mon.blockByHash[block.ParentHash]
	if !found {
		// fmt.Printf("- parent of %d %s not found (%s), downloading...\n", block.Number, block.Hash, block.ParentHash)
		_, _, err := mon.ensureBlock(block.ParentHash, OriginGetParent)
		if err != nil {
			return errors.Wrap(err, "get-parent error")
		}
	}

	// Check uncles
	for _, uncleHeader := range block.Block.Uncles() {
		// fmt.Printf("- block %d %s has uncle: %s\n", block.Number, block.Hash, uncleHeader.Hash())
		_, _, err := mon.ensureBlock(uncleHeader.Hash(), OriginUncle)
		if err != nil {
			return errors.Wrap(err, "get-uncle error")
		}
	}

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
	ethBlock, err := mon.client.BlockByHash(context.Background(), blockHash)
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
