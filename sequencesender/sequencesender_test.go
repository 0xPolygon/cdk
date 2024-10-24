package sequencesender

import (
	"errors"
	"math/big"
	"os"
	"testing"
	"time"

	types2 "github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	rpctypes "github.com/0xPolygon/cdk/rpc/types"
	"github.com/0xPolygon/cdk/sequencesender/mocks"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/state"
	ethtxtypes "github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

const (
	txStreamEncoded1 = "f86508843b9aca0082520894617b3a3528f9cdd6630fd3301b9c8911f7bf063d0a808207f5a0579b72a1c1ffdd845fba45317540982109298e2ec8d67ddf2cdaf22e80903677a01831e9a01291c7ea246742a5b5a543ca6938bfc3f6958c22be06fad99274e4ac"
	txStreamEncoded2 = "f86509843b9aca0082520894617b3a3528f9cdd6630fd3301b9c8911f7bf063d0a808207f5a0908a522075e09485166ffa7630cd2b7013897fa1f1238013677d6f0a86efb3d2a0068b12435fcdc8ee254f3b1df8c5b29ed691eeee6065704f061130935976ca99"
	txStreamEncoded3 = "b8b402f8b101268505d21dba0085076c363d8982dc60941929761e87667283f087ea9ab8370c174681b4e980b844095ea7b300000000000000000000000080a64c6d7f12c47b7c66c5b4e20e72bc1fcd5d9effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc001a0dd4db494969139a120e8721842455ec13f82757a4fc49b66d447c7d32d095a1da06ef54068a9aa67ecc4f52d885299a04feb6f3531cdfc771f1412cd3331d1ba4c"
)

var (
	now = time.Now()
)

func TestMain(t *testing.M) {
	t.Run()
}

func Test_encoding(t *testing.T) {
	tx1, err := state.DecodeTx(txStreamEncoded1)
	require.NoError(t, err)
	tx2, err := state.DecodeTx(txStreamEncoded2)
	require.NoError(t, err)
	tx3, err := state.DecodeTx(txStreamEncoded3)
	require.NoError(t, err)

	txTest := state.L2TxRaw{
		EfficiencyPercentage: 129,
		TxAlreadyEncoded:     false,
		Tx:                   tx1,
	}
	txTestEncoded := make([]byte, 0)
	txTestEncoded, err = txTest.Encode(txTestEncoded)
	require.NoError(t, err)
	log.Debugf("%s", common.Bytes2Hex(txTestEncoded))

	batch := state.BatchRawV2{
		Blocks: []state.L2BlockRaw{
			{
				ChangeL2BlockHeader: state.ChangeL2BlockHeader{
					DeltaTimestamp:  3633752,
					IndexL1InfoTree: 0,
				},
				Transactions: []state.L2TxRaw{
					{
						EfficiencyPercentage: 129,
						TxAlreadyEncoded:     false,
						Tx:                   tx1,
					},
					{
						EfficiencyPercentage: 97,
						TxAlreadyEncoded:     false,
						Tx:                   tx2,
					},
					{
						EfficiencyPercentage: 97,
						TxAlreadyEncoded:     false,
						Tx:                   tx3,
					},
				},
			},
		},
	}

	encodedBatch, err := state.EncodeBatchV2(&batch)
	require.NoError(t, err)

	decodedBatch, err := state.DecodeBatchV2(encodedBatch)
	require.NoError(t, err)

	require.Equal(t, batch.String(), decodedBatch.String())
}

func Test_Start(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                      string
		getEthTxManager           func(t *testing.T) *mocks.EthTxManagerMock
		getEtherman               func(t *testing.T) *mocks.EthermanMock
		getRPC                    func(t *testing.T) *mocks.RPCInterfaceMock
		batchWaitDuration         types2.Duration
		expectNonce               uint64
		expectLastVirtualBatch    uint64
		expectFromStreamBatch     uint64
		expectWipBatch            uint64
		expectLatestSentToL1Batch uint64
	}{
		{
			name: "successfully started",
			getEtherman: func(t *testing.T) *mocks.EthermanMock {
				t.Helper()

				mngr := mocks.NewEthermanMock(t)
				mngr.On("GetLatestBatchNumber").Return(uint64(1), nil)
				return mngr
			},
			getEthTxManager: func(t *testing.T) *mocks.EthTxManagerMock {
				t.Helper()

				mngr := mocks.NewEthTxManagerMock(t)
				mngr.On("Start").Return(nil)
				mngr.On("ResultsByStatus", mock.Anything, []ethtxtypes.MonitoredTxStatus(nil)).Return(nil, nil)
				return mngr
			},
			getRPC: func(t *testing.T) *mocks.RPCInterfaceMock {
				t.Helper()

				mngr := mocks.NewRPCInterfaceMock(t)
				mngr.On("GetBatch", mock.Anything).Return(&rpctypes.RPCBatch{}, nil)
				return mngr
			},

			batchWaitDuration:         types2.NewDuration(time.Millisecond),
			expectNonce:               3,
			expectLastVirtualBatch:    1,
			expectFromStreamBatch:     1,
			expectWipBatch:            2,
			expectLatestSentToL1Batch: 1,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpFile, err := os.CreateTemp(os.TempDir(), tt.name+".tmp")
			require.NoError(t, err)
			defer os.RemoveAll(tmpFile.Name() + ".tmp")

			s := SequenceSender{
				etherman:     tt.getEtherman(t),
				ethTxManager: tt.getEthTxManager(t),
				cfg: Config{
					SequencesTxFileName:    tmpFile.Name() + ".tmp",
					GetBatchWaitInterval:   tt.batchWaitDuration,
					WaitPeriodSendSequence: types2.NewDuration(1 * time.Millisecond),
				},
				logger:    log.GetDefaultLogger(),
				rpcClient: tt.getRPC(t),
			}

			ctx, cancel := context.WithCancel(context.Background())
			s.Start(ctx)
			time.Sleep(time.Second)
			cancel()
			time.Sleep(time.Second)

			require.Equal(t, tt.expectLatestSentToL1Batch, s.latestSentToL1Batch)
		})
	}
}

func Test_purgeSequences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		seqSendingStopped        uint32
		sequenceList             []uint64
		sequenceData             map[uint64]*sequenceData
		latestVirtualBatchNumber uint64
		expectedSequenceList     []uint64
		expectedSequenceData     map[uint64]*sequenceData
	}{
		{
			name:              "sequences purged when seqSendingStopped",
			seqSendingStopped: 1,
			sequenceList:      []uint64{1, 2},
			sequenceData: map[uint64]*sequenceData{
				1: {},
				2: {},
			},
			expectedSequenceList: []uint64{1, 2},
			expectedSequenceData: map[uint64]*sequenceData{
				1: {},
				2: {},
			},
		},
		{
			name:              "no sequences purged",
			seqSendingStopped: 0,
			sequenceList:      []uint64{4, 5},
			sequenceData: map[uint64]*sequenceData{
				4: {},
				5: {},
			},
			expectedSequenceList: []uint64{4, 5},
			expectedSequenceData: map[uint64]*sequenceData{
				4: {},
				5: {},
			},
		},
		{
			name:              "sequences purged",
			seqSendingStopped: 0,
			sequenceList:      []uint64{4, 5, 6},
			sequenceData: map[uint64]*sequenceData{
				4: {},
				5: {},
				6: {},
			},
			latestVirtualBatchNumber: 5,
			expectedSequenceList:     []uint64{6},
			expectedSequenceData: map[uint64]*sequenceData{
				6: {},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ss := SequenceSender{
				seqSendingStopped:        tt.seqSendingStopped,
				sequenceList:             tt.sequenceList,
				sequenceData:             tt.sequenceData,
				latestVirtualBatchNumber: tt.latestVirtualBatchNumber,
				logger:                   log.GetDefaultLogger(),
			}

			ss.purgeSequences()

			require.Equal(t, tt.expectedSequenceList, ss.sequenceList)
			require.Equal(t, tt.expectedSequenceData, ss.sequenceData)
		})
	}
}

func Test_tryToSendSequence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		getEthTxManager     func(t *testing.T) *mocks.EthTxManagerMock
		getEtherman         func(t *testing.T) *mocks.EthermanMock
		getTxBuilder        func(t *testing.T) *mocks.TxBuilderMock
		maxPendingTxn       uint64
		sequenceList        []uint64
		latestSentToL1Batch uint64
		sequenceData        map[uint64]*sequenceData
		ethTransactions     map[common.Hash]*ethTxData
		ethTxData           map[common.Hash][]byte

		expectErr error
	}{
		{
			name: "successfully sent",
			getEtherman: func(t *testing.T) *mocks.EthermanMock {
				t.Helper()

				mngr := mocks.NewEthermanMock(t)
				mngr.On("GetLatestBatchNumber").Return(uint64(1), nil)
				return mngr
			},
			getEthTxManager: func(t *testing.T) *mocks.EthTxManagerMock {
				t.Helper()

				mngr := mocks.NewEthTxManagerMock(t)
				return mngr
			},
			getTxBuilder: func(t *testing.T) *mocks.TxBuilderMock {
				t.Helper()

				mngr := mocks.NewTxBuilderMock(t)
				mngr.On("NewSequenceIfWorthToSend", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(txbuilder.NewBananaSequence(etherman.SequenceBanana{}), nil)
				return mngr
			},
			maxPendingTxn:       10,
			sequenceList:        []uint64{2},
			latestSentToL1Batch: 1,
			sequenceData: map[uint64]*sequenceData{
				2: {
					batchClosed: true,
					batch:       txbuilder.NewBananaBatch(&etherman.Batch{}),
				},
			},
		},
		{
			name: "successfully sent new sequence",
			getEtherman: func(t *testing.T) *mocks.EthermanMock {
				t.Helper()

				mngr := mocks.NewEthermanMock(t)
				mngr.On("GetLatestBatchNumber").Return(uint64(1), nil)
				mngr.On("GetLatestBlockHeader", mock.Anything).Return(&types.Header{
					Number: big.NewInt(1),
				}, nil)
				mngr.On("EstimateGas", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(uint64(100500), nil)
				return mngr
			},
			getEthTxManager: func(t *testing.T) *mocks.EthTxManagerMock {
				t.Helper()

				mngr := mocks.NewEthTxManagerMock(t)
				mngr.On("AddWithGas", mock.Anything, mock.Anything, big.NewInt(0), mock.Anything, mock.Anything, mock.Anything, uint64(100500)).Return(common.Hash{}, nil)
				mngr.On("Result", mock.Anything, common.Hash{}).Return(ethtxtypes.MonitoredTxResult{
					ID:   common.Hash{},
					Data: []byte{1, 2, 3},
				}, nil)
				return mngr
			},
			getTxBuilder: func(t *testing.T) *mocks.TxBuilderMock {
				t.Helper()

				mngr := mocks.NewTxBuilderMock(t)
				mngr.On("NewSequenceIfWorthToSend", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				mngr.On("NewSequence", mock.Anything, mock.Anything, mock.Anything).Return(txbuilder.NewBananaSequence(etherman.SequenceBanana{
					Batches: []etherman.Batch{{
						BatchNumber: 2,
					}},
				}), nil)
				mngr.On("BuildSequenceBatchesTx", mock.Anything, mock.Anything).Return(types.NewTx(&types.LegacyTx{}), nil)
				return mngr
			},
			maxPendingTxn:       10,
			sequenceList:        []uint64{2},
			latestSentToL1Batch: 1,
			sequenceData: map[uint64]*sequenceData{
				2: {
					batchClosed: true,
					batch:       txbuilder.NewBananaBatch(&etherman.Batch{}),
				},
			},
			ethTransactions: map[common.Hash]*ethTxData{},
			ethTxData:       map[common.Hash][]byte{},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpFile, err := os.CreateTemp(os.TempDir(), tt.name+".tmp")
			require.NoError(t, err)
			defer os.RemoveAll(tmpFile.Name() + ".tmp")

			s := SequenceSender{
				ethTxManager: tt.getEthTxManager(t),
				etherman:     tt.getEtherman(t),
				TxBuilder:    tt.getTxBuilder(t),
				cfg: Config{
					SequencesTxFileName:    tmpFile.Name() + ".tmp",
					MaxPendingTx:           tt.maxPendingTxn,
					WaitPeriodSendSequence: types2.NewDuration(time.Millisecond),
				},
				sequenceList:        tt.sequenceList,
				latestSentToL1Batch: tt.latestSentToL1Batch,
				sequenceData:        tt.sequenceData,
				ethTransactions:     tt.ethTransactions,
				ethTxData:           tt.ethTxData,
				logger:              log.GetDefaultLogger(),
			}

			s.tryToSendSequence(context.Background())
		})
	}
}

func Test_getSequencesToSend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		sequenceList           []uint64
		latestSentToL1Batch    uint64
		forkUpgradeBatchNumber uint64
		sequenceData           map[uint64]*sequenceData
		getTxBuilder           func(t *testing.T) *mocks.TxBuilderMock
		expectedSequence       seqsendertypes.Sequence
		expectedErr            error
	}{
		{
			name:                "successfully get sequence",
			sequenceList:        []uint64{2},
			latestSentToL1Batch: 1,
			sequenceData: map[uint64]*sequenceData{
				2: {
					batchClosed: true,
					batch:       txbuilder.NewBananaBatch(&etherman.Batch{}),
				},
			},
			getTxBuilder: func(t *testing.T) *mocks.TxBuilderMock {
				t.Helper()

				mngr := mocks.NewTxBuilderMock(t)
				mngr.On("NewSequenceIfWorthToSend", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(txbuilder.NewBananaSequence(etherman.SequenceBanana{
					Batches: []etherman.Batch{{
						BatchNumber: 2,
					}},
				}), nil)
				return mngr
			},
			expectedSequence: txbuilder.NewBananaSequence(etherman.SequenceBanana{
				Batches: []etherman.Batch{{
					BatchNumber: 2,
				}},
			}),
			expectedErr: nil,
		},
		{
			name:                "different coinbase",
			sequenceList:        []uint64{2, 3},
			latestSentToL1Batch: 1,
			sequenceData: map[uint64]*sequenceData{
				2: {
					batchClosed: true,
					batch:       txbuilder.NewBananaBatch(&etherman.Batch{}),
				},
				3: {
					batchClosed: true,
					batch: txbuilder.NewBananaBatch(&etherman.Batch{
						LastCoinbase: common.HexToAddress("0x2"),
					}),
				},
			},
			getTxBuilder: func(t *testing.T) *mocks.TxBuilderMock {
				t.Helper()

				mngr := mocks.NewTxBuilderMock(t)
				mngr.On("NewSequenceIfWorthToSend", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				mngr.On("NewSequence", mock.Anything, mock.Anything, mock.Anything).Return(txbuilder.NewBananaSequence(etherman.SequenceBanana{
					Batches: []etherman.Batch{{
						BatchNumber: 2,
					}},
				}), nil)
				return mngr
			},
			expectedSequence: txbuilder.NewBananaSequence(etherman.SequenceBanana{
				Batches: []etherman.Batch{{
					BatchNumber: 2,
				}},
			}),
			expectedErr: nil,
		},
		{
			name:                "NewSequenceIfWorthToSend return error",
			sequenceList:        []uint64{2},
			latestSentToL1Batch: 1,
			sequenceData: map[uint64]*sequenceData{
				2: {
					batchClosed: true,
					batch:       txbuilder.NewBananaBatch(&etherman.Batch{}),
				},
			},
			getTxBuilder: func(t *testing.T) *mocks.TxBuilderMock {
				t.Helper()

				mngr := mocks.NewTxBuilderMock(t)
				mngr.On("NewSequenceIfWorthToSend", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("test error"))
				return mngr
			},
			expectedErr: errors.New("test error"),
		},
		{
			name:                   "fork upgrade",
			sequenceList:           []uint64{2},
			latestSentToL1Batch:    1,
			forkUpgradeBatchNumber: 2,
			sequenceData: map[uint64]*sequenceData{
				2: {
					batchClosed: true,
					batch:       txbuilder.NewBananaBatch(&etherman.Batch{}),
				},
			},
			getTxBuilder: func(t *testing.T) *mocks.TxBuilderMock {
				t.Helper()

				mngr := mocks.NewTxBuilderMock(t)
				mngr.On("NewSequenceIfWorthToSend", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				mngr.On("NewSequence", mock.Anything, mock.Anything, mock.Anything).Return(txbuilder.NewBananaSequence(etherman.SequenceBanana{
					Batches: []etherman.Batch{{
						BatchNumber: 2,
					}},
				}), nil)
				return mngr
			},
			expectedSequence: txbuilder.NewBananaSequence(etherman.SequenceBanana{
				Batches: []etherman.Batch{{
					BatchNumber: 2,
				}},
			}),
			expectedErr: nil,
		},
		{
			name:                   "fork upgrade passed",
			sequenceList:           []uint64{2},
			latestSentToL1Batch:    1,
			forkUpgradeBatchNumber: 1,
			sequenceData: map[uint64]*sequenceData{
				2: {
					batchClosed: true,
					batch:       txbuilder.NewBananaBatch(&etherman.Batch{}),
				},
			},
			getTxBuilder: func(t *testing.T) *mocks.TxBuilderMock {
				t.Helper()

				mngr := mocks.NewTxBuilderMock(t)
				return mngr
			},
			expectedErr: errors.New("aborting sequencing process as we reached the batch 2 where a new forkid is applied (upgrade)"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ss := SequenceSender{
				sequenceList:        tt.sequenceList,
				latestSentToL1Batch: tt.latestSentToL1Batch,
				cfg: Config{
					ForkUpgradeBatchNumber: tt.forkUpgradeBatchNumber,
				},
				sequenceData: tt.sequenceData,
				TxBuilder:    tt.getTxBuilder(t),
				logger:       log.GetDefaultLogger(),
			}

			sequence, err := ss.getSequencesToSend(context.Background())
			if tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedSequence, sequence)
			}
		})
	}
}

func Test_marginTimeElapsed(t *testing.T) {
	t.Parallel()

	type args struct {
		l2BlockTimestamp uint64
		currentTime      uint64
		timeMargin       int64
	}
	tests := []struct {
		name              string
		args              args
		expectedIsElapsed bool
		expectedWaitTime  int64
	}{
		{
			name: "time elapsed",
			args: args{
				l2BlockTimestamp: 100,
				currentTime:      200,
				timeMargin:       50,
			},
			expectedIsElapsed: true,
			expectedWaitTime:  0,
		},
		{
			name: "time not elapsed",
			args: args{
				l2BlockTimestamp: 100,
				currentTime:      200,
				timeMargin:       150,
			},
			expectedIsElapsed: false,
			expectedWaitTime:  50,
		},
		{
			name: "l2 block in the future (time margin not enough)",
			args: args{
				l2BlockTimestamp: 300,
				currentTime:      200,
				timeMargin:       50,
			},
			expectedIsElapsed: true,
			expectedWaitTime:  0,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			isElapsed, waitTime := marginTimeElapsed(tt.args.l2BlockTimestamp, tt.args.currentTime, tt.args.timeMargin)
			require.Equal(t, tt.expectedIsElapsed, isElapsed, "marginTimeElapsed() isElapsed = %t, want %t", isElapsed, tt.expectedIsElapsed)
			require.Equal(t, tt.expectedWaitTime, waitTime, "marginTimeElapsed() got1 = %v, want %v", waitTime, tt.expectedWaitTime)
		})
	}
}
