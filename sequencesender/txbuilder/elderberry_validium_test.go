package txbuilder_test

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk/dataavailability/mocks_da"
	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestElderberryValidiumName(t *testing.T) {
	testData := newElderberryValidiumSUT(t)
	require.NotNil(t, testData.sut)
	require.True(t, strings.Contains(testData.sut.String(), "Elderberry"))
	require.True(t, strings.Contains(testData.sut.String(), "Validium"))
}

func TestElderberryValidiumBuildSequenceBatchesTxEmtpySequence(t *testing.T) {
	testData := newElderberryValidiumSUT(t)
	ctx := context.TODO()
	_, err := testData.sut.BuildSequenceBatchesTx(ctx, nil)
	require.Error(t, err)

	seq, err := testData.sut.NewSequence(nil, common.Address{})
	require.NoError(t, err)
	_, err = testData.sut.BuildSequenceBatchesTx(ctx, seq)
	require.Error(t, err)
}

func TestElderberryValidiumBuildSequenceBatchesTxSequenceErrorsFromDA(t *testing.T) {
	testData := newElderberryValidiumSUT(t)
	ctx := context.TODO()
	l2Block := &datastream.L2Block{
		Timestamp:       1,
		BatchNumber:     1,
		L1InfotreeIndex: 3,
		Coinbase:        []byte{1, 2, 3},
		GlobalExitRoot:  []byte{4, 5, 6},
	}
	batchElder := testData.sut.NewBatchFromL2Block(l2Block)
	batches := []seqsendertypes.Batch{
		batchElder,
	}
	seq, err := testData.sut.NewSequence(batches, common.Address{})
	require.NoError(t, err)
	testData.mockDA.EXPECT().PostSequenceElderberry(ctx, mock.Anything).Return(nil, nil)
	_, err = testData.sut.BuildSequenceBatchesTx(ctx, seq)
	require.Error(t, err, "data availability message is nil")
	testData.mockDA.EXPECT().PostSequenceElderberry(ctx, mock.Anything).Return(nil, fmt.Errorf("test error"))
	_, err = testData.sut.BuildSequenceBatchesTx(ctx, seq)
	require.Error(t, err, "error posting sequences to the data availability protocol: test error")
}

func TestElderberryValidiumBuildSequenceBatchesTxSequenceDAOk(t *testing.T) {
	testData := newElderberryValidiumSUT(t)
	ctx := context.TODO()
	l2Block := &datastream.L2Block{
		Timestamp:       1,
		BatchNumber:     1,
		L1InfotreeIndex: 3,
		Coinbase:        []byte{1, 2, 3},
		GlobalExitRoot:  []byte{4, 5, 6},
	}
	batchElder := testData.sut.NewBatchFromL2Block(l2Block)
	batches := []seqsendertypes.Batch{
		batchElder,
	}
	seq, err := testData.sut.NewSequence(batches, common.Address{})
	require.NoError(t, err)
	testData.mockDA.EXPECT().PostSequenceElderberry(ctx, mock.Anything).Return([]byte{1}, nil)
	tx, err := testData.sut.BuildSequenceBatchesTx(ctx, seq)
	require.NoError(t, err)
	require.NotNil(t, tx)
}

type testDataElderberryValidium struct {
	mockDA *mocks_da.SequenceSenderElderberry
	sut    *txbuilder.TxBuilderElderberryValidium
}

func newElderberryValidiumSUT(t *testing.T) *testDataElderberryValidium {
	zkevmContract, err := contracts.NewContractBase(polygonvalidiumetrog.NewPolygonvalidiumetrog, common.Address{}, nil, contracts.ContractNameRollup, contracts.VersionElderberry)
	require.NoError(t, err)
	privateKey, err := crypto.HexToECDSA("64e679029f5032046955d41713dcc4b565de77ab891748d31bcf38864b54c175")
	require.NoError(t, err)
	opts, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1))
	require.NoError(t, err)

	da := mocks_da.NewSequenceSenderElderberry(t)

	sut := txbuilder.NewTxBuilderElderberryValidium(*zkevmContract, da, *opts, uint64(100))
	require.NotNil(t, sut)
	return &testDataElderberryValidium{
		mockDA: da,
		sut:    sut,
	}
}
