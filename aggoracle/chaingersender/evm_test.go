package chaingersender

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/aggoracle/mocks"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEVMChainGERSender_InjectGER(t *testing.T) {
	insertGERFuncABI := `[{
		"inputs": [
			{
				"internalType": "bytes32",
				"name": "_newRoot",
				"type": "bytes32"
			}
		],
		"name": "insertGlobalExitRoot",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}]`
	l2GERManagerAddr := common.HexToAddress("0x123")
	l2GERManagerAbi, err := abi.JSON(strings.NewReader(insertGERFuncABI))
	require.NoError(t, err)

	ger := common.HexToHash("0x456")
	txID := common.HexToHash("0x789")

	tests := []struct {
		name            string
		addReturnTxID   common.Hash
		addReturnErr    error
		resultReturn    types.MonitoredTxResult
		resultReturnErr error
		expectedErr     string
	}{
		{
			name:            "successful injection",
			addReturnTxID:   txID,
			addReturnErr:    nil,
			resultReturn:    types.MonitoredTxResult{Status: types.MonitoredTxStatusMined, MinedAtBlockNumber: big.NewInt(123)},
			resultReturnErr: nil,
			expectedErr:     "",
		},
		{
			name:            "injection fails due to transaction failure",
			addReturnTxID:   txID,
			addReturnErr:    nil,
			resultReturn:    types.MonitoredTxResult{Status: types.MonitoredTxStatusFailed},
			resultReturnErr: nil,
			expectedErr:     "inject GER tx",
		},
		{
			name:            "injection fails due to Add method error",
			addReturnTxID:   common.Hash{},
			addReturnErr:    errors.New("add error"),
			resultReturn:    types.MonitoredTxResult{},
			resultReturnErr: nil,
			expectedErr:     "add error",
		},
		{
			name:            "injection fails due to Result method error",
			addReturnTxID:   txID,
			addReturnErr:    nil,
			resultReturn:    types.MonitoredTxResult{},
			resultReturnErr: errors.New("result error"),
			expectedErr:     "result error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancelFn := context.WithTimeout(context.Background(), time.Millisecond*500)
			defer cancelFn()

			ethTxMan := new(mocks.EthTxManagerMock)
			ethTxMan.
				On("Add", ctx, &l2GERManagerAddr, common.Big0, mock.Anything, mock.Anything, mock.Anything).
				Return(tt.addReturnTxID, tt.addReturnErr)
			ethTxMan.
				On("Result", ctx, tt.addReturnTxID).
				Return(tt.resultReturn, tt.resultReturnErr)

			sender := &EVMChainGERSender{
				logger:              log.GetDefaultLogger(),
				l2GERManagerAddr:    l2GERManagerAddr,
				l2GERManagerAbi:     &l2GERManagerAbi,
				ethTxMan:            ethTxMan,
				waitPeriodMonitorTx: time.Millisecond * 10,
			}

			err := sender.InjectGER(ctx, ger)
			if tt.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
			}
		})
	}
}

func TestEVMChainGERSender_IsGERInjected(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     *big.Int
		mockError      error
		expectedResult bool
		expectedErrMsg string
	}{
		{
			name:           "GER is injected",
			mockReturn:     big.NewInt(1),
			mockError:      nil,
			expectedResult: true,
			expectedErrMsg: "",
		},
		{
			name:           "GER is not injected",
			mockReturn:     big.NewInt(0),
			mockError:      nil,
			expectedResult: false,
			expectedErrMsg: "",
		},
		{
			name:           "Error checking GER injection",
			mockReturn:     nil,
			mockError:      errors.New("some error"),
			expectedResult: false,
			expectedErrMsg: "failed to check if global exit root is injected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockL2GERManager := new(mocks.L2GERManagerMock)
			mockL2GERManager.On("GlobalExitRootMap", mock.Anything, mock.Anything).
				Return(tt.mockReturn, tt.mockError)

			evmChainGERSender := &EVMChainGERSender{
				l2GERManager: mockL2GERManager,
			}

			ger := common.HexToHash("0x12345")
			result, err := evmChainGERSender.IsGERInjected(ger)
			if tt.expectedErrMsg != "" {
				require.ErrorContains(t, err, tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expectedResult, result)

			mockL2GERManager.AssertExpectations(t)
		})
	}
}
