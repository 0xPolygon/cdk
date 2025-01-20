package sync

import "github.com/ethereum/go-ethereum/common"

type EVMBlocks []EVMBlock

func (e EVMBlocks) Len() int {
	return len(e)
}

func (e EVMBlocks) LastFinalizedBlock(lastFinalizedBlockOnNetwork uint64) (uint64, int, bool) {
	for i := len(e) - 1; i >= 0; i-- {
		if e[i].Num <= lastFinalizedBlockOnNetwork {
			return e[i].Num, i, true
		}
	}

	return 0, 0, false // no finalized block found
}

type EVMBlock struct {
	EVMBlockHeader
	Events []interface{}
}

type EVMBlockHeader struct {
	Num        uint64
	Hash       common.Hash
	ParentHash common.Hash
	Timestamp  uint64
}
