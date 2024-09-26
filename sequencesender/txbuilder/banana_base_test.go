package txbuilder_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder/mocks_txbuilder"
	"github.com/0xPolygon/cdk/state"
	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBananaBaseNewSequenceEmpty(t *testing.T) {
	testData := newBananaBaseTestData(t)
	testData.l1Client.On("HeaderByNumber", mock.Anything, mock.Anything).
		Return(&types.Header{Number: big.NewInt(69)}, nil)
	testData.getContract.On("L1InfoRootMap", mock.Anything, uint32(70)).
		Return([32]byte{}, nil)
	testData.l1InfoTreeSync.On("GetLatestInfoUntilBlock", mock.Anything, mock.Anything).
		Return(&l1infotreesync.L1InfoTreeLeaf{L1InfoTreeIndex: 69}, nil)
	lastAcc := common.HexToHash("0x8aca9664752dbae36135fd0956c956fc4a370feeac67485b49bcd4b99608ae41")
	testData.rollupContract.EXPECT().LastAccInputHash(mock.Anything).Return(lastAcc, nil)
	testData.l1InfoTreeSync.EXPECT().GetInitL1InfoRootMap(mock.Anything).Return(nil, l1infotreesync.ErrNotFound)
	seq, err := testData.sut.NewSequence(context.TODO(), nil, common.Address{})
	require.NotNil(t, seq)
	require.NoError(t, err)
}

func TestBananaBaseNewSequenceErrorHeaderByNumber(t *testing.T) {
	testData := newBananaBaseTestData(t)
	testData.l1Client.On("HeaderByNumber", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("error"))
	seq, err := testData.sut.NewSequence(context.TODO(), nil, common.Address{})
	require.Nil(t, seq)
	require.Error(t, err)
}

func TestBananaBaseNewBatchFromL2Block(t *testing.T) {
	testData := newBananaBaseTestData(t)
	l2Block := &datastream.L2Block{
		Timestamp:       1,
		BatchNumber:     2,
		L1InfotreeIndex: 3,
		Coinbase:        []byte{1, 2, 3},
		GlobalExitRoot:  []byte{4, 5, 6},
	}
	batch := testData.sut.NewBatchFromL2Block(l2Block)
	require.NotNil(t, batch)
	require.Equal(t, l2Block.Timestamp, batch.LastL2BLockTimestamp())
	require.Equal(t, l2Block.BatchNumber, batch.BatchNumber())
	require.Equal(t, l2Block.L1InfotreeIndex, batch.L1InfoTreeIndex())
	require.Equal(t, common.BytesToAddress(l2Block.Coinbase), batch.LastCoinbase())
	require.Equal(t, common.BytesToHash(l2Block.GlobalExitRoot), batch.GlobalExitRoot())
}

func TestBananaBaseNewSequenceBatch(t *testing.T) {
	testData := newBananaBaseTestData(t)
	testData.l1Client.On("HeaderByNumber", mock.Anything, mock.Anything).
		Return(&types.Header{Number: big.NewInt(69)}, nil)
	l2Block := &datastream.L2Block{
		Timestamp:       1,
		BatchNumber:     2,
		L1InfotreeIndex: 3,
		Coinbase:        []byte{1, 2, 3},
		GlobalExitRoot:  []byte{4, 5, 6},
	}
	testData.l1InfoTreeSync.EXPECT().GetInitL1InfoRootMap(mock.Anything).Return(nil, l1infotreesync.ErrNotFound).Once()

	batch := testData.sut.NewBatchFromL2Block(l2Block)
	batches := []seqsendertypes.Batch{batch}
	lastAcc := common.HexToHash("0x8aca9664752dbae36135fd0956c956fc4a370feeac67485b49bcd4b99608ae41")
	testData.rollupContract.EXPECT().LastAccInputHash(mock.Anything).Return(lastAcc, nil)
	l1infoRoot := common.HexToHash("0x66ca9664752dbae36135fd0956c956fc4a370feeac67485b49bcd4b99608ae41")
	testData.l1InfoTreeSync.On("GetLatestInfoUntilBlock", mock.Anything, mock.Anything).
		Return(&l1infotreesync.L1InfoTreeLeaf{L1InfoTreeIndex: 7}, nil)
	testData.getContract.EXPECT().L1InfoRootMap(mock.Anything, uint32(8)).Return(l1infoRoot, nil)

	seq, err := testData.sut.NewSequence(context.TODO(), batches, common.Address{})
	require.NotNil(t, seq)
	require.NoError(t, err)
	// TODO: check that the seq have the right values
}

func TestBananaSanityCheck(t *testing.T) {
	batch := state.BatchRawV2{
		Blocks: []state.L2BlockRaw{
			{
				BlockNumber: 1,
				ChangeL2BlockHeader: state.ChangeL2BlockHeader{
					DeltaTimestamp:  1,
					IndexL1InfoTree: 1,
				},
			},
		},
	}
	data, err := state.EncodeBatchV2(&batch)
	require.NoError(t, err)
	require.NotNil(t, data)
	seq := etherman.SequenceBanana{
		CounterL1InfoRoot: 2,
		Batches: []etherman.Batch{
			{
				L2Data: data,
			},
		},
	}
	err = txbuilder.SequenceSanityCheck(&seq)
	require.NoError(t, err, "inside batchl2data max is 1 and counter is 2 (2>=1+1)")
	seq.CounterL1InfoRoot = 1
	err = txbuilder.SequenceSanityCheck(&seq)
	require.Error(t, err, "inside batchl2data max is 1 and counter is 1. The batchl2data is not included in counter")
}

func TestBananaSanityCheckNilSeq(t *testing.T) {
	err := txbuilder.SequenceSanityCheck(nil)
	require.Error(t, err, "nil sequence")
}

func TestBananaEmptyL1InfoTree(t *testing.T) {
	testData := newBananaBaseTestData(t)

	testData.l1Client.On("HeaderByNumber", mock.Anything, mock.Anything).
		Return(&types.Header{Number: big.NewInt(69)}, nil)
	testData.l1InfoTreeSync.EXPECT().GetLatestInfoUntilBlock(testData.ctx, uint64(69)).Return(nil, l1infotreesync.ErrNotFound)
	testData.l1InfoTreeSync.EXPECT().GetInitL1InfoRootMap(testData.ctx).Return(&l1infotreesync.L1InfoTreeInitial{LeafCount: 10}, nil)

	leafCounter, err := testData.sut.GetCounterL1InfoRoot(testData.ctx, 0)
	require.NoError(t, err)
	require.Equal(t, uint32(10), leafCounter)
}

type testDataBananaBase struct {
	rollupContract *mocks_txbuilder.RollupBananaBaseContractor
	getContract    *mocks_txbuilder.GlobalExitRootBananaContractor
	opts           bind.TransactOpts
	sut            *txbuilder.TxBuilderBananaBase
	l1InfoTreeSync *mocks_txbuilder.L1InfoSyncer
	l1Client       *mocks_txbuilder.L1Client
	ctx            context.Context
}

func newBananaBaseTestData(t *testing.T) *testDataBananaBase {
	t.Helper()

	zkevmContractMock := mocks_txbuilder.NewRollupBananaBaseContractor(t)
	gerContractMock := mocks_txbuilder.NewGlobalExitRootBananaContractor(t)
	opts := bind.TransactOpts{}
	l1Client := mocks_txbuilder.NewL1Client(t)
	l1InfoSyncer := mocks_txbuilder.NewL1InfoSyncer(t)
	sut := txbuilder.NewTxBuilderBananaBase(
		log.GetDefaultLogger(),
		zkevmContractMock,
		gerContractMock,
		l1InfoSyncer, l1Client, big.NewInt(0), opts,
	)
	require.NotNil(t, sut)
	return &testDataBananaBase{
		rollupContract: zkevmContractMock,
		getContract:    gerContractMock,
		opts:           opts,
		sut:            sut,
		l1InfoTreeSync: l1InfoSyncer,
		l1Client:       l1Client,
		ctx:            context.TODO(),
	}
}
