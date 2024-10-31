package types

import "github.com/0xPolygon/cdk/etherman"

type EventNewBlock struct {
	BlockNumber       uint64
	BlockFinalityType etherman.BlockNumberFinality
}

type BlockNotifier interface {
	// NotifyEpochStarted notifies the epoch has started.
	Subscribe(id string) <-chan EventNewBlock
	String() string
}
