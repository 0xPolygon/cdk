package test

import (
	"math/big"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/test/operations"
	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"
)

func TestEthTransfer(t *testing.T) {
	const amountOfTxsToSend = 10
	client, err := jsonrpc.NewClient(operations.L2RPCURL)
	require.NoError(t, err)
	key, err := operations.GetL2Wallet()
	require.NoError(t, err)
	signer := wallet.NewEIP155Signer(operations.L2ChainID)
	initialNonce, err := client.Eth().GetNonce(key.Address(), ethgo.Latest)
	require.NoError(t, err)
	gasPrice, err := client.Eth().GasPrice()
	require.NoError(t, err)

	txHashes := []ethgo.Hash{}
	for i := 0; i < amountOfTxsToSend; i++ {
		tx := &ethgo.Transaction{
			ChainID:  big.NewInt(operations.L2ChainID),
			Nonce:    initialNonce + uint64(i),
			GasPrice: gasPrice,
			Gas:      21000,
			To:       &ethgo.ZeroAddress,
		}
		signedTxn, err := signer.SignTx(tx, key)
		require.NoError(t, err)
		rawTx, err := signedTxn.MarshalRLPTo(nil)
		require.NoError(t, err)
		hash, err := client.Eth().SendRawTransaction(rawTx)
		require.NoError(t, err)
		txHashes = append(txHashes, hash)
	}
	for _, h := range txHashes {
		require.NoError(t, operations.WaitTxToBeMined(h, operations.Verified, client, time.Minute*2))
	}
}
