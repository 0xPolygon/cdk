package sync

import (
	"context"
	"errors"
)

var ErrInconsistentState = errors.New("state is inconsistent, try again later once the state is consolidated")

type Block struct {
	Num    uint64
	Events []interface{}
}

type ProcessorInterface interface {
	GetLastProcessedBlock(ctx context.Context) (uint64, error)
	ProcessBlock(block Block) error
	Reorg(firstReorgedBlock uint64) error
}
