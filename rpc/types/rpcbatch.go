package types

import (
	"fmt"

	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/ethereum/go-ethereum/common"
)

type RPCBatch struct {
	batchNumber          uint64
	accInputHash         common.Hash
	blockHashes          []string
	batchL2Data          []byte
	globalExitRoot       common.Hash
	localExitRoot        common.Hash
	stateRoot            common.Hash
	coinbase             common.Address
	closed               bool
	lastL2BlockTimestamp uint64
	l1InfoTreeIndex      uint32
}

func NewRPCBatch(batchNumber uint64, accInputHash common.Hash, blockHashes []string, batchL2Data []byte,
	globalExitRoot common.Hash, localExitRoot common.Hash, stateRoot common.Hash,
	coinbase common.Address, closed bool) *RPCBatch {
	return &RPCBatch{
		batchNumber:    batchNumber,
		accInputHash:   accInputHash,
		blockHashes:    blockHashes,
		batchL2Data:    batchL2Data,
		globalExitRoot: globalExitRoot,
		localExitRoot:  localExitRoot,
		stateRoot:      stateRoot,
		coinbase:       coinbase,
		closed:         closed,
	}
}

// DeepCopy
func (b *RPCBatch) DeepCopy() seqsendertypes.Batch {
	return &RPCBatch{
		accInputHash:         b.accInputHash,
		batchNumber:          b.batchNumber,
		blockHashes:          b.blockHashes,
		batchL2Data:          b.batchL2Data,
		globalExitRoot:       b.globalExitRoot,
		localExitRoot:        b.localExitRoot,
		stateRoot:            b.stateRoot,
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

// LocalExitRoot
func (b *RPCBatch) LocalExitRoot() common.Hash {
	return b.localExitRoot
}

// StateRoot
func (b *RPCBatch) StateRoot() common.Hash {
	return b.stateRoot
}

// AccInputHash
func (b *RPCBatch) AccInputHash() common.Hash {
	return b.accInputHash
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
	return fmt.Sprintf(
		"Batch/RPC: LastCoinbase: %s, ForcedBatchTimestamp: %d, ForcedGlobalExitRoot: %x, ForcedBlockHashL1: %x"+
			", L2Data: %x, LastL2BLockTimestamp: %d, BatchNumber: %d, GlobalExitRoot: %x, L1InfoTreeIndex: %d",
		b.LastCoinbase().String(),
		b.ForcedBatchTimestamp(),
		b.ForcedGlobalExitRoot().String(),
		b.ForcedBlockHashL1().String(),
		b.L2Data(),
		b.LastL2BLockTimestamp(),
		b.BatchNumber(),
		b.GlobalExitRoot().String(),
		b.L1InfoTreeIndex(),
	)
}

// IsClosed
func (b *RPCBatch) IsClosed() bool {
	return b.closed
}
