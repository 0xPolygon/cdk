package aggsender

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/etherman"
)

var (
	timeNowFunc = time.Now
)

const (
	AutomaticBlockInterval = time.Second * 0
	// minBlockInterval is the minimum interval at which the AggSender will check for new blocks
	minBlockInterval = time.Second
	// maxBlockInterval is the maximum interval at which the AggSender will check for new blocks
	maxBlockInterval = time.Minute
)

type ConfigBlockNotifierPolling struct {
	// BlockFinalityType is the finality of the block to be notified
	BlockFinalityType etherman.BlockNumberFinality
	// CheckNewBlockInterval is the interval at which the AggSender will check for new blocks
	// if is 0 it will be calculated automatically
	CheckNewBlockInterval time.Duration
}

type BlockNotifierPolling struct {
	ethClient     types.EthClient
	blockFinality *big.Int
	logger        types.Logger
	config        ConfigBlockNotifierPolling
	mu            sync.Mutex
	lastStatus    *blockNotifierPollingInternalStatus
	types.GenericSubscriber[types.EventNewBlock]
}

// NewBlockNotifierPolling creates a new BlockNotifierPolling.
// if param `subscriber` is nil a new GenericSubscriberImpl[types.EventNewBlock] will be created.
// To use this class you need to subscribe and each time that a new block appear the subscriber
//
//	will be notified through the channel. (check unit tests TestExploratoryBlockNotifierPolling
//
// for more information)
func NewBlockNotifierPolling(ethClient types.EthClient,
	config ConfigBlockNotifierPolling,
	logger types.Logger,
	subscriber types.GenericSubscriber[types.EventNewBlock]) (*BlockNotifierPolling, error) {
	if subscriber == nil {
		subscriber = NewGenericSubscriberImpl[types.EventNewBlock]()
	}
	finality, err := config.BlockFinalityType.ToBlockNum()
	if err != nil {
		return nil, fmt.Errorf("failed to convert block finality type to block number: %w", err)
	}

	return &BlockNotifierPolling{
		ethClient:         ethClient,
		blockFinality:     finality,
		logger:            logger,
		config:            config,
		GenericSubscriber: subscriber,
	}, nil
}

func (b *BlockNotifierPolling) String() string {
	status := b.getGlobalStatus()
	res := fmt.Sprintf("BlockNotifierPolling: finality=%s", b.config.BlockFinalityType)
	if status != nil {
		res += fmt.Sprintf(" lastBlockSeen=%d", status.lastBlockSeen)
	} else {
		res += " lastBlockSeen=none"
	}
	return res
}

// StartAsync starts the BlockNotifierPolling in a new goroutine
func (b *BlockNotifierPolling) StartAsync(ctx context.Context) {
	go b.Start(ctx)
}

// Start starts the BlockNotifierPolling blocking the current goroutine
func (b *BlockNotifierPolling) Start(ctx context.Context) {
	ticker := time.NewTimer(b.config.CheckNewBlockInterval)
	defer ticker.Stop()

	var status *blockNotifierPollingInternalStatus = nil

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			delay, newStatus := b.step(ctx, status)
			status = newStatus
			b.setGlobalStatus(status)
			ticker.Reset(delay)
		}
	}
}

func (b *BlockNotifierPolling) setGlobalStatus(status *blockNotifierPollingInternalStatus) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lastStatus = status
}

func (b *BlockNotifierPolling) getGlobalStatus() *blockNotifierPollingInternalStatus {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.lastStatus == nil {
		return nil
	}
	copyStatus := *b.lastStatus
	return &copyStatus
}

func (b *BlockNotifierPolling) step(ctx context.Context,
	status *blockNotifierPollingInternalStatus) (time.Duration, *blockNotifierPollingInternalStatus) {
	currentBlock, err := b.ethClient.HeaderByNumber(ctx, b.blockFinality)
	if err != nil || currentBlock == nil {
		b.logger.Errorf("Failed to get block number: %v", err)
		return b.nextBlockRequestDelay(status, err), status.clear()
	}
	if status == nil {
		status = status.intialBlock(currentBlock.Number.Uint64())
		return b.nextBlockRequestDelay(status, nil), status
	}

	if currentBlock.Number.Uint64() != status.lastBlockSeen {
		b.Publish(types.EventNewBlock{
			BlockNumber:       currentBlock.Number.Uint64(),
			BlockFinalityType: b.config.BlockFinalityType,
		})
		now := timeNowFunc()
		timePreviousBlock := now.Sub(status.lastBlockTime)
		status.previousBlockTime = &timePreviousBlock
		status.lastBlockTime = now

		if currentBlock.Number.Uint64()-status.lastBlockSeen != 1 {
			b.logger.Warnf("Missed block(s) [finality:%s]: %d -> %d",
				b.config.BlockFinalityType, status.lastBlockSeen, currentBlock.Number.Uint64())
			status.previousBlockTime = nil
			status.lastBlockSeen = currentBlock.Number.Uint64()
			return b.nextBlockRequestDelay(status, nil), status
		}

		status.lastBlockSeen = currentBlock.Number.Uint64()

		b.logger.Debugf("New block seen [finality:%s]: %d. blockRate:%s",
			b.config.BlockFinalityType, currentBlock.Number.Uint64(), status.previousBlockTime)
	}
	return b.nextBlockRequestDelay(status, nil), status
}

func (b *BlockNotifierPolling) nextBlockRequestDelay(status *blockNotifierPollingInternalStatus,
	err error) time.Duration {
	if b.config.CheckNewBlockInterval == AutomaticBlockInterval {
		if status == nil {
			return minBlockInterval
		}
		if status.previousBlockTime == nil {
			// First interation is done with maximum precision
			return minBlockInterval
		}
		if status.previousBlockTime != nil {
			now := timeNowFunc()
			expectedTimeNextBlock := status.lastBlockTime.Add(*status.previousBlockTime)
			distanceToNextBlock := expectedTimeNextBlock.Sub(now)
			interval := distanceToNextBlock * 4 / 5 //nolint:mnd //  80% of for reach the next block
			return max(minBlockInterval, min(maxBlockInterval, interval))
		}
	}
	if err == nil {
		return b.config.CheckNewBlockInterval
	}
	// If error we wait twice the interval
	return b.config.CheckNewBlockInterval * 2 //nolint:mnd // 2 times the interval
}

type blockNotifierPollingInternalStatus struct {
	lastBlockSeen     uint64
	lastBlockTime     time.Time      // first appear of block lastBlockSeen
	previousBlockTime *time.Duration // time of the previous block to appear
}

func (s *blockNotifierPollingInternalStatus) String() string {
	if s == nil {
		return "nil"
	}
	return fmt.Sprintf("lastBlockSeen=%d lastBlockTime=%s previousBlockTime=%s",
		s.lastBlockSeen, s.lastBlockTime, s.previousBlockTime)
}

func (s *blockNotifierPollingInternalStatus) clear() *blockNotifierPollingInternalStatus {
	return &blockNotifierPollingInternalStatus{}
}

func (s *blockNotifierPollingInternalStatus) intialBlock(block uint64) *blockNotifierPollingInternalStatus {
	return &blockNotifierPollingInternalStatus{
		lastBlockSeen: block,
		lastBlockTime: timeNowFunc(),
	}
}

func (s *blockNotifierPollingInternalStatus) incommingBlock(block uint64) *blockNotifierPollingInternalStatus {
	return &blockNotifierPollingInternalStatus{
		lastBlockSeen: block,
		lastBlockTime: timeNowFunc(),
	}
}
