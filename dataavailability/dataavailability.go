package dataavailability

import (
	"context"

	"github.com/0xPolygon/cdk/etherman"
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

// PostSequenceBanana sends sequence data to the backend and returns a response.
func (d *DataAvailability) PostSequenceBanana(
	ctx context.Context, sequenceBanana etherman.SequenceBanana,
) ([]byte, error) {
	return d.backend.PostSequenceBanana(ctx, sequenceBanana)
}

// PostSequenceElderberry sends batch data to the backend and returns a response.
func (d *DataAvailability) PostSequenceElderberry(ctx context.Context, batchesData [][]byte) ([]byte, error) {
	return d.backend.PostSequenceElderberry(ctx, batchesData)
}
