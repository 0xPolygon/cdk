package executionlayer

import "context"

type ExecutionLayer interface {
	// Start starts the sequencer
	Start(ctx context.Context) error
}
