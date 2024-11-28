package types

import (
	"time"

	"github.com/0xPolygon/cdk/etherman"
)

type EventNewBlock struct {
	BlockNumber       uint64
	BlockFinalityType etherman.BlockNumberFinality
	BlockRate         time.Duration
}

// BlockNotifier is the interface that wraps the basic methods to notify a new block.
type BlockNotifier interface {
	// NotifyEpochStarted notifies the epoch has started.
	Subscribe(id string) <-chan EventNewBlock
	String() string
}
