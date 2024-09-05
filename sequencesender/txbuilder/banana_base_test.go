package txbuilder_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder/mocks_txbuilder"
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
	seq, err := testData.sut.NewSequence(context.TODO(), nil, common.Address{})
	require.NotNil(t, seq)
	require.NoError(t, err)
	// TODO check values
	//require.Equal(t, lastAcc, seq.LastAccInputHash())
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

type testDataBananaBase struct {
	rollupContract *mocks_txbuilder.RollupBananaBaseContractor
	getContract    *mocks_txbuilder.GlobalExitRootBananaContractor
	opts           bind.TransactOpts
	sut            *txbuilder.TxBuilderBananaBase
	l1InfoTreeSync *mocks_txbuilder.L1InfoSyncer
	l1Client       *mocks_txbuilder.L1Client
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
	}
}
