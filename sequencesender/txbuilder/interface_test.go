package txbuilder_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder/mocks_txbuilder"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

/*
This test ara auxiliars function based on the common behaviour of the interfaces
*/

func testSequenceIfWorthToSendNoNewSeq(t *testing.T, sut txbuilder.TxBuilder) {
	t.Helper()

	cond := mocks_txbuilder.NewCondNewSequence(t)
	sut.SetCondNewSeq(cond)
	cond.EXPECT().NewSequenceIfWorthToSend(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Once()
	seq, err := sut.NewSequenceIfWorthToSend(context.TODO(), nil, common.Address{}, 0)
	require.NoError(t, err)
	require.Nil(t, seq)
}

func testSequenceIfWorthToSendErr(t *testing.T, sut txbuilder.TxBuilder) {
	t.Helper()

	cond := mocks_txbuilder.NewCondNewSequence(t)
	sut.SetCondNewSeq(cond)
	returnErr := fmt.Errorf("test-error")
	cond.EXPECT().NewSequenceIfWorthToSend(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, returnErr).Once()
	seq, err := sut.NewSequenceIfWorthToSend(context.TODO(), nil, common.Address{}, 0)
	require.ErrorIs(t, returnErr, err)
	require.Nil(t, seq)
}

func testSetCondNewSeq(t *testing.T, sut txbuilder.TxBuilder) {
	t.Helper()

	cond := mocks_txbuilder.NewCondNewSequence(t)
	sut.SetCondNewSeq(cond)
	cond2 := mocks_txbuilder.NewCondNewSequence(t)
	previous := sut.SetCondNewSeq(cond2)
	require.Equal(t, cond, previous)
}
