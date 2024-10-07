package sequencesender

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/mocks"
	"github.com/0xPolygon/zkevm-ethtx-manager/ethtxmanager"
	ethtxtypes "github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_sendTx(t *testing.T) {
	t.Parallel()

	addr := common.BytesToAddress([]byte{1, 2, 3})
	hash := common.HexToHash("0x1")
	oldHash := common.HexToHash("0x2")

	type args struct {
		resend    bool
		txOldHash *common.Hash
		to        *common.Address
		fromBatch uint64
		toBatch   uint64
		data      []byte
		gas       uint64
	}

	type state struct {
		currentNonce        uint64
		ethTxData           map[common.Hash][]byte
		ethTransactions     map[common.Hash]*ethTxData
		latestSentToL1Batch uint64
	}

	tests := []struct {
		name            string
		args            args
		state           state
		getEthTxManager func(t *testing.T) *mocks.EthTxManagerMock
		expectedState   state
		expectedErr     error
	}{
		{
			name: "successfully sent",
			args: args{
				resend:    false,
				txOldHash: nil,
				to:        &addr,
				fromBatch: 1,
				toBatch:   2,
				data:      []byte("test"),
				gas:       100500,
			},
			getEthTxManager: func(t *testing.T) *mocks.EthTxManagerMock {
				t.Helper()

				mngr := mocks.NewEthTxManagerMock(t)
				mngr.On("AddWithGas", mock.Anything, &addr, big.NewInt(0), []byte("test"), uint64(0), mock.Anything, uint64(100500)).Return(hash, nil)
				mngr.On("Result", mock.Anything, hash).Return(ethtxtypes.MonitoredTxResult{
					ID:   hash,
					Data: []byte{1, 2, 3},
				}, nil)
				return mngr
			},
			state: state{
				currentNonce: 10,
				ethTxData: map[common.Hash][]byte{
					hash: {},
				},
				ethTransactions: map[common.Hash]*ethTxData{
					hash: {},
				},
				latestSentToL1Batch: 0,
			},
			expectedState: state{
				currentNonce: 11,
				ethTxData: map[common.Hash][]byte{
					hash: {1, 2, 3},
				},
				ethTransactions: map[common.Hash]*ethTxData{
					hash: {
						SentL1Timestamp: now,
						StatusTimestamp: now,
						FromBatch:       1,
						ToBatch:         2,
						OnMonitor:       true,
						To:              addr,
						Gas:             100500,
						StateHistory:    []string{now.Format("2006-01-02T15:04:05.000-07:00") + ", *new, "},
						Txs:             map[common.Hash]ethTxAdditionalData{},
					},
				},
				latestSentToL1Batch: 2,
			},
			expectedErr: nil,
		},
		{
			name: "successfully sent with resend",
			args: args{
				resend:    true,
				txOldHash: &oldHash,
				gas:       100500,
			},
			getEthTxManager: func(t *testing.T) *mocks.EthTxManagerMock {
				t.Helper()

				mngr := mocks.NewEthTxManagerMock(t)
				mngr.On("AddWithGas", mock.Anything, &addr, big.NewInt(0), []byte(nil), uint64(0), mock.Anything, uint64(100500)).Return(hash, nil)
				mngr.On("Result", mock.Anything, hash).Return(ethtxtypes.MonitoredTxResult{
					ID:   hash,
					Data: []byte{1, 2, 3},
				}, nil)
				return mngr
			},
			state: state{
				ethTxData: map[common.Hash][]byte{
					hash: []byte("test"),
				},
				ethTransactions: map[common.Hash]*ethTxData{
					oldHash: {
						To:        addr,
						Nonce:     10,
						FromBatch: 1,
						ToBatch:   2,
					},
				},
				latestSentToL1Batch: 0,
			},
			expectedState: state{
				currentNonce: 0,
				ethTxData: map[common.Hash][]byte{
					hash: {1, 2, 3},
				},
				ethTransactions: map[common.Hash]*ethTxData{
					hash: {
						SentL1Timestamp: now,
						StatusTimestamp: now,
						FromBatch:       1,
						ToBatch:         2,
						OnMonitor:       true,
						To:              addr,
						Gas:             100500,
						StateHistory:    []string{now.Format("2006-01-02T15:04:05.000-07:00") + ", *new, "},
						Txs:             map[common.Hash]ethTxAdditionalData{},
					},
				},
				latestSentToL1Batch: 0,
			},
			expectedErr: nil,
		},
		{
			name: "add with gas returns error",
			args: args{
				resend:    true,
				txOldHash: &oldHash,
				gas:       100500,
			},
			getEthTxManager: func(t *testing.T) *mocks.EthTxManagerMock {
				t.Helper()

				mngr := mocks.NewEthTxManagerMock(t)
				mngr.On("AddWithGas", mock.Anything, &addr, big.NewInt(0), []byte(nil), uint64(0), mock.Anything, uint64(100500)).Return(nil, errors.New("failed to add with gas"))
				return mngr
			},
			state: state{
				ethTxData: map[common.Hash][]byte{
					hash: []byte("test"),
				},
				ethTransactions: map[common.Hash]*ethTxData{
					oldHash: {
						To:        addr,
						Nonce:     10,
						FromBatch: 1,
						ToBatch:   2,
					},
				},
				latestSentToL1Batch: 0,
			},
			expectedErr: errors.New("failed to add with gas"),
		},
		{
			name: "empty old hash",
			args: args{
				resend: true,
				gas:    100500,
			},
			getEthTxManager: func(t *testing.T) *mocks.EthTxManagerMock {
				t.Helper()

				mngr := mocks.NewEthTxManagerMock(t)
				return mngr
			},
			state: state{
				ethTxData: map[common.Hash][]byte{
					hash: []byte("test"),
				},
				ethTransactions: map[common.Hash]*ethTxData{
					oldHash: {
						To:        addr,
						Nonce:     10,
						FromBatch: 1,
						ToBatch:   2,
					},
				},
				latestSentToL1Batch: 0,
			},
			expectedErr: errors.New("resend tx with nil hash monitor id"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpFile, err := os.CreateTemp(os.TempDir(), tt.name+".tmp")
			require.NoError(t, err)
			defer os.RemoveAll(tmpFile.Name() + ".tmp")

			ss := SequenceSender{
				ethTxData:           tt.state.ethTxData,
				ethTransactions:     tt.state.ethTransactions,
				ethTxManager:        tt.getEthTxManager(t),
				latestSentToL1Batch: tt.state.latestSentToL1Batch,
				cfg: Config{
					SequencesTxFileName: tmpFile.Name() + ".tmp",
				},
				logger: log.GetDefaultLogger(),
			}

			err = ss.sendTx(context.Background(), tt.args.resend, tt.args.txOldHash, tt.args.to, tt.args.fromBatch, tt.args.toBatch, tt.args.data, tt.args.gas)
			if tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedState.ethTxData, ss.ethTxData)
				require.Equal(t, len(tt.expectedState.ethTransactions), len(ss.ethTransactions))
				for k, v := range tt.expectedState.ethTransactions {
					require.Equal(t, v.Gas, ss.ethTransactions[k].Gas)
					require.Equal(t, v.To, ss.ethTransactions[k].To)
					require.Equal(t, v.Nonce, ss.ethTransactions[k].Nonce)
					require.Equal(t, v.Status, ss.ethTransactions[k].Status)
					require.Equal(t, v.FromBatch, ss.ethTransactions[k].FromBatch)
					require.Equal(t, v.ToBatch, ss.ethTransactions[k].ToBatch)
					require.Equal(t, v.OnMonitor, ss.ethTransactions[k].OnMonitor)
				}
				require.Equal(t, tt.expectedState.latestSentToL1Batch, ss.latestSentToL1Batch)
			}
		})
	}
}

func Test_purgeEthTx(t *testing.T) {
	t.Parallel()

	firstTimestamp := time.Now().Add(-time.Hour)
	secondTimestamp := time.Now().Add(time.Hour)

	tests := []struct {
		name                    string
		seqSendingStopped       uint32
		ethTransactions         map[common.Hash]*ethTxData
		ethTxData               map[common.Hash][]byte
		getEthTxManager         func(t *testing.T) *mocks.EthTxManagerMock
		sequenceList            []uint64
		expectedEthTransactions map[common.Hash]*ethTxData
		expectedEthTxData       map[common.Hash][]byte
	}{
		{
			name:              "sequence sender stopped",
			seqSendingStopped: 1,
			ethTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {
					StatusTimestamp: firstTimestamp,
					OnMonitor:       true,
					Status:          ethtxtypes.MonitoredTxStatusFinalized.String(),
				},
			},
			ethTxData: map[common.Hash][]byte{
				common.HexToHash("0x1"): {1, 2, 3},
			},
			getEthTxManager: func(t *testing.T) *mocks.EthTxManagerMock {
				t.Helper()

				return mocks.NewEthTxManagerMock(t)
			},
			sequenceList: []uint64{1, 2},
			expectedEthTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {
					StatusTimestamp: firstTimestamp,
					OnMonitor:       true,
					Status:          ethtxtypes.MonitoredTxStatusFinalized.String(),
				},
			},
			expectedEthTxData: map[common.Hash][]byte{
				common.HexToHash("0x1"): {1, 2, 3},
			},
		},
		{
			name: "transactions purged",
			ethTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {
					StatusTimestamp: firstTimestamp,
					OnMonitor:       true,
					Status:          ethtxtypes.MonitoredTxStatusFinalized.String(),
				},
				common.HexToHash("0x2"): {
					StatusTimestamp: secondTimestamp,
				},
			},
			ethTxData: map[common.Hash][]byte{
				common.HexToHash("0x1"): {1, 2, 3},
				common.HexToHash("0x2"): {4, 5, 6},
			},
			getEthTxManager: func(t *testing.T) *mocks.EthTxManagerMock {
				t.Helper()

				mngr := mocks.NewEthTxManagerMock(t)
				mngr.On("Remove", mock.Anything, common.HexToHash("0x1")).Return(nil)
				return mngr
			},
			sequenceList: []uint64{1, 2},
			expectedEthTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x2"): {
					StatusTimestamp: secondTimestamp,
				},
			},
			expectedEthTxData: map[common.Hash][]byte{
				common.HexToHash("0x2"): {4, 5, 6},
			},
		},
		{
			name: "removed with error",
			ethTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {
					StatusTimestamp: firstTimestamp,
					OnMonitor:       true,
					Status:          ethtxtypes.MonitoredTxStatusFinalized.String(),
				},
				common.HexToHash("0x2"): {
					StatusTimestamp: secondTimestamp,
				},
			},
			ethTxData: map[common.Hash][]byte{
				common.HexToHash("0x1"): {1, 2, 3},
				common.HexToHash("0x2"): {4, 5, 6},
			},
			getEthTxManager: func(t *testing.T) *mocks.EthTxManagerMock {
				t.Helper()

				mngr := mocks.NewEthTxManagerMock(t)
				mngr.On("Remove", mock.Anything, common.HexToHash("0x1")).Return(errors.New("test err"))
				return mngr
			},
			sequenceList: []uint64{1, 2},
			expectedEthTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x2"): {
					StatusTimestamp: secondTimestamp,
				},
			},
			expectedEthTxData: map[common.Hash][]byte{
				common.HexToHash("0x2"): {4, 5, 6},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mngr := tt.getEthTxManager(t)
			ss := SequenceSender{
				seqSendingStopped: tt.seqSendingStopped,
				ethTransactions:   tt.ethTransactions,
				ethTxData:         tt.ethTxData,
				ethTxManager:      mngr,
				logger:            log.GetDefaultLogger(),
			}

			ss.purgeEthTx(context.Background())

			mngr.AssertExpectations(t)
			require.Equal(t, tt.expectedEthTransactions, ss.ethTransactions)
			require.Equal(t, tt.expectedEthTxData, ss.ethTxData)
		})
	}
}

func Test_syncEthTxResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		ethTransactions map[common.Hash]*ethTxData
		getEthTxManager func(t *testing.T) *mocks.EthTxManagerMock

		expectErr        error
		expectPendingTxs uint64
	}{
		{
			name: "successfully synced",
			ethTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {
					StatusTimestamp: time.Now(),
					OnMonitor:       true,
					Status:          ethtxtypes.MonitoredTxStatusCreated.String(),
				},
			},
			getEthTxManager: func(t *testing.T) *mocks.EthTxManagerMock {
				t.Helper()

				mngr := mocks.NewEthTxManagerMock(t)
				mngr.On("Result", mock.Anything, common.HexToHash("0x1")).Return(ethtxtypes.MonitoredTxResult{
					ID:   common.HexToHash("0x1"),
					Data: []byte{1, 2, 3},
				}, nil)
				return mngr
			},
			expectPendingTxs: 1,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpFile, err := os.CreateTemp(os.TempDir(), tt.name)
			require.NoError(t, err)

			mngr := tt.getEthTxManager(t)
			ss := SequenceSender{
				ethTransactions: tt.ethTransactions,
				ethTxManager:    mngr,
				ethTxData:       make(map[common.Hash][]byte),
				cfg: Config{
					SequencesTxFileName: tmpFile.Name() + ".tmp",
				},
				logger: log.GetDefaultLogger(),
			}

			pendingTxs, err := ss.syncEthTxResults(context.Background())
			if tt.expectErr != nil {
				require.Equal(t, tt.expectErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectPendingTxs, pendingTxs)
			}

			mngr.AssertExpectations(t)

			err = os.RemoveAll(tmpFile.Name() + ".tmp")
			require.NoError(t, err)
		})
	}
}

func Test_syncAllEthTxResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		ethTransactions map[common.Hash]*ethTxData
		getEthTxManager func(t *testing.T) *mocks.EthTxManagerMock

		expectErr        error
		expectPendingTxs uint64
	}{
		{
			name: "successfully synced",
			ethTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {
					StatusTimestamp: time.Now(),
					OnMonitor:       true,
					Status:          ethtxtypes.MonitoredTxStatusCreated.String(),
				},
			},
			getEthTxManager: func(t *testing.T) *mocks.EthTxManagerMock {
				t.Helper()

				mngr := mocks.NewEthTxManagerMock(t)
				mngr.On("ResultsByStatus", mock.Anything, []ethtxtypes.MonitoredTxStatus(nil)).Return([]ethtxtypes.MonitoredTxResult{
					{
						ID:   common.HexToHash("0x1"),
						Data: []byte{1, 2, 3},
					},
				}, nil)
				return mngr
			},
			expectPendingTxs: 1,
		},
		{
			name:            "successfully synced with missing tx",
			ethTransactions: map[common.Hash]*ethTxData{},
			getEthTxManager: func(t *testing.T) *mocks.EthTxManagerMock {
				t.Helper()

				mngr := mocks.NewEthTxManagerMock(t)
				mngr.On("ResultsByStatus", mock.Anything, []ethtxtypes.MonitoredTxStatus(nil)).Return([]ethtxtypes.MonitoredTxResult{
					{
						ID:   common.HexToHash("0x1"),
						Data: []byte{1, 2, 3},
					},
				}, nil)
				return mngr
			},
			expectPendingTxs: 1,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpFile, err := os.CreateTemp(os.TempDir(), tt.name)
			require.NoError(t, err)

			mngr := tt.getEthTxManager(t)
			ss := SequenceSender{
				ethTransactions: tt.ethTransactions,
				ethTxManager:    mngr,
				ethTxData:       make(map[common.Hash][]byte),
				cfg: Config{
					SequencesTxFileName: tmpFile.Name() + ".tmp",
				},
				logger: log.GetDefaultLogger(),
			}

			err = ss.syncAllEthTxResults(context.Background())
			if tt.expectErr != nil {
				require.Equal(t, tt.expectErr, err)
			} else {
				require.NoError(t, err)
			}

			mngr.AssertExpectations(t)

			err = os.RemoveAll(tmpFile.Name() + ".tmp")
			require.NoError(t, err)
		})
	}
}

func Test_copyTxData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                    string
		txHash                  common.Hash
		txData                  []byte
		txsResults              map[common.Hash]ethtxtypes.TxResult
		ethTxData               map[common.Hash][]byte
		ethTransactions         map[common.Hash]*ethTxData
		expectedRthTxData       map[common.Hash][]byte
		expectedEthTransactions map[common.Hash]*ethTxData
	}{
		{
			name:   "successfully copied",
			txHash: common.HexToHash("0x1"),
			txData: []byte{1, 2, 3},
			txsResults: map[common.Hash]ethtxtypes.TxResult{
				common.HexToHash("0x1"): {},
			},
			ethTxData: map[common.Hash][]byte{
				common.HexToHash("0x1"): {0, 2, 3},
			},
			ethTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {},
			},
			expectedRthTxData: map[common.Hash][]byte{
				common.HexToHash("0x1"): {1, 2, 3},
			},
			expectedEthTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {
					Txs: map[common.Hash]ethTxAdditionalData{
						common.HexToHash("0x1"): {},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := SequenceSender{
				ethTxData:       tt.ethTxData,
				ethTransactions: tt.ethTransactions,
			}

			s.copyTxData(tt.txHash, tt.txData, tt.txsResults)
			require.Equal(t, tt.expectedRthTxData, s.ethTxData)
			require.Equal(t, tt.expectedEthTransactions, s.ethTransactions)
		})
	}
}

func Test_getResultAndUpdateEthTx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		hash            common.Hash
		ethTransactions map[common.Hash]*ethTxData
		ethTxData       map[common.Hash][]byte
		getEthTxManager func(t *testing.T) *mocks.EthTxManagerMock
		expectedErr     error
	}{
		{
			name: "successfully updated",
			hash: common.HexToHash("0x1"),
			ethTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {},
			},
			ethTxData: map[common.Hash][]byte{
				common.HexToHash("0x1"): {},
			},
			getEthTxManager: func(t *testing.T) *mocks.EthTxManagerMock {
				t.Helper()

				mngr := mocks.NewEthTxManagerMock(t)
				mngr.On("Result", mock.Anything, common.HexToHash("0x1")).Return(ethtxtypes.MonitoredTxResult{
					ID:   common.HexToHash("0x1"),
					Data: []byte{1, 2, 3},
				}, nil)
				return mngr
			},
			expectedErr: nil,
		},
		{
			name: "not found",
			hash: common.HexToHash("0x1"),
			ethTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {
					Gas: 100500,
				},
			},
			ethTxData: map[common.Hash][]byte{
				common.HexToHash("0x1"): {},
			},
			getEthTxManager: func(t *testing.T) *mocks.EthTxManagerMock {
				t.Helper()

				mngr := mocks.NewEthTxManagerMock(t)
				mngr.On("Result", mock.Anything, common.HexToHash("0x1")).Return(ethtxtypes.MonitoredTxResult{}, ethtxmanager.ErrNotFound)
				mngr.On("AddWithGas", mock.Anything, mock.Anything, big.NewInt(0), mock.Anything, mock.Anything, mock.Anything, uint64(100500)).Return(common.Hash{}, nil)
				mngr.On("Result", mock.Anything, common.Hash{}).Return(ethtxtypes.MonitoredTxResult{
					ID:   common.HexToHash("0x1"),
					Data: []byte{1, 2, 3},
				}, nil)
				return mngr
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpFile, err := os.CreateTemp(os.TempDir(), tt.name+".tmp")
			require.NoError(t, err)
			defer os.RemoveAll(tmpFile.Name() + ".tmp")

			ss := SequenceSender{
				ethTransactions: tt.ethTransactions,
				ethTxData:       tt.ethTxData,
				ethTxManager:    tt.getEthTxManager(t),
				cfg: Config{
					SequencesTxFileName: tmpFile.Name() + ".tmp",
				},
				logger: log.GetDefaultLogger(),
			}

			err = ss.getResultAndUpdateEthTx(context.Background(), tt.hash)
			if tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_loadSentSequencesTransactions(t *testing.T) {
	t.Parallel()

	tx := &ethTxData{
		FromBatch:    1,
		ToBatch:      2,
		OnMonitor:    true,
		To:           common.BytesToAddress([]byte{1, 2, 3}),
		Gas:          100500,
		StateHistory: []string{"2021-09-01T15:04:05.000-07:00, *new, "},
		Txs:          map[common.Hash]ethTxAdditionalData{},
	}

	tests := []struct {
		name                  string
		getFilename           func(t *testing.T) string
		expectEthTransactions map[common.Hash]*ethTxData
		expectErr             error
	}{
		{
			name: "successfully loaded",
			getFilename: func(t *testing.T) string {
				t.Helper()

				tmpFile, err := os.CreateTemp(os.TempDir(), "test")
				require.NoError(t, err)

				ethTxDataBytes, err := json.Marshal(map[common.Hash]*ethTxData{
					common.HexToHash("0x1"): tx,
				})
				require.NoError(t, err)

				_, err = tmpFile.Write(ethTxDataBytes)
				require.NoError(t, err)

				t.Cleanup(func() {
					err := os.Remove(tmpFile.Name())
					require.NoError(t, err)
				})

				return tmpFile.Name()
			},
			expectEthTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): tx,
			},
		},
		{
			name: "file does not exist",
			getFilename: func(t *testing.T) string {
				t.Helper()

				return "does not exist.tmp"
			},
			expectEthTransactions: map[common.Hash]*ethTxData{},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := SequenceSender{
				cfg: Config{
					SequencesTxFileName: tt.getFilename(t),
				},
				ethTransactions: map[common.Hash]*ethTxData{},
				logger:          log.GetDefaultLogger(),
			}

			err := s.loadSentSequencesTransactions()
			if tt.expectErr != nil {
				require.Equal(t, tt.expectErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectEthTransactions, s.ethTransactions)
			}
		})
	}
}
