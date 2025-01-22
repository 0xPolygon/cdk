package sync

import "github.com/ethereum/go-ethereum/common"

type EVMBlocks []EVMBlock

func (e EVMBlocks) Len() int {
	return len(e)
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
	SafeBlock  bool
}
