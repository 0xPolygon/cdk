package sync

import "context"

type Block struct {
	Num    uint64
	Events []interface{}
}

type ProcessorInterface interface {
	GetLastProcessedBlock(ctx context.Context) (uint64, error)
	ProcessBlock(block Block) error
	Reorg(firstReorgedBlock uint64) error
}
