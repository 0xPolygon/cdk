package txbuilder_test

import (
	"strings"
	"testing"

	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder/mocks_txbuilder"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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

type testDataBananaZKEVM struct {
	rollupContract *mocks_txbuilder.RollupBananaZKEVMContractor
	getContract    *mocks_txbuilder.GlobalExitRootBananaContractor
	cond           *mocks_txbuilder.CondNewSequence
	opts           bind.TransactOpts
	sut            *txbuilder.TxBuilderBananaZKEVM
}

func newBananaZKEVMTestData(t *testing.T, maxTxSizeForL1 uint64) *testDataBananaZKEVM {
	zkevmContractMock := mocks_txbuilder.NewRollupBananaZKEVMContractor(t)
	gerContractMock := mocks_txbuilder.NewGlobalExitRootBananaContractor(t)
	condMock := mocks_txbuilder.NewCondNewSequence(t)
	opts := bind.TransactOpts{}
	sut := txbuilder.NewTxBuilderBananaZKEVM(zkevmContractMock, gerContractMock, opts, maxTxSizeForL1)
	require.NotNil(t, sut)
	sut.SetCondNewSeq(condMock)
	return &testDataBananaZKEVM{
		rollupContract: zkevmContractMock,
		getContract:    gerContractMock,
		cond:           condMock,
		opts:           opts,
		sut:            sut,
	}
}
