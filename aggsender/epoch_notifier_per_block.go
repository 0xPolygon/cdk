package aggsender

import (
	"context"
	"fmt"

	"github.com/0xPolygon/cdk/agglayer"
	"github.com/0xPolygon/cdk/aggsender/types"
)

type ExtraInfoEventEpoch struct {
	PendingBlocks int
}

func (e *ExtraInfoEventEpoch) String() string {
	return fmt.Sprintf("ExtraInfoEventEpoch: pendingBlocks=%d", e.PendingBlocks)
}

type ConfigEpochNotifierPerBlock struct {
	StartingEpochBlock uint64
	NumBlockPerEpoch   uint

	// Num blocks before the end of the epoch to notify it
	NotifyPendingBlocksBeforeEndEpoch uint
}

func NewConfigEpochNotifierPerBlock(aggLayer agglayer.AggLayerClientGetEpochConfiguration,
	notifyPendingBlocksBeforeEndEpoch uint) (*ConfigEpochNotifierPerBlock, error) {
	if aggLayer == nil {
		return nil, fmt.Errorf("newConfigEpochNotifierPerBlock: aggLayerClient is required")
	}
	clockConfig, err := aggLayer.GetEpochConfiguration()
	if err != nil {
		return nil, fmt.Errorf("newConfigEpochNotifierPerBlock: error getting clock configuration from AggLayer: %w", err)
	}
	return &ConfigEpochNotifierPerBlock{
		StartingEpochBlock:                clockConfig.GenesisBlock,
		NumBlockPerEpoch:                  uint(clockConfig.EpochDuration),
		NotifyPendingBlocksBeforeEndEpoch: notifyPendingBlocksBeforeEndEpoch,
	}, nil
}

func (c *ConfigEpochNotifierPerBlock) Validate() error {
	if c.NumBlockPerEpoch == 0 {
		return fmt.Errorf("numBlockPerEpoch: num block per epoch is required > 0 ")
	}
	if c.NumBlockPerEpoch-c.NotifyPendingBlocksBeforeEndEpoch == 0 {
		return fmt.Errorf("notifyPendingBlocksBeforeEndEpoch: Notify before num blocks end of epoch  is required > 0")
	}
	return nil
}

type EpochNotifierPerBlock struct {
	blockNotifier types.BlockNotifier
	logger        types.Logger

	lastStartingEpochBlock uint64

	Config ConfigEpochNotifierPerBlock
	types.GenericSubscriber[types.EpochEvent]
}

func NewEpochNotifierPerBlock(blockNotifier types.BlockNotifier,
	logger types.Logger,
	config ConfigEpochNotifierPerBlock,
	subscriber types.GenericSubscriber[types.EpochEvent]) (*EpochNotifierPerBlock, error) {
	if subscriber == nil {
		subscriber = NewGenericSubscriberImpl[types.EpochEvent]()
	}

	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &EpochNotifierPerBlock{
		blockNotifier:          blockNotifier,
		logger:                 logger,
		lastStartingEpochBlock: config.StartingEpochBlock,
		Config:                 config,
		GenericSubscriber:      subscriber,
	}, nil
}

func (e *EpochNotifierPerBlock) String() string {
	return fmt.Sprintf("EpochNotifierPerBlock: startingEpochBlock=%d, numBlockPerEpoch=%d,"+
		" notifyPendingBlocksBeforeEndEpoch=%d",
		e.Config.StartingEpochBlock, e.Config.NumBlockPerEpoch, e.Config.NotifyPendingBlocksBeforeEndEpoch)
}

// StartAsync starts the notifier in a goroutine
func (e *EpochNotifierPerBlock) StartAsync(ctx context.Context) {
	eventNewBlockChannel := e.blockNotifier.Subscribe("EpochNotifierPerBlock")
	go e.startInternal(ctx, eventNewBlockChannel)
}

// Start starts the notifier synchronously
func (e *EpochNotifierPerBlock) Start(ctx context.Context) {
	eventNewBlockChannel := e.blockNotifier.Subscribe("EpochNotifierPerBlock")
	e.startInternal(ctx, eventNewBlockChannel)
}

func (e *EpochNotifierPerBlock) startInternal(ctx context.Context, eventNewBlockChannel <-chan types.EventNewBlock) {
	status := internalStatus{
		lastBlockSeen:   e.Config.StartingEpochBlock,
		waitingForEpoch: e.epochNumber(e.Config.StartingEpochBlock),
	}
	for {
		select {
		case <-ctx.Done():
			return
		case newBlock := <-eventNewBlockChannel:
			var event *types.EpochEvent
			status, event = e.step(status, newBlock)
			if event != nil {
				e.logger.Debugf("new Epoch Event: %s", event.String())
				e.GenericSubscriber.Publish(*event)
			}
		}
	}
}

type internalStatus struct {
	lastBlockSeen   uint64
	waitingForEpoch uint64
}

func (e *EpochNotifierPerBlock) step(status internalStatus,
	newBlock types.EventNewBlock) (internalStatus, *types.EpochEvent) {
	currentBlock := newBlock.BlockNumber
	if currentBlock < e.Config.StartingEpochBlock {
		// This is a bit strange, the first epoch is in the future
		e.logger.Warnf("Block number %d is before the starting first epoch block %d."+
			" Please check your config", currentBlock, e.Config.StartingEpochBlock)
		return status, nil
	}
	// No new block
	if currentBlock <= status.lastBlockSeen {
		return status, nil
	}
	status.lastBlockSeen = currentBlock

	needNotify, closingEpoch := e.isClosingEpochBlock(currentBlock, status.waitingForEpoch)
	if needNotify {
		// Notify the epoch has started
		info := e.infoEpoch(currentBlock, closingEpoch)
		if status.waitingForEpoch == 0 && info.PendingBlocks > int(e.Config.NotifyPendingBlocksBeforeEndEpoch) {
			// We are in the first epoch, but we are not near the end of the epoch
			return status, nil
		}
		status.waitingForEpoch = closingEpoch + 1
		return status, &types.EpochEvent{
			Epoch:     closingEpoch,
			ExtraInfo: info,
		}
	}
	return status, nil
}

func (e *EpochNotifierPerBlock) infoEpoch(currentBlock, newEpochNotified uint64) *ExtraInfoEventEpoch {
	nextBlockStartingEpoch := e.endBlockEpoch(newEpochNotified)
	return &ExtraInfoEventEpoch{
		PendingBlocks: int(nextBlockStartingEpoch - currentBlock),
	}
}
func (e *EpochNotifierPerBlock) isClosingEpochBlock(currentBlock, lastEpochNotified uint64) (bool, uint64) {
	nextEpoch := e.epochNumber(currentBlock + uint64(e.Config.NotifyPendingBlocksBeforeEndEpoch))
	return nextEpoch > lastEpochNotified, e.epochNumber(currentBlock)
}
func (e *EpochNotifierPerBlock) startingBlockEpoch(epoch uint64) uint64 {
	if epoch == 0 {
		return e.Config.StartingEpochBlock - 1
	}
	return e.Config.StartingEpochBlock + ((epoch - 1) * uint64(e.Config.NumBlockPerEpoch))
}

func (e *EpochNotifierPerBlock) endBlockEpoch(epoch uint64) uint64 {
	return e.startingBlockEpoch(epoch + 1)
}
func (e *EpochNotifierPerBlock) epochNumber(currentBlock uint64) uint64 {
	if currentBlock < e.Config.StartingEpochBlock {
		return 0
	}
	return 1 + ((currentBlock - e.Config.StartingEpochBlock) / uint64(e.Config.NumBlockPerEpoch))
}
