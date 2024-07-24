package aggoracle

import (
	"context"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

var (
	waitPeriodNextGER = time.Second * 30
)

type EthClienter interface {
	ethereum.LogFilterer
	ethereum.BlockNumberReader
	ethereum.ChainReader
	bind.ContractBackend
}

type L1InfoTreer interface {
	GetLatestInfoUntilBlock(ctx context.Context, blockNum uint64) (*l1infotreesync.L1InfoTreeLeaf, error)
}

type ChainSender interface {
	IsGERAlreadyInjected(ger common.Hash) (bool, error)
	UpdateGERWaitUntilMined(ctx context.Context, ger common.Hash) error
}

type AggOracle struct {
	ticker        *time.Ticker
	l1Client      EthClienter
	l1Info        L1InfoTreer
	chainSender   ChainSender
	blockFinality *big.Int
}

func New(
	chainSender ChainSender,
	l1Client EthClienter,
	l1InfoTreeSyncer L1InfoTreer,
	blockFinalityType etherman.BlockNumberFinality,
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
	for {
		select {
		case <-a.ticker.C:
			gerToInject, err := a.getLastFinalisedGER(ctx)
			if err != nil {
				log.Error("error calling getLastFinalisedGER: ", err)
				continue
			}
			if alreadyInjectd, err := a.chainSender.IsGERAlreadyInjected(gerToInject); err != nil {
				log.Error("error calling isGERAlreadyInjected: ", err)
				continue
			} else if alreadyInjectd {
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

func (a *AggOracle) getLastFinalisedGER(ctx context.Context) (common.Hash, error) {
	header, err := a.l1Client.HeaderByNumber(ctx, a.blockFinality)
	if err != nil {
		return common.Hash{}, err
	}
	info, err := a.l1Info.GetLatestInfoUntilBlock(ctx, header.Number.Uint64())
	if err != nil {
		return common.Hash{}, err
	}
	return info.GlobalExitRoot, nil
}
