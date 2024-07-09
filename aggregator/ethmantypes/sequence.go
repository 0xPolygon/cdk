package ethmantypes

import (
	"reflect"

	"github.com/ethereum/go-ethereum/common"
)

// Sequence represents an operation sent to the PoE smart contract to be
// processed.
type Sequence struct {
	GlobalExitRoot, StateRoot, LocalExitRoot common.Hash
	AccInputHash                             common.Hash
	LastL2BLockTimestamp                     uint64
	BatchL2Data                              []byte
	IsSequenceTooBig                         bool
	BatchNumber                              uint64
	ForcedBatchTimestamp                     int64
	PrevBlockHash                            common.Hash
	LastCoinbase                             common.Address
}

// IsEmpty checks is sequence struct is empty
func (s Sequence) IsEmpty() bool {
	return reflect.DeepEqual(s, Sequence{})
}

type Batch struct {
	L2Data               []byte
	LastCoinbase         common.Address
	ForcedGlobalExitRoot common.Hash
	ForcedBlockHashL1    common.Hash
	ForcedBatchTimestamp uint64
	BatchNumber          uint64
	L1InfoTreeIndex      uint32
	LastL2BLockTimestamp uint64
	GlobalExitRoot       common.Hash
}

type SequenceBanana struct {
	Batches              []Batch
	AccInputHash         common.Hash
	L1InfoRoot           common.Hash
	MaxSequenceTimestamp uint64
	IndexL1InfoRoot      uint32
	L2Coinbase           common.Address
}

func NewSequenceBanana(batches []Batch, l2Coinbase common.Address) *SequenceBanana {
	var (
		maxSequenceTimestamp uint64
		indexL1InfoRoot      uint32
	)

	for _, batch := range batches {
		if batch.LastL2BLockTimestamp > maxSequenceTimestamp {
			maxSequenceTimestamp = batch.LastL2BLockTimestamp
		}

		if batch.L1InfoTreeIndex > indexL1InfoRoot {
			indexL1InfoRoot = batch.L1InfoTreeIndex
		}
	}

	return &SequenceBanana{
		Batches:              batches,
		MaxSequenceTimestamp: maxSequenceTimestamp,
		IndexL1InfoRoot:      indexL1InfoRoot,
		L2Coinbase:           l2Coinbase,
	}
}

func (s *SequenceBanana) Len() int {
	return len(s.Batches)
}
