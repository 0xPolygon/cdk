package helpers

import (
	"context"
	"encoding/hex"
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
		argReceiverIdx = 1
		argTxInputIdx  = 3
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

			to, ok := args.Get(argReceiverIdx).(*common.Address)
			if !ok {
				log.Error("expected *common.Address for tx receiver arg")
				return
			}

			data, ok := args.Get(argTxInputIdx).([]byte)
			if !ok {
				log.Error("expected []byte for tx input data arg")
				return
			}

			log.Debugf("receiver %s, data: %s", to, hex.EncodeToString(data))

			msg := ethereum.CallMsg{
				From: auth.From,
				To:   to,
				Data: data,
			}

			gas, err := client.Client().EstimateGas(ctx, msg)
			if err != nil {
				log.Errorf("eth_estimateGas invocation failed: %s", err)

				res, err := client.Client().CallContract(ctx, msg, nil)
				if err != nil {
					log.Errorf("eth_call invocation failed: %s", err)
				} else {
					log.Debugf("contract call result: %s", hex.EncodeToString(res))
				}
				return
			}
			price, err := client.Client().SuggestGasPrice(ctx)
			if err != nil {
				log.Error(err)
			}

			tx := types.NewTx(&types.LegacyTx{
				To:       to,
				Nonce:    nonce,
				Value:    common.Big0,
				Data:     data,
				Gas:      gas,
				GasPrice: price,
			})
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
	ethTxMock.On("Result", mock.Anything, mock.Anything).
		Return(ethtxtypes.MonitoredTxResult{Status: ethtxtypes.MonitoredTxStatusMined}, nil)

	return ethTxMock
}
