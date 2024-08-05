package sync

import "github.com/ethereum/go-ethereum/common"

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
