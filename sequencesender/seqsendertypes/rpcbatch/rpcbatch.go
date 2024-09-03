package rpcbatch

import (
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/ethereum/go-ethereum/common"
)

type RPCBatch struct {
	batchNumber          uint64         `mapstructure:"batchNumber"`
	blockHashes          []string       `mapstructure:"blocks"`
	batchL2Data          []byte         `mapstructure:"batchL2Data"`
	globalExitRoot       common.Hash    `mapstructure:"globalExitRoot"`
	coinbase             common.Address `mapstructure:"coinbase"`
	closed               bool           `mapstructure:"closed"`
	lastL2BlockTimestamp uint64         `mapstructure:"lastL2BlockTimestamp"`
	l1InfoTreeIndex      uint32         `mapstructure:"l1InfoTreeIndex"`
}

func New(batchNumber uint64, blockHashes []string, batchL2Data []byte, globalExitRoot common.Hash, coinbase common.Address, closed bool) (*RPCBatch, error) {
	return &RPCBatch{
		batchNumber:    batchNumber,
		blockHashes:    blockHashes,
		batchL2Data:    batchL2Data,
		globalExitRoot: globalExitRoot,
		coinbase:       coinbase,
		closed:         closed,
	}, nil
}

// DeepCopy
func (b *RPCBatch) DeepCopy() seqsendertypes.Batch {
	return &RPCBatch{
		batchNumber:          b.batchNumber,
		blockHashes:          b.blockHashes,
		batchL2Data:          b.batchL2Data,
		globalExitRoot:       b.globalExitRoot,
		coinbase:             b.coinbase,
		closed:               b.closed,
		lastL2BlockTimestamp: b.lastL2BlockTimestamp,
		l1InfoTreeIndex:      b.l1InfoTreeIndex,
	}
}

// LastCoinbase
func (b *RPCBatch) LastCoinbase() common.Address {
	return b.coinbase
}

// ForcedBatchTimestamp
func (b *RPCBatch) ForcedBatchTimestamp() uint64 {
	return 0
}

// ForcedGlobalExitRoot
func (b *RPCBatch) ForcedGlobalExitRoot() common.Hash {
	return common.Hash{}
}

// ForcedBlockHashL1
func (b *RPCBatch) ForcedBlockHashL1() common.Hash {
	return common.Hash{}
}

// L2Data
func (b *RPCBatch) L2Data() []byte {
	return b.batchL2Data
}

// LastL2BLockTimestamp
func (b *RPCBatch) LastL2BLockTimestamp() uint64 {
	return b.lastL2BlockTimestamp
}

// BatchNumber
func (b *RPCBatch) BatchNumber() uint64 {
	return b.batchNumber
}

// GlobalExitRoot
func (b *RPCBatch) GlobalExitRoot() common.Hash {
	return b.globalExitRoot
}

// L1InfoTreeIndex
func (b *RPCBatch) L1InfoTreeIndex() uint32 {
	return b.l1InfoTreeIndex
}

// SetL2Data
func (b *RPCBatch) SetL2Data(data []byte) {
	b.batchL2Data = data
}

// SetLastCoinbase
func (b *RPCBatch) SetLastCoinbase(address common.Address) {
	b.coinbase = address
}

// SetLastL2BLockTimestamp
func (b *RPCBatch) SetLastL2BLockTimestamp(ts uint64) {
	b.lastL2BlockTimestamp = ts
}

// SetL1InfoTreeIndex
func (b *RPCBatch) SetL1InfoTreeIndex(index uint32) {
	b.l1InfoTreeIndex = index
}

// String
func (b *RPCBatch) String() string {
	return "RPCBatch"
}

// IsClosed
func (b *RPCBatch) IsClosed() bool {
	return b.closed
}
