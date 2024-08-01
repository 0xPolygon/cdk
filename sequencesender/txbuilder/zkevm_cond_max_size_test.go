package txbuilder_test

import (
	"context"
	"testing"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder/mocks_txbuilder"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConditionalMaxSizeDisabled(t *testing.T) {
	mockTxBuilder := mocks_txbuilder.NewTxBuilder(t)
	sut := txbuilder.NewConditionalNewSequenceMaxSize(txbuilder.MaxTxSizeForL1Disabled)

	tx, err := sut.NewSequenceIfWorthToSend(nil, mockTxBuilder, nil, common.Address{})
	require.NoError(t, err)
	require.Nil(t, tx)
}

func TestConditionalMaxSizeTxBuilderNewSequenceReturnsNil(t *testing.T) {
	mockTxBuilder := mocks_txbuilder.NewTxBuilder(t)
	sut := txbuilder.NewConditionalNewSequenceMaxSize(1024)
	var sequenceBatches []seqsendertypes.Batch
	sequenceBatches = append(sequenceBatches, &txbuilder.BananaBatch{})
	mockTxBuilder.EXPECT().NewSequence(sequenceBatches, common.Address{}).Return(nil, nil)
	_, err := sut.NewSequenceIfWorthToSend(nil, mockTxBuilder, sequenceBatches, common.Address{})
	require.Error(t, err)
}

func TestConditionalMaxSizeTxBuilderBuildSequenceBatchesTxReturnsNil(t *testing.T) {
	mockTxBuilder := mocks_txbuilder.NewTxBuilder(t)
	sut := txbuilder.NewConditionalNewSequenceMaxSize(1024)
	var sequenceBatches []seqsendertypes.Batch
	sequenceBatches = append(sequenceBatches, &txbuilder.BananaBatch{})
	seq := &txbuilder.ElderberrySequence{}
	mockTxBuilder.EXPECT().NewSequence(sequenceBatches, common.Address{}).Return(seq, nil)
	mockTxBuilder.EXPECT().BuildSequenceBatchesTx(mock.Anything, mock.Anything).Return(nil, nil)
	_, err := sut.NewSequenceIfWorthToSend(nil, mockTxBuilder, sequenceBatches, common.Address{})
	require.Error(t, err)
}

func TestConditionalMaxSizeTxBuilderDontFulFill(t *testing.T) {
	mockTxBuilder := mocks_txbuilder.NewTxBuilder(t)
	sut := txbuilder.NewConditionalNewSequenceMaxSize(1024)
	var sequenceBatches []seqsendertypes.Batch
	sequenceBatches = append(sequenceBatches, &txbuilder.BananaBatch{})
	seq := &txbuilder.ElderberrySequence{}
	mockTxBuilder.EXPECT().NewSequence(sequenceBatches, common.Address{}).Return(seq, nil)
	inner := &ethtypes.LegacyTx{}
	tx := ethtypes.NewTx(inner)
	mockTxBuilder.EXPECT().BuildSequenceBatchesTx(mock.Anything, mock.Anything).Return(tx, nil)

	res, err := sut.NewSequenceIfWorthToSend(nil, mockTxBuilder, sequenceBatches, common.Address{})

	require.NoError(t, err)
	require.Nil(t, res)
}

func TestConditionalMaxSizeTxBuilderFulFill(t *testing.T) {
	mockTxBuilder := mocks_txbuilder.NewTxBuilder(t)
	sut := txbuilder.NewConditionalNewSequenceMaxSize(10)
	l2coinbase := common.Address{}
	ctx := context.TODO()

	newSeq := newTestSeq(3, 100, l2coinbase)
	mockTxBuilder.EXPECT().NewSequence(newSeq.Batches(), l2coinbase).Return(newSeq, nil)
	inner := &ethtypes.LegacyTx{
		Data: []byte{0x01, 0x02, 0x03, 0x04},
	}
	tx := ethtypes.NewTx(inner)
	mockTxBuilder.EXPECT().BuildSequenceBatchesTx(ctx, newSeq).Return(tx, nil)
	// The size of result Tx is 14 that is > 10, so it reduce 1 batch
	newSeqReduced := newTestSeq(2, 100, l2coinbase)
	mockTxBuilder.EXPECT().NewSequence(newSeqReduced.Batches(), l2coinbase).Return(newSeqReduced, nil)
	mockTxBuilder.EXPECT().BuildSequenceBatchesTx(ctx, newSeqReduced).Return(tx, nil)

	res, err := sut.NewSequenceIfWorthToSend(ctx, mockTxBuilder, newSeq.Batches(), l2coinbase)

	require.NoError(t, err)
	require.NotNil(t, res)
}

func newTestSeq(numBatches int, firstBatch uint64, l2coinbase common.Address) *txbuilder.ElderberrySequence {
	var sequenceBatches []seqsendertypes.Batch
	for i := 0; i < numBatches; i++ {
		sequenceBatches = append(sequenceBatches, txbuilder.NewBananaBatch(&etherman.Batch{BatchNumber: firstBatch + uint64(i)}))
	}
	return txbuilder.NewElderberrySequence(sequenceBatches, l2coinbase)
}
