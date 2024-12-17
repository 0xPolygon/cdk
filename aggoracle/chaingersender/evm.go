package chaingersender

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/l2-sovereign-chain-paris/globalexitrootmanagerl2sovereignchain"
	cdkcommon "github.com/0xPolygon/cdk/common"
	cfgtypes "github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/zkevm-ethtx-manager/ethtxmanager"
	ethtxtypes "github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const insertGERFuncName = "insertGlobalExitRoot"

type EthClienter interface {
	ethereum.LogFilterer
	ethereum.BlockNumberReader
	ethereum.ChainReader
	bind.ContractBackend
}

type EthTxManager interface {
	Remove(ctx context.Context, id common.Hash) error
	ResultsByStatus(ctx context.Context,
		statuses []ethtxtypes.MonitoredTxStatus,
	) ([]ethtxtypes.MonitoredTxResult, error)
	Result(ctx context.Context, id common.Hash) (ethtxtypes.MonitoredTxResult, error)
	Add(ctx context.Context,
		to *common.Address,
		value *big.Int,
		data []byte,
		gasOffset uint64,
		sidecar *types.BlobTxSidecar,
	) (common.Hash, error)
}

type EVMConfig struct {
	GlobalExitRootL2Addr common.Address      `mapstructure:"GlobalExitRootL2"`
	URLRPCL2             string              `mapstructure:"URLRPCL2"`
	ChainIDL2            uint64              `mapstructure:"ChainIDL2"`
	GasOffset            uint64              `mapstructure:"GasOffset"`
	WaitPeriodMonitorTx  cfgtypes.Duration   `mapstructure:"WaitPeriodMonitorTx"`
	EthTxManager         ethtxmanager.Config `mapstructure:"EthTxManager"`
}

type EVMChainGERSender struct {
	logger *log.Logger

	l2GERManager     *globalexitrootmanagerl2sovereignchain.Globalexitrootmanagerl2sovereignchain
	l2GERManagerAddr common.Address
	l2GERManagerAbi  *abi.ABI

	ethTxMan            EthTxManager
	gasOffset           uint64
	waitPeriodMonitorTx time.Duration
}

func NewEVMChainGERSender(
	logger *log.Logger,
	l2GERManagerAddr common.Address,
	l2Client EthClienter,
	ethTxMan EthTxManager,
	gasOffset uint64,
	waitPeriodMonitorTx time.Duration,
) (*EVMChainGERSender, error) {
	l2GERManager, err := globalexitrootmanagerl2sovereignchain.NewGlobalexitrootmanagerl2sovereignchain(
		l2GERManagerAddr, l2Client)
	if err != nil {
		return nil, err
	}

	l2GERAbi, err := globalexitrootmanagerl2sovereignchain.Globalexitrootmanagerl2sovereignchainMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	return &EVMChainGERSender{
		logger:              logger,
		l2GERManager:        l2GERManager,
		l2GERManagerAddr:    l2GERManagerAddr,
		l2GERManagerAbi:     l2GERAbi,
		ethTxMan:            ethTxMan,
		gasOffset:           gasOffset,
		waitPeriodMonitorTx: waitPeriodMonitorTx,
	}, nil
}

func (c *EVMChainGERSender) IsGERInjected(ger common.Hash) (bool, error) {
	blockHashBigInt, err := c.l2GERManager.GlobalExitRootMap(&bind.CallOpts{Pending: false}, ger)
	if err != nil {
		return false, fmt.Errorf("failed to check if global exit root is injected %s: %w", ger, err)
	}

	return common.BigToHash(blockHashBigInt) != cdkcommon.ZeroHash, nil
}

func (c *EVMChainGERSender) InjectGER(ctx context.Context, ger common.Hash) error {
	ticker := time.NewTicker(c.waitPeriodMonitorTx)
	defer ticker.Stop()

	updateGERTxInput, err := c.l2GERManagerAbi.Pack(insertGERFuncName, ger)
	if err != nil {
		return err
	}

	id, err := c.ethTxMan.Add(ctx, &c.l2GERManagerAddr, common.Big0, updateGERTxInput, c.gasOffset, nil)
	if err != nil {
		return err
	}

	for {
		<-ticker.C

		c.logger.Debugf("waiting for tx %s to be mined", id.Hex())
		res, err := c.ethTxMan.Result(ctx, id)
		if err != nil {
			c.logger.Errorf("failed to check the transaction %s status: %s", id.Hex(), err)
			continue
		}

		switch res.Status {
		case ethtxtypes.MonitoredTxStatusCreated,
			ethtxtypes.MonitoredTxStatusSent:
			continue
		case ethtxtypes.MonitoredTxStatusFailed:
			return fmt.Errorf("tx %s failed", id.Hex())
		case ethtxtypes.MonitoredTxStatusMined,
			ethtxtypes.MonitoredTxStatusSafe,
			ethtxtypes.MonitoredTxStatusFinalized:
			return nil
		default:
			c.logger.Error("unexpected tx status:", res.Status)
		}
	}
}
