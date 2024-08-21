package reorgdetector

import (
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type header struct {
	Num        uint64
	Hash       common.Hash
	ParentHash common.Hash
}

// newHeader returns a new instance of header
func newHeader(num uint64, hash, parentHash common.Hash) header {
	return header{
		Num:        num,
		Hash:       hash,
		ParentHash: parentHash,
	}
}

type headersList struct {
	sync.RWMutex
	headers map[uint64]header
}

// newHeadersList returns a new instance of headersList
func newHeadersList(headers ...header) *headersList {
	headersMap := make(map[uint64]header, len(headers))

	for _, b := range headers {
		headersMap[b.Num] = b
	}

	return &headersList{
		headers: headersMap,
	}
}

// len returns the number of headers in the headers list
func (hl *headersList) len() int {
	hl.RLock()
	ln := len(hl.headers)
	hl.RUnlock()
	return ln
}

// isEmpty returns true if the headers list is empty
func (hl *headersList) isEmpty() bool {
	return hl.len() == 0
}

// add adds a header to the headers list
func (hl *headersList) add(h header) {
	hl.Lock()
	hl.headers[h.Num] = h
	hl.Unlock()
}

// copy returns a copy of the headers list
func (hl *headersList) copy() *headersList {
	hl.RLock()
	defer hl.RUnlock()

	headersMap := make(map[uint64]header, len(hl.headers))
	for k, v := range hl.headers {
		headersMap[k] = v
	}

	return &headersList{
		headers: headersMap,
	}
}

// get returns a header by block number
func (hl *headersList) get(num uint64) *header {
	hl.RLock()
	defer hl.RUnlock()

	if b, ok := hl.headers[num]; ok {
		return &b
	}

	return nil
}

// getSorted returns headers in sorted order
func (hl *headersList) getSorted() []header {
	sortedBlocks := make([]header, 0, len(hl.headers))

	hl.RLock()
	for _, b := range hl.headers {
		sortedBlocks = append(sortedBlocks, b)
	}
	hl.RUnlock()

	sort.Slice(sortedBlocks, func(i, j int) bool {
		return sortedBlocks[i].Num < sortedBlocks[j].Num
	})

	return sortedBlocks
}

// getFromBlockSorted returns blocks from blockNum in sorted order without including the blockNum
func (hl *headersList) getFromBlockSorted(blockNum uint64) []header {
	sortedHeaders := hl.getSorted()

	index := -1
	for i, b := range sortedHeaders {
		if b.Num > blockNum {
			index = i
			break
		}
	}

	if index == -1 {
		return nil
	}

	return sortedHeaders[index:]
}

// getClosestHigherBlock returns the closest higher block to the given blockNum
func (hl *headersList) getClosestHigherBlock(blockNum uint64) (*header, bool) {
	hdr := hl.get(blockNum)
	if hdr != nil {
		return hdr, true
	}

	sorted := hl.getFromBlockSorted(blockNum)
	if len(sorted) == 0 {
		return nil, false
	}

	return &sorted[0], true
}

// removeRange removes headers from "from" to "to"
func (hl *headersList) removeRange(from, to uint64) {
	hl.Lock()
	for i := from; i <= to; i++ {
		delete(hl.headers, i)
	}
	hl.Unlock()
}

// detectReorg detects a reorg in the given headers list.
// Returns the first reorged headers or nil.
func (hl *headersList) detectReorg() *header {
	hl.RLock()
	defer hl.RUnlock()

	// Find the highest block number
	maxBlockNum := uint64(0)
	for blockNum := range hl.headers {
		if blockNum > maxBlockNum {
			maxBlockNum = blockNum
		}
	}

	// Iterate from the highest block number to the lowest
	reorgDetected := false
	for i := maxBlockNum; i > 1; i-- {
		currentBlock, currentExists := hl.headers[i]
		previousBlock, previousExists := hl.headers[i-1]

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

	// No reorg detected
	return nil
}
