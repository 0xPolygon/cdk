package ethtxmanager

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/stretchr/testify/assert"
)

func TestTx(t *testing.T) {
	to := common.HexToAddress("0x2")
	nonce := uint64(1)
	value := big.NewInt(2)
	data := []byte("data")
	gas := uint64(3)
	gasOffset := uint64(4)
	gasPrice := big.NewInt(5)

	mTx := monitoredTx{
		To:        &to,
		Nonce:     nonce,
		Value:     value,
		Data:      data,
		Gas:       gas,
		GasOffset: gasOffset,
		GasPrice:  gasPrice,
	}

	tx := mTx.Tx()

	assert.Equal(t, &to, tx.To())
	assert.Equal(t, nonce, tx.Nonce())
	assert.Equal(t, value, tx.Value())
	assert.Equal(t, data, tx.Data())
	assert.Equal(t, gas+gasOffset, tx.Gas())
	assert.Equal(t, gasPrice, tx.GasPrice())
}

func TestBlobTx(t *testing.T) {
	client, _ := New(Config{}, nil, common.Address{})
	to := common.HexToAddress("0x2")
	nonce := uint64(1)
	value := big.NewInt(2)
	data := []byte{}
	gas := uint64(3)
	gasOffset := uint64(4)
	blobGas := uint64(131072)
	blobGasPrice := big.NewInt(10)

	blobBytes := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	blob, err := client.EncodeBlobData(blobBytes)
	assert.NoError(t, err)
	blobSidecar := client.MakeBlobSidecar([]kzg4844.Blob{blob})

	mTx := monitoredTx{
		To:           &to,
		Nonce:        nonce,
		Value:        value,
		Data:         data,
		Gas:          gas,
		GasOffset:    gasOffset,
		BlobSidecar:  blobSidecar,
		BlobGas:      blobGas,
		BlobGasPrice: blobGasPrice,
	}

	tx := mTx.Tx()

	assert.Equal(t, &to, tx.To())
	assert.Equal(t, nonce, tx.Nonce())
	assert.Equal(t, value, tx.Value())
	assert.Equal(t, data, tx.Data())
	assert.Equal(t, blobSidecar, tx.BlobTxSidecar())
	assert.Equal(t, blobGas, tx.BlobGas())
	assert.Equal(t, blobGasPrice, tx.BlobGasFeeCap())
}

func TestAddHistory(t *testing.T) {
	tx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), nil)
	mTx := monitoredTx{
		History: make(map[common.Hash]bool),
	}

	err := mTx.AddHistory(tx)
	assert.NoError(t, err)

	// Adding the same transaction again should return an error
	err = mTx.AddHistory(tx)
	assert.ErrorContains(t, err, "already exists")

	// should have only one history
	historySlice := mTx.historyHashSlice()
	assert.Len(t, historySlice, 1)
}
