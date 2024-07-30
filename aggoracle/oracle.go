package aggoracle

import (
	"context"
	"math/big"
	"time"

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
	IsGERAlreadyInjected(ger common.Hash) (bool, error)
	UpdateGERWaitUntilMined(ctx context.Context, ger common.Hash) error
}

type AggOracle struct {
	ticker        *time.Ticker
	l1Client      ethereum.ChainReader
	l1Info        L1InfoTreer
	chainSender   ChainSender
	blockFinality *big.Int
}

func New(
	chainSender ChainSender,
	l1Client ethereum.ChainReader,
	l1InfoTreeSyncer L1InfoTreer,
	blockFinalityType etherman.BlockNumberFinality,
	waitPeriodNextGER time.Duration,
) (*AggOracle, error) {
	ticker := time.NewTicker(waitPeriodNextGER)
	finality, err := blockFinalityType.ToBlockNum()
	if err != nil {
		return nil, err
	}
	return &AggOracle{
		ticker:        ticker,
		l1Client:      l1Client,
		l1Info:        l1InfoTreeSyncer,
		chainSender:   chainSender,
		blockFinality: finality,
	}, nil
}

func (a *AggOracle) Start(ctx context.Context) {
	var (
		blockNumToFetch uint64
		gerToInject     common.Hash
		err             error
	)
	for {
		select {
		case <-a.ticker.C:
			blockNumToFetch, gerToInject, err = a.getLastFinalisedGER(ctx, blockNumToFetch)
			if err != nil {
				if err == l1infotreesync.ErrBlockNotProcessed {
					log.Debugf("syncer is not ready for the block %d", blockNumToFetch)
				} else if err == l1infotreesync.ErrNotFound {
					blockNumToFetch = 0
					log.Debugf("syncer has not found any GER until block %d", blockNumToFetch)
				} else {
					log.Error("error calling getLastFinalisedGER: ", err)
				}
				continue
			}
			if alreadyInjected, err := a.chainSender.IsGERAlreadyInjected(gerToInject); err != nil {
				log.Error("error calling isGERAlreadyInjected: ", err)
				continue
			} else if alreadyInjected {
				log.Debugf("GER %s already injected", gerToInject.Hex())
				continue
			}
			log.Debugf("injecting new GER: %s", gerToInject.Hex())
			if err := a.chainSender.UpdateGERWaitUntilMined(ctx, gerToInject); err != nil {
				log.Errorf("error calling updateGERWaitUntilMined, when trying to inject GER %s: %v", gerToInject.Hex(), err)
				continue
			}
			log.Debugf("GER %s injected", gerToInject.Hex())
		case <-ctx.Done():
			return
		}
	}
}

// getLastFinalisedGER tries to return a finalised GER:
// If blockNumToFetch != 0: it will try to fetch it until the given block
// Else it will ask the L1 client for the latest finalised block and use that
// If it fails to get the GER from the syncer, it will retunr the block number that used to query
func (a *AggOracle) getLastFinalisedGER(ctx context.Context, blockNumToFetch uint64) (uint64, common.Hash, error) {
	if blockNumToFetch == 0 {
		header, err := a.l1Client.HeaderByNumber(ctx, a.blockFinality)
		if err != nil {
			return 0, common.Hash{}, err
		}
		blockNumToFetch = header.Number.Uint64()
	}
	info, err := a.l1Info.GetLatestInfoUntilBlock(ctx, blockNumToFetch)
	if err != nil {
		return blockNumToFetch, common.Hash{}, err
	}
	return 0, info.GlobalExitRoot, nil
}
