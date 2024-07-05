package localbridgesync

import (
	"context"
	"errors"
)

type driver struct {
}

func newDriver() (*driver, error) {
	return nil, errors.New("not implemented")
}

func (lbs *LocalBridgeSync) Sync(ctx *context.Context) {
	// Download data
	// Process data
}
