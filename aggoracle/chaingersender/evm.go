package chaingersender

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/l2-sovereign-chain-paris/polygonzkevmglobalexitrootv2"
	cfgTypes "github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/zkevm-ethtx-manager/ethtxmanager"
	ethtxtypes "github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

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

type EVMChainGERSender struct {
	logger              *log.Logger
	gerContract         *polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2
	gerAddr             common.Address
	client              EthClienter
	ethTxMan            EthTxManager
	gasOffset           uint64
	waitPeriodMonitorTx time.Duration
}

type EVMConfig struct {
	GlobalExitRootL2Addr common.Address      `mapstructure:"GlobalExitRootL2"`
	URLRPCL2             string              `mapstructure:"URLRPCL2"`
	ChainIDL2            uint64              `mapstructure:"ChainIDL2"`
	GasOffset            uint64              `mapstructure:"GasOffset"`
	WaitPeriodMonitorTx  cfgTypes.Duration   `mapstructure:"WaitPeriodMonitorTx"`
	EthTxManager         ethtxmanager.Config `mapstructure:"EthTxManager"`
}

func NewEVMChainGERSender(
	logger *log.Logger,
	l2GlobalExitRoot common.Address,
	l2Client EthClienter,
	ethTxMan EthTxManager,
	gasOffset uint64,
	waitPeriodMonitorTx time.Duration,
) (*EVMChainGERSender, error) {
	gerContract, err := polygonzkevmglobalexitrootv2.NewPolygonzkevmglobalexitrootv2(l2GlobalExitRoot, l2Client)
	if err != nil {
		return nil, err
	}

	return &EVMChainGERSender{
		logger:              logger,
		gerContract:         gerContract,
		gerAddr:             l2GlobalExitRoot,
		client:              l2Client,
		ethTxMan:            ethTxMan,
		gasOffset:           gasOffset,
		waitPeriodMonitorTx: waitPeriodMonitorTx,
	}, nil
}

func (c *EVMChainGERSender) IsGERInjected(ger common.Hash) (bool, error) {
	blockHashBigInt, err := c.gerContract.GlobalExitRootMap(&bind.CallOpts{Pending: false}, ger)
	if err != nil {
		return false, fmt.Errorf("error calling gerContract.GlobalExitRootMap: %w", err)
	}

	return common.BigToHash(blockHashBigInt) != common.Hash{}, nil
}

func (c *EVMChainGERSender) InjectGER(ctx context.Context, ger common.Hash) error {
	ticker := time.NewTicker(c.waitPeriodMonitorTx)
	defer ticker.Stop()

	gerABI, err := polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2MetaData.GetAbi()
	if err != nil {
		return err
	}

	// TODO: @Stefan-Ethernal should we invoke the PolygonZkEVMBridgeV2.updateGlobalExitRoot?
	updateGERTxInput, err := gerABI.Pack("updateExitRoot", ger)
	if err != nil {
		return err
	}

	id, err := c.ethTxMan.Add(ctx, &c.gerAddr, big.NewInt(0), updateGERTxInput, c.gasOffset, nil)
	if err != nil {
		return err
	}
	for {
		<-ticker.C

		c.logger.Debugf("waiting for tx %s to be mined", id.Hex())
		res, err := c.ethTxMan.Result(ctx, id)
		if err != nil {
			c.logger.Error("error calling ethTxMan.Result: ", err)
		}

		switch res.Status {
		case ethtxtypes.MonitoredTxStatusCreated,
			ethtxtypes.MonitoredTxStatusSent:
			continue
		case ethtxtypes.MonitoredTxStatusFailed:
			return fmt.Errorf("tx %s failed", res.ID)
		case ethtxtypes.MonitoredTxStatusMined,
			ethtxtypes.MonitoredTxStatusSafe,
			ethtxtypes.MonitoredTxStatusFinalized:
			return nil
		default:
			c.logger.Error("unexpected tx status: ", res.Status)
		}
	}
}
