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

func (d *DataAvailability) PostSequence(ctx context.Context, sequenceBanana etherman.SequenceBanana) ([]byte, error) {
	return d.backend.PostSequence(ctx, sequenceBanana)
}

func (d *DataAvailability) PostSequenceElderberry(ctx context.Context, batchesData [][]byte) ([]byte, error) {
	return d.backend.PostSequenceElderberry(ctx, batchesData)
}
