package etherman

import (
	"fmt"
	"math/big"
	"reflect"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	"github.com/ethereum/go-ethereum/common"
)

// Block struct
type Block struct {
	BlockNumber           uint64
	BlockHash             common.Hash
	ParentHash            common.Hash
	ForcedBatches         []ForcedBatch
	SequencedBatches      [][]SequencedBatch
	VerifiedBatches       []VerifiedBatch
	SequencedForceBatches [][]SequencedForceBatch
	ForkIDs               []ForkID
	ReceivedAt            time.Time
	// GER data
	GlobalExitRoots, L1InfoTree []GlobalExitRoot
}

// GlobalExitRoot struct
type GlobalExitRoot struct {
	BlockNumber       uint64
	MainnetExitRoot   common.Hash
	RollupExitRoot    common.Hash
	GlobalExitRoot    common.Hash
	Timestamp         time.Time
	PreviousBlockHash common.Hash
}

// PolygonZkEVMBatchData represents PolygonZkEVMBatchData
type PolygonZkEVMBatchData struct {
	Transactions       []byte
	GlobalExitRoot     [32]byte
	Timestamp          uint64
	MinForcedTimestamp uint64
}

// SequencedBatch represents virtual batch
type SequencedBatch struct {
	BatchNumber   uint64
	L1InfoRoot    *common.Hash
	SequencerAddr common.Address
	TxHash        common.Hash
	Nonce         uint64
	Coinbase      common.Address
	// Struct used in preEtrog forks
	*PolygonZkEVMBatchData
	// Struct used in Etrog
	*polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData
}

// ForcedBatch represents a ForcedBatch
type ForcedBatch struct {
	BlockNumber       uint64
	ForcedBatchNumber uint64
	Sequencer         common.Address
	GlobalExitRoot    common.Hash
	RawTxsData        []byte
	ForcedAt          time.Time
}

// VerifiedBatch represents a VerifiedBatch
type VerifiedBatch struct {
	BlockNumber uint64
	BatchNumber uint64
	Aggregator  common.Address
	StateRoot   common.Hash
	TxHash      common.Hash
}

// SequencedForceBatch is a sturct to track the ForceSequencedBatches event.
type SequencedForceBatch struct {
	BatchNumber uint64
	Coinbase    common.Address
	TxHash      common.Hash
	Timestamp   time.Time
	Nonce       uint64
	polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData
}

// ForkID is a sturct to track the ForkID event.
type ForkID struct {
	BatchNumber uint64
	ForkID      uint64
	Version     string
}

type BlockNumberFinality string

const (
	FinalizedBlock = BlockNumberFinality("FinalizedBlock")
	SafeBlock      = BlockNumberFinality("SafeBlock")
	PendingBlock   = BlockNumberFinality("PendingBlock")
	LatestBlock    = BlockNumberFinality("LatestBlock")
	EarliestBlock  = BlockNumberFinality("EarliestBlock")
)

func (b *BlockNumberFinality) ToBlockNum() (*big.Int, error) {
	switch *b {
	case FinalizedBlock:
		return big.NewInt(int64(Finalized)), nil
	case SafeBlock:
		return big.NewInt(int64(Safe)), nil
	case PendingBlock:
		return big.NewInt(int64(Pending)), nil
	case LatestBlock:
		return big.NewInt(int64(Latest)), nil
	case EarliestBlock:
		return big.NewInt(int64(Earliest)), nil
	default:
		return nil, fmt.Errorf("invalid finality keyword: %s", string(*b))
	}
}

type BlockNumber int64

const (
	Finalized = BlockNumber(-5)
	Safe      = BlockNumber(-4)
	Pending   = BlockNumber(-3)
	Latest    = BlockNumber(-2)
	Earliest  = BlockNumber(-1)
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
	Batches                []Batch
	OldAccInputHash        common.Hash
	AccInputHash           common.Hash
	L1InfoRoot             common.Hash
	MaxSequenceTimestamp   uint64
	CounterL1InfoRoot      uint32
	L2Coinbase             common.Address
	LastVirtualBatchNumber uint64
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
		CounterL1InfoRoot:    indexL1InfoRoot,
		L2Coinbase:           l2Coinbase,
	}
}

func (s *SequenceBanana) Len() int {
	return len(s.Batches)
}

func (s *SequenceBanana) SetLastVirtualBatchNumber(batchNumber uint64) {
	s.LastVirtualBatchNumber = batchNumber
}
