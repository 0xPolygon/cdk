package sync

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

var ErrInconsistentState = errors.New("state is inconsistent, try again later once the state is consolidated")

type Block struct {
	Num    uint64
	Events []interface{}
	Hash   common.Hash
}

type ProcessorInterface interface {
	GetLastProcessedBlock(ctx context.Context) (uint64, error)
	ProcessBlock(block Block) error
	Reorg(firstReorgedBlock uint64) error
}
