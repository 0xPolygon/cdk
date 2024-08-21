package reorgdetector

import (
	"sort"

	common "github.com/ethereum/go-ethereum/common"
)

type block struct {
	Num        uint64
	Hash       common.Hash
	ParentHash common.Hash
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
