package state

import (
	"time"
)

// ZKCounters counters for the tx
type ZKCounters struct {
	GasUsed              uint64
	UsedKeccakHashes     uint32
	UsedPoseidonHashes   uint32
	UsedPoseidonPaddings uint32
	UsedMemAligns        uint32
	UsedArithmetics      uint32
	UsedBinaries         uint32
	UsedSteps            uint32
	UsedSha256Hashes_V2  uint32 //nolint:stylecheck
}

// BatchResources is a struct that contains the ZKEVM resources used by a batch/tx
type BatchResources struct {
	ZKCounters ZKCounters
	Bytes      uint64
}

// Proof struct
type Proof struct {
	BatchNumber      uint64
	BatchNumberFinal uint64
	Proof            string
	InputProver      string
	ProofID          *string
	// Prover name, unique identifier across prover reboots.
	Prover *string
	// ProverID prover process identifier.
	ProverID *string
	// GeneratingSince holds the timestamp for the moment in which the
	// proof generation has started by a prover. Nil if the proof is not
	// currently generating.
	GeneratingSince *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Sequence represents the sequence interval
type Sequence struct {
	FromBatchNumber uint64
	ToBatchNumber   uint64
}

// DBBatch struct is a wrapper for the state.Batch and its metadata
type DBBatch struct {
	Batch      Batch
	Datastream []byte
	Witness    []byte
}
