package dataavailability

import (
	"context"

	"github.com/0xPolygon/cdk/etherman/types"
)

// DataAvailability implements an abstract data availability integration
type DataAvailability struct {
	backend DABackender
}

// New creates a DataAvailability instance
func New(backend DABackender) (*DataAvailability, error) {
	da := &DataAvailability{
		backend: backend,
	}

	return da, da.backend.Init()
}

// PostSequence sends the sequence data to the data availability backend, and returns the dataAvailabilityMessage
// as expected by the contract
func (d *DataAvailability) PostSequence(ctx context.Context, sequences []types.Sequence) ([]byte, error) {
	batchesData := [][]byte{}
	for _, batch := range sequences {
		// Do not send to the DA backend data that will be stored to L1
		if batch.ForcedBatchTimestamp == 0 {
			batchesData = append(batchesData, batch.BatchL2Data)
		}
	}
	return d.backend.PostSequence(ctx, batchesData)
}
