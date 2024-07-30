package txbuilder_test

import (
	"testing"

	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder/mocks_txbuilder"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConditionalNumBatchesDisabled(t *testing.T) {
	mockTxBuilder := mocks_txbuilder.NewTxBuilder(t)
	sut := txbuilder.NewConditionalNewSequenceNumBatches(0)

	tx, err := sut.NewSequenceIfWorthToSend(nil, mockTxBuilder, nil, common.Address{})
	require.NoError(t, err)
	require.Nil(t, tx)
}

// It have 1 batch and minium are 2, so no new sequence
func TestConditionalNumBatchesDontFulfillCondition(t *testing.T) {
	mockTxBuilder := mocks_txbuilder.NewTxBuilder(t)
	sut := txbuilder.NewConditionalNewSequenceNumBatches(2)
	var sequenceBatches []seqsendertypes.Batch
	sequenceBatches = append(sequenceBatches, &txbuilder.BananaBatch{})
	tx, err := sut.NewSequenceIfWorthToSend(nil, mockTxBuilder, sequenceBatches, common.Address{})
	require.NoError(t, err)
	require.Nil(t, tx)
}

// It have 2 batch and minium are 2, so  new sequence
func TestConditionalNumBatchesFulfillCondition(t *testing.T) {
	mockTxBuilder := mocks_txbuilder.NewTxBuilder(t)
	sut := txbuilder.NewConditionalNewSequenceNumBatches(2)
	var sequenceBatches []seqsendertypes.Batch
	sequenceBatches = append(sequenceBatches, &txbuilder.BananaBatch{})
	sequenceBatches = append(sequenceBatches, &txbuilder.BananaBatch{})
	mockTxBuilder.EXPECT().NewSequence(mock.Anything, mock.Anything).Return(nil, nil)
	tx, err := sut.NewSequenceIfWorthToSend(nil, mockTxBuilder, sequenceBatches, common.Address{})
	require.NoError(t, err)
	require.Nil(t, tx)
}
