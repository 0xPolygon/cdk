package state

import (
	"time"

	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Batch struct
type Batch struct {
	BatchNumber     uint64
	Coinbase        common.Address
	BatchL2Data     []byte
	StateRoot       common.Hash
	LocalExitRoot   common.Hash
	AccInputHash    common.Hash
	L1InfoTreeIndex uint32
	L1InfoRoot      common.Hash
	// Timestamp (<=incaberry) -> batch time
	// 			 (>incaberry) -> minTimestamp used in batch creation, real timestamp is in virtual_batch.batch_timestamp
	Timestamp      time.Time
	Transactions   []types.Transaction
	GlobalExitRoot common.Hash
	ForcedBatchNum *uint64
	Resources      BatchResources
	// WIP: if WIP == true is a openBatch
	WIP     bool
	ChainID uint64
	ForkID  uint64
	Type    datastream.BatchType
}
