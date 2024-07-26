package dataavailability

import (
	"context"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/ethereum/go-ethereum/common"
)

// DABackender is an interface for components that store and retrieve batch data
type DABackender interface {
	SequenceRetriever
	SequenceSender
	// Init initializes the DABackend
	Init() error
}

// SequenceSender is used to send provided sequence of batches
type SequenceSender interface {
	// PostSequence sends the sequence data to the data availability backend, and returns the dataAvailabilityMessage
	// as expected by the contract
	PostSequence(ctx context.Context, sequence etherman.SequenceBanana) ([]byte, error)
	// PostSequenceElderberry sends the sequence data to the data availability backend, and returns the dataAvailabilityMessage
	// as expected by the contract
	PostSequenceElderberry(ctx context.Context, batchesData [][]byte) ([]byte, error)
}

type SequenceSenderElderberry interface {
	// PostSequence sends the sequence data to the data availability backend, and returns the dataAvailabilityMessage
	// as expected by the contract
	PostSequenceElderberry(ctx context.Context, batchesData [][]byte) ([]byte, error)
}

// SequenceRetriever is used to retrieve batch data
type SequenceRetriever interface {
	// GetSequence retrieves the sequence data from the data availability backend
	GetSequence(ctx context.Context, batchHashes []common.Hash, dataAvailabilityMessage []byte) ([][]byte, error)
}

// BatchDataProvider is used to retrieve batch data
type BatchDataProvider interface {
	// GetBatchL2Data retrieve the data of a batch from the DA backend. The returned data must be the pre-image of the hash
	GetBatchL2Data(batchNum []uint64, batchHashes []common.Hash, dataAvailabilityMessage []byte) ([][]byte, error)
}

// DataManager is an interface for components that send and retrieve batch data
type DataManager interface {
	BatchDataProvider
	SequenceSender
}
