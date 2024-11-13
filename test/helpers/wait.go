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
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		lpb, err := processor.GetLastProcessedBlock(ctx)
		require.NoError(t, err)
		if targetBlock <= lpb {
			return
		}
		time.Sleep(time.Millisecond * 10)
	}
	require.NoError(t, errors.New("processor not updated"))
}
