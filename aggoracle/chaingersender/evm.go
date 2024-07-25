package chaingersender

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/manual/pessimisticglobalexitroot"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygonHermez/zkevm-ethtx-manager/ethtxmanager"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	waitPeriodMonitorTx = time.Second * 5
)

type EthClienter interface {
	ethereum.LogFilterer
	ethereum.BlockNumberReader
	ethereum.ChainReader
	bind.ContractBackend
}

type EthTxManager interface {
	Remove(ctx context.Context, id common.Hash) error
	ResultsByStatus(ctx context.Context, statuses []ethtxmanager.MonitoredTxStatus) ([]ethtxmanager.MonitoredTxResult, error)
	Result(ctx context.Context, id common.Hash) (ethtxmanager.MonitoredTxResult, error)
	Add(ctx context.Context, to *common.Address, forcedNonce *uint64, value *big.Int, data []byte, gasOffset uint64, sidecar *types.BlobTxSidecar) (common.Hash, error)
}

type EVMChainGERSender struct {
	gerContract *pessimisticglobalexitroot.Pessimisticglobalexitroot
	gerAddr     common.Address
	sender      common.Address
	client      EthClienter
	ethTxMan    EthTxManager
	gasOffset   uint64
}

func NewEVMChainGERSender(
	globalExitRoot, sender common.Address,
	client EthClienter,
	ethTxMan EthTxManager,
	gasOffset uint64,
) (*EVMChainGERSender, error) {
	gerContract, err := pessimisticglobalexitroot.NewPessimisticglobalexitroot(globalExitRoot, client)
	if err != nil {
		return nil, err
	}
	return &EVMChainGERSender{
		gerContract: gerContract,
		gerAddr:     globalExitRoot,
		sender:      sender,
		client:      client,
		ethTxMan:    ethTxMan,
		gasOffset:   gasOffset,
	}, nil
}

func (c *EVMChainGERSender) IsGERAlreadyInjected(ger common.Hash) (bool, error) {
	timestamp, err := c.gerContract.GlobalExitRootMap(&bind.CallOpts{Pending: false}, ger)
	if err != nil {
		return false, err
	}
	return timestamp.Cmp(big.NewInt(0)) != 0, nil
}

func (c *EVMChainGERSender) UpdateGERWaitUntilMined(ctx context.Context, ger common.Hash) error {
	abi, err := pessimisticglobalexitroot.PessimisticglobalexitrootMetaData.GetAbi()
	if err != nil {
		return err
	}
	data, err := abi.Pack("updateGlobalExitRoot", ger)
	id, err := c.ethTxMan.Add(ctx, &c.gerAddr, nil, big.NewInt(0), data, c.gasOffset, nil)
	if err != nil {
		return err
	}
	for {
		time.Sleep(waitPeriodMonitorTx)
		res, err := c.ethTxMan.Result(ctx, id)
		if err != nil {
			log.Error("error calling ethTxMan.Result: ", err)
		}
		switch res.Status {
		case ethtxmanager.MonitoredTxStatusCreated:
			continue
		case ethtxmanager.MonitoredTxStatusSent:
			continue
		case ethtxmanager.MonitoredTxStatusFailed:
			return fmt.Errorf("tx %s failed", res.ID)
		case ethtxmanager.MonitoredTxStatusMined:
			return nil
		case ethtxmanager.MonitoredTxStatusSafe:
			return nil
		case ethtxmanager.MonitoredTxStatusFinalized:
			return nil
		default:
			log.Error("unexpected tx status: ", res.Status)
		}
	}
}
