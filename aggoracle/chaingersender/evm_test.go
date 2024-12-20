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

	gasOffset := uint64(1000)

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
			ethTxMan.On("Add", ctx, &l2GERManagerAddr, common.Big0, mock.Anything, gasOffset, mock.Anything).Return(tt.addReturnTxID, tt.addReturnErr)
			ethTxMan.On("Result", ctx, tt.addReturnTxID).Return(tt.resultReturn, tt.resultReturnErr)

			sender := &EVMChainGERSender{
				logger:              log.GetDefaultLogger(),
				l2GERManagerAddr:    l2GERManagerAddr,
				l2GERManagerAbi:     &l2GERManagerAbi,
				ethTxMan:            ethTxMan,
				gasOffset:           gasOffset,
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
