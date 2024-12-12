package aggoracle

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

type L1InfoTreer interface {
	GetLatestInfoUntilBlock(ctx context.Context, blockNum uint64) (*l1infotreesync.L1InfoTreeLeaf, error)
}

type ChainSender interface {
	IsGERInjected(ger common.Hash) (bool, error)
	InjectGER(ctx context.Context, ger common.Hash) error
}

type AggOracle struct {
	logger            *log.Logger
	waitPeriodNextGER time.Duration
	l1Client          ethereum.ChainReader
	l1Info            L1InfoTreer
	chainSender       ChainSender
	blockFinality     *big.Int
}

func New(
	logger *log.Logger,
	chainSender ChainSender,
	l1Client ethereum.ChainReader,
	l1InfoTreeSyncer L1InfoTreer,
	blockFinalityType etherman.BlockNumberFinality,
	waitPeriodNextGER time.Duration,
) (*AggOracle, error) {
	finality, err := blockFinalityType.ToBlockNum()
	if err != nil {
		return nil, err
	}

	return &AggOracle{
		logger:            logger,
		chainSender:       chainSender,
		l1Client:          l1Client,
		l1Info:            l1InfoTreeSyncer,
		blockFinality:     finality,
		waitPeriodNextGER: waitPeriodNextGER,
	}, nil
}

func (a *AggOracle) Start(ctx context.Context) {
	ticker := time.NewTicker(a.waitPeriodNextGER)
	defer ticker.Stop()

	var blockNumToFetch uint64

	for {
		select {
		case <-ticker.C:
			if err := a.processLatestGER(ctx, &blockNumToFetch); err != nil {
				a.handleGERProcessingError(err, blockNumToFetch)
			}
		case <-ctx.Done():
			return
		}
	}
}

// processLatestGER fetches the latest finalized GER, checks if it is already injected and injects it if not
func (a *AggOracle) processLatestGER(ctx context.Context, blockNumToFetch *uint64) error {
	// Fetch the latest GER
	blockNum, gerToInject, err := a.getLastFinalizedGER(ctx, *blockNumToFetch)
	if err != nil {
		return err
	}

	// Update the block number for the next iteration
	*blockNumToFetch = blockNum

	alreadyInjected, err := a.chainSender.IsGERInjected(gerToInject)
	if err != nil {
		return fmt.Errorf("error checking if GER is already injected: %w", err)
	}
	if alreadyInjected {
		a.logger.Debugf("GER %s already injected", gerToInject.Hex())
		return nil
	}

	a.logger.Infof("injecting new GER: %s", gerToInject.Hex())
	if err := a.chainSender.InjectGER(ctx, gerToInject); err != nil {
		return fmt.Errorf("error injecting GER %s: %w", gerToInject.Hex(), err)
	}

	a.logger.Infof("GER %s is injected successfully", gerToInject.Hex())
	return nil
}

// handleGERProcessingError handles global exit root processing error
func (a *AggOracle) handleGERProcessingError(err error, blockNumToFetch uint64) {
	switch {
	case errors.Is(err, l1infotreesync.ErrBlockNotProcessed):
		a.logger.Debugf("syncer is not ready for the block %d", blockNumToFetch)
	case errors.Is(err, db.ErrNotFound):
		a.logger.Debugf("syncer has not found any GER until block %d", blockNumToFetch)
	default:
		a.logger.Error("unexpected error processing GER: ", err)
	}
}

// getLastFinalizedGER tries to return a finalised GER:
// If targetBlockNum != 0: it will try to fetch it until the given block
// Else it will ask the L1 client for the latest finalised block and use that.
// If it fails to get the GER from the syncer, it will return the block number that used to query
func (a *AggOracle) getLastFinalizedGER(ctx context.Context, targetBlockNum uint64) (uint64, common.Hash, error) {
	if targetBlockNum == 0 {
		header, err := a.l1Client.HeaderByNumber(ctx, a.blockFinality)
		if err != nil {
			return 0, common.Hash{}, err
		}
		targetBlockNum = header.Number.Uint64()
	}

	info, err := a.l1Info.GetLatestInfoUntilBlock(ctx, targetBlockNum)
	if err != nil {
		return targetBlockNum, common.Hash{}, err
	}

	return 0, info.GlobalExitRoot, nil
}
