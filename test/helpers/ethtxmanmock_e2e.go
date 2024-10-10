package helpers

import (
	"context"
	"math/big"
	"testing"

	"github.com/0xPolygon/cdk/log"
	ethtxtypes "github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/mock"
)

func NewEthTxManMock(
	t *testing.T,
	client *simulated.Backend,
	auth *bind.TransactOpts,
) *EthTxManagerMock {
	t.Helper()

	const (
		ArgToIndex   = 1
		ArgDataIndex = 3
		ZeroValue    = 0
	)

	ethTxMock := NewEthTxManagerMock(t)
	ethTxMock.On(
		"Add", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ctx := context.Background()
			nonce, err := client.Client().PendingNonceAt(ctx, auth.From)
			if err != nil {
				log.Error(err)
				return
			}
			gas, err := client.Client().EstimateGas(ctx, ethereum.CallMsg{
				From:  auth.From,
				To:    args.Get(ArgToIndex).(*common.Address),
				Value: big.NewInt(ZeroValue),
				Data:  args.Get(ArgDataIndex).([]byte),
			})
			if err != nil {
				log.Error(err)
				res, err := client.Client().CallContract(ctx, ethereum.CallMsg{
					From:  auth.From,
					To:    args.Get(ArgToIndex).(*common.Address),
					Value: big.NewInt(ZeroValue),
					Data:  args.Get(ArgDataIndex).([]byte),
				}, nil)
				log.Debugf("contract call: %s", res)
				if err != nil {
					log.Errorf("%+v", err)
				}
				return
			}
			price, err := client.Client().SuggestGasPrice(ctx)
			if err != nil {
				log.Error(err)
			}

			to, ok := args.Get(ArgToIndex).(*common.Address)
			if !ok {
				log.Error("expected *common.Address for ArgToIndex")
				return
			}
			data, ok := args.Get(ArgDataIndex).([]byte)
			if !ok {
				log.Error("expected []byte for ArgDataIndex")
				return
			}
			tx := types.NewTx(&types.LegacyTx{
				To:       to,
				Nonce:    nonce,
				Value:    big.NewInt(ZeroValue),
				Data:     data,
				Gas:      gas,
				GasPrice: price,
			})
			tx.Gas()
			signedTx, err := auth.Signer(auth.From, tx)
			if err != nil {
				log.Error(err)
				return
			}
			err = client.Client().SendTransaction(ctx, signedTx)
			if err != nil {
				log.Error(err)
				return
			}
			client.Commit()
		}).
		Return(common.Hash{}, nil)
	// res, err := c.ethTxMan.Result(ctx, id)
	ethTxMock.On("Result", mock.Anything, mock.Anything).
		Return(ethtxtypes.MonitoredTxResult{Status: ethtxtypes.MonitoredTxStatusMined}, nil)

	return ethTxMock
}
