package txbuilder_test

import (
	"context"
	"math/big"
	"strings"
	"testing"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestElderberryZkevmName(t *testing.T) {
	zkevmContract := contracts.RollupElderberryType{}
	opts := bind.TransactOpts{}
	sut := txbuilder.NewTxBuilderElderberryZKEVM(zkevmContract, opts, 100)
	require.NotNil(t, sut)
	require.True(t, strings.Contains(sut.String(), "Elderberry"))
	require.True(t, strings.Contains(sut.String(), "ZKEVM"))
}

func TestElderberryZkevmNewSequence(t *testing.T) {
	zkevmContract := contracts.RollupElderberryType{}
	opts := bind.TransactOpts{}
	sut := txbuilder.NewTxBuilderElderberryZKEVM(zkevmContract, opts, 100)
	require.NotNil(t, sut)
	seq, err := sut.NewSequence(context.TODO(), nil, common.Address{})
	require.NoError(t, err)
	require.NotNil(t, seq)
}

func TestElderberryZkevmBuildSequenceBatchesTxEmtpySequence(t *testing.T) {
	sut := newElderberryZkevmSUT(t)
	ctx := context.TODO()
	_, err := sut.BuildSequenceBatchesTx(ctx, nil)
	require.Error(t, err)

	seq, err := sut.NewSequence(context.TODO(), nil, common.Address{})
	require.NoError(t, err)
	_, err = sut.BuildSequenceBatchesTx(ctx, seq)
	require.Error(t, err)
}

func TestElderberryZkevmBuildSequenceBatchesTxSequence1Batch(t *testing.T) {
	sut := newElderberryZkevmSUT(t)
	ctx := context.TODO()
	l2Block := &datastream.L2Block{
		Timestamp:       1,
		BatchNumber:     1,
		L1InfotreeIndex: 3,
		Coinbase:        []byte{1, 2, 3},
		GlobalExitRoot:  []byte{4, 5, 6},
	}
	batchElder := sut.NewBatchFromL2Block(l2Block)
	batches := []seqsendertypes.Batch{
		batchElder,
	}
	seq, err := sut.NewSequence(context.TODO(), batches, common.Address{})
	require.NoError(t, err)
	_, err = sut.BuildSequenceBatchesTx(ctx, seq)
	require.NoError(t, err)
}

// This have to signer so produce an error
func TestElderberryZkevmBuildSequenceBatchesTxSequence1BatchError(t *testing.T) {
	sut := newElderberryZkevmSUT(t)
	sut.SetAuth(&bind.TransactOpts{})
	ctx := context.TODO()
	l2Block := &datastream.L2Block{
		Timestamp:       1,
		BatchNumber:     1,
		L1InfotreeIndex: 3,
		Coinbase:        []byte{1, 2, 3},
		GlobalExitRoot:  []byte{4, 5, 6},
	}
	batchElder := sut.NewBatchFromL2Block(l2Block)
	batches := []seqsendertypes.Batch{
		batchElder,
	}
	seq, err := sut.NewSequence(context.TODO(), batches, common.Address{})
	require.NoError(t, err)
	_, err = sut.BuildSequenceBatchesTx(ctx, seq)
	require.Error(t, err)
}

func TestElderberryZkevmNewSequenceIfWorthToSend(t *testing.T) {
	sut := newElderberryZkevmSUT(t)
	testSequenceIfWorthToSendNoNewSeq(t, sut)
	testSequenceIfWorthToSendErr(t, sut)
	testSetCondNewSeq(t, sut)
}

func newElderberryZkevmSUT(t *testing.T) *txbuilder.TxBuilderElderberryZKEVM {
	zkevmContract, err := contracts.NewContractMagic[contracts.RollupElderberryType](polygonvalidiumetrog.NewPolygonvalidiumetrog, common.Address{}, nil, contracts.ContractNameRollup, contracts.VersionElderberry)
	require.NoError(t, err)
	privateKey, err := crypto.HexToECDSA("64e679029f5032046955d41713dcc4b565de77ab891748d31bcf38864b54c175")
	require.NoError(t, err)
	opts, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1))
	require.NoError(t, err)
	sut := txbuilder.NewTxBuilderElderberryZKEVM(*zkevmContract, *opts, 100)
	require.NotNil(t, sut)
	return sut
}
