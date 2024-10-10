package txbuilder_test

import (
	"context"
	"fmt"
	"math/big"
	"strings"
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

func TestBananaZkevmName(t *testing.T) {
	testData := newBananaZKEVMTestData(t, txbuilder.MaxTxSizeForL1Disabled)
	require.True(t, strings.Contains(testData.sut.String(), "Banana"))
	require.True(t, strings.Contains(testData.sut.String(), "ZKEVM"))
}

func TestBananaZkevmNewSequenceIfWorthToSend(t *testing.T) {
	testData := newBananaZKEVMTestData(t, txbuilder.MaxTxSizeForL1Disabled)

	testSequenceIfWorthToSendNoNewSeq(t, testData.sut)
	testSequenceIfWorthToSendErr(t, testData.sut)
	testSetCondNewSeq(t, testData.sut)
}

func TestBananaZkevmBuildSequenceBatchesTxOk(t *testing.T) {
	testData := newBananaZKEVMTestData(t, txbuilder.MaxTxSizeForL1Disabled)
	testData.l1Client.On("HeaderByNumber", mock.Anything, mock.Anything).
		Return(&types.Header{Number: big.NewInt(69)}, nil)
	testData.l1InfoTreeSync.On("GetLatestInfoUntilBlock", mock.Anything, mock.Anything).
		Return(&l1infotreesync.L1InfoTreeLeaf{L1InfoTreeIndex: 7}, nil)
	testData.l1InfoTreeSync.EXPECT().GetInitL1InfoRootMap(mock.Anything).Return(nil, nil)

	seq, err := newSequenceBananaZKEVMForTest(testData)
	require.NoError(t, err)

	inner := &types.LegacyTx{}
	tx := types.NewTx(inner)

	// It check that SequenceBatches is not going to be send
	testData.rollupContract.EXPECT().SequenceBatches(mock.MatchedBy(func(opts *bind.TransactOpts) bool {
		return opts.NoSend == true
	}), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tx, nil).Once()
	returnTx, err := testData.sut.BuildSequenceBatchesTx(context.TODO(), seq)
	require.NoError(t, err)
	require.Equal(t, tx, returnTx)
}

func TestBananaZkevmBuildSequenceBatchesTxErr(t *testing.T) {
	testData := newBananaZKEVMTestData(t, txbuilder.MaxTxSizeForL1Disabled)
	testData.l1Client.On("HeaderByNumber", mock.Anything, mock.Anything).
		Return(&types.Header{Number: big.NewInt(69)}, nil)
	testData.l1InfoTreeSync.On("GetLatestInfoUntilBlock", mock.Anything, mock.Anything).
		Return(&l1infotreesync.L1InfoTreeLeaf{L1InfoTreeIndex: 7}, nil)
	testData.l1InfoTreeSync.EXPECT().GetInitL1InfoRootMap(mock.Anything).Return(nil, nil)

	seq, err := newSequenceBananaZKEVMForTest(testData)
	require.NoError(t, err)

	err = fmt.Errorf("test-error")
	testData.rollupContract.EXPECT().SequenceBatches(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, err).Once()
	returnedTx, returnedErr := testData.sut.BuildSequenceBatchesTx(context.TODO(), seq)
	require.ErrorContains(t, returnedErr, err.Error())
	require.Nil(t, returnedTx)
}

type testDataBananaZKEVM struct {
	rollupContract *mocks_txbuilder.RollupBananaZKEVMContractor
	getContract    *mocks_txbuilder.GlobalExitRootBananaContractor
	cond           *mocks_txbuilder.CondNewSequence
	opts           bind.TransactOpts
	sut            *txbuilder.TxBuilderBananaZKEVM
	l1InfoTreeSync *mocks_txbuilder.L1InfoSyncer
	l1Client       *mocks_txbuilder.L1Client
}

func newBananaZKEVMTestData(t *testing.T, maxTxSizeForL1 uint64) *testDataBananaZKEVM {
	t.Helper()

	zkevmContractMock := mocks_txbuilder.NewRollupBananaZKEVMContractor(t)
	gerContractMock := mocks_txbuilder.NewGlobalExitRootBananaContractor(t)
	condMock := mocks_txbuilder.NewCondNewSequence(t)
	opts := bind.TransactOpts{}
	l1Client := mocks_txbuilder.NewL1Client(t)
	l1InfoSyncer := mocks_txbuilder.NewL1InfoSyncer(t)
	sut := txbuilder.NewTxBuilderBananaZKEVM(
		log.GetDefaultLogger(),
		zkevmContractMock,
		gerContractMock,
		opts,
		maxTxSizeForL1,
		l1InfoSyncer,
		l1Client,
		big.NewInt(0),
	)
	require.NotNil(t, sut)
	sut.SetCondNewSeq(condMock)
	return &testDataBananaZKEVM{
		rollupContract: zkevmContractMock,
		getContract:    gerContractMock,
		cond:           condMock,
		opts:           opts,
		sut:            sut,
		l1InfoTreeSync: l1InfoSyncer,
		l1Client:       l1Client,
	}
}

func newSequenceBananaZKEVMForTest(testData *testDataBananaZKEVM) (seqsendertypes.Sequence, error) {
	l2Block := &datastream.L2Block{
		Timestamp:       1,
		BatchNumber:     1,
		L1InfotreeIndex: 3,
		Coinbase:        []byte{1, 2, 3},
		GlobalExitRoot:  []byte{4, 5, 6},
	}
	batch := testData.sut.NewBatchFromL2Block(l2Block)
	batches := []seqsendertypes.Batch{
		batch,
	}
	lastAcc := common.HexToHash("0x8aca9664752dbae36135fd0956c956fc4a370feeac67485b49bcd4b99608ae41")
	testData.rollupContract.EXPECT().LastAccInputHash(mock.Anything).Return(lastAcc, nil).Once()
	l1infoRoot := common.HexToHash("0x66ca9664752dbae36135fd0956c956fc4a370feeac67485b49bcd4b99608ae41")
	testData.getContract.EXPECT().L1InfoRootMap(mock.Anything, uint32(8)).Return(l1infoRoot, nil).Once()
	return testData.sut.NewSequence(context.TODO(), batches, common.Address{})
}
