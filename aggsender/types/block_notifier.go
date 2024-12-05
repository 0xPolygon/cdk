package types

import "github.com/0xPolygon/cdk/etherman"

type EventNewBlock struct {
	BlockNumber       uint64
	BlockFinalityType etherman.BlockNumberFinality
}

// BlockNotifier is the interface that wraps the basic methods to notify a new block.
type BlockNotifier interface {
	// NotifyEpochStarted notifies the epoch has started.
	Subscribe(id string) <-chan EventNewBlock
	String() string
}
