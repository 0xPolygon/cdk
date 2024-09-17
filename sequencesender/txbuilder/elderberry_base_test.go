package txbuilder

import (
	"context"
	"testing"

	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestElderberryBaseNewSequence(t *testing.T) {
	opts := bind.TransactOpts{}
	sut := NewTxBuilderElderberryBase(log.GetDefaultLogger(), opts)
	require.NotNil(t, sut)
	seq, err := sut.NewSequence(context.TODO(), nil, common.Address{})
	require.NotNil(t, seq)
	require.NoError(t, err)
}

func TestElderberryBaseNewBatchFromL2Block(t *testing.T) {
	sut := newElderberryBaseSUT(t)
	l2Block := &datastream.L2Block{
		Timestamp:       1,
		BatchNumber:     2,
		L1InfotreeIndex: 3,
		Coinbase:        []byte{1, 2, 3},
		GlobalExitRoot:  []byte{4, 5, 6},
	}
	batch := sut.NewBatchFromL2Block(l2Block)
	require.NotNil(t, batch)
	require.Equal(t, l2Block.Timestamp, batch.LastL2BLockTimestamp())
	require.Equal(t, l2Block.BatchNumber, batch.BatchNumber())
	require.Equal(t, l2Block.L1InfotreeIndex, batch.L1InfoTreeIndex())
	require.Equal(t, common.BytesToAddress(l2Block.Coinbase), batch.LastCoinbase())
	require.Equal(t, common.BytesToHash(l2Block.GlobalExitRoot), batch.GlobalExitRoot())
}

func TestElderberryBasegetLastSequencedBatchNumberEmpty(t *testing.T) {
	sut := newElderberryBaseSUT(t)
	seq, err := sut.NewSequence(context.TODO(), nil, common.Address{})
	require.NoError(t, err)

	require.Equal(t, uint64(0), getLastSequencedBatchNumber(seq))
}

func TestElderberryBasegetLastSequencedBatch1Batch(t *testing.T) {
	sut := newElderberryBaseSUT(t)
	l2Block := &datastream.L2Block{
		Timestamp:       1,
		BatchNumber:     2,
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

	require.Equal(t, l2Block.BatchNumber-1, getLastSequencedBatchNumber(seq))
}

func TestElderberryBaseGetLastSequencedBatchFirstBatchIsZeroThrowAPanic(t *testing.T) {
	sut := newElderberryBaseSUT(t)
	l2Block := &datastream.L2Block{
		Timestamp:       1,
		BatchNumber:     0,
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
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	getLastSequencedBatchNumber(seq)
}

func newElderberryBaseSUT(t *testing.T) *TxBuilderElderberryBase {
	t.Helper()

	opts := bind.TransactOpts{}
	sut := NewTxBuilderElderberryBase(log.GetDefaultLogger(), opts)
	require.NotNil(t, sut)
	return sut
}
