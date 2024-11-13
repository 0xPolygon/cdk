package helpers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type Processorer interface {
	GetLastProcessedBlock(ctx context.Context) (uint64, error)
}

func RequireProcessorUpdated(t *testing.T, processor Processorer, targetBlock uint64) {
	t.Helper()
	const (
		maxIterations         = 100
		sleepTimePerIteration = time.Millisecond * 10
	)
	ctx := context.Background()
	for i := 0; i < maxIterations; i++ {
		lpb, err := processor.GetLastProcessedBlock(ctx)
		require.NoError(t, err)
		if targetBlock <= lpb {
			return
		}
		time.Sleep(sleepTimePerIteration)
	}
	require.NoError(t, errors.New("processor not updated"))
}
