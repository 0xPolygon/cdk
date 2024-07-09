package dataavailability

import (
	"context"

	ethmanTypes "github.com/0xPolygon/cdk/aggregator/ethmantypes"
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

func (d *DataAvailability) PostSequence(ctx context.Context, sequenceBanana ethmanTypes.SequenceBanana) ([]byte, error) {
	return d.backend.PostSequence(ctx, sequenceBanana)
}
