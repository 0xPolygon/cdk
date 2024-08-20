package seqsendertypes

import (
	"github.com/ethereum/go-ethereum/common"
)

type Batch interface {
	//underlyingType *ethmantypes.Batch
	DeepCopy() Batch
	LastCoinbase() common.Address
	ForcedBatchTimestamp() uint64
	ForcedGlobalExitRoot() common.Hash
	ForcedBlockHashL1() common.Hash
	L2Data() []byte
	LastL2BLockTimestamp() uint64
	BatchNumber() uint64
	GlobalExitRoot() common.Hash
	L1InfoTreeIndex() uint32

	String() string

	// WRITE
	SetL2Data(data []byte)
	SetLastCoinbase(address common.Address)
	SetLastL2BLockTimestamp(ts uint64)
	SetL1InfoTreeIndex(index uint32)
}

type Sequence interface {
	IndexL1InfoRoot() uint32
	MaxSequenceTimestamp() uint64
	L1InfoRoot() common.Hash
	Batches() []Batch
	FirstBatch() Batch
	LastBatch() Batch
	Len() int
	L2Coinbase() common.Address
	LastVirtualBatchNumber() uint64

	String() string
	// WRITE
	SetLastVirtualBatchNumber(batchNumber uint64)
	//SetL1InfoRoot(hash common.Hash)
	//SetOldAccInputHash(hash common.Hash)
	//SetAccInputHash(hash common.Hash)
}
