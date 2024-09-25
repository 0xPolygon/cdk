package sequencesender

import (
	"os"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/state"
	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/0xPolygonHermez/zkevm-ethtx-manager/ethtxmanager"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

const (
	txStreamEncoded1 = "f86508843b9aca0082520894617b3a3528f9cdd6630fd3301b9c8911f7bf063d0a808207f5a0579b72a1c1ffdd845fba45317540982109298e2ec8d67ddf2cdaf22e80903677a01831e9a01291c7ea246742a5b5a543ca6938bfc3f6958c22be06fad99274e4ac"
	txStreamEncoded2 = "f86509843b9aca0082520894617b3a3528f9cdd6630fd3301b9c8911f7bf063d0a808207f5a0908a522075e09485166ffa7630cd2b7013897fa1f1238013677d6f0a86efb3d2a0068b12435fcdc8ee254f3b1df8c5b29ed691eeee6065704f061130935976ca99"
	txStreamEncoded3 = "b8b402f8b101268505d21dba0085076c363d8982dc60941929761e87667283f087ea9ab8370c174681b4e980b844095ea7b300000000000000000000000080a64c6d7f12c47b7c66c5b4e20e72bc1fcd5d9effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc001a0dd4db494969139a120e8721842455ec13f82757a4fc49b66d447c7d32d095a1da06ef54068a9aa67ecc4f52d885299a04feb6f3531cdfc771f1412cd3331d1ba4c"
)

func TestStreamTx(t *testing.T) {
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

	printBatch(&batch, true, true)

	encodedBatch, err := state.EncodeBatchV2(&batch)
	require.NoError(t, err)

	decodedBatch, err := state.DecodeBatchV2(encodedBatch)
	require.NoError(t, err)

	printBatch(decodedBatch, true, true)
}

func TestAddNewBatchL2Block(t *testing.T) {
	logger := log.GetDefaultLogger()
	txBuilder := txbuilder.NewTxBuilderBananaZKEVM(logger, nil, nil, bind.TransactOpts{}, 100, nil, nil, nil)
	sut := SequenceSender{
		logger:            logger,
		cfg:               Config{},
		ethTransactions:   make(map[common.Hash]*ethTxData),
		ethTxData:         make(map[common.Hash][]byte),
		sequenceData:      make(map[uint64]*sequenceData),
		validStream:       false,
		latestStreamBatch: 0,
		seqSendingStopped: false,
		TxBuilder:         txBuilder,
	}

	l2Block := datastream.L2Block{
		Number:          1,
		BatchNumber:     1,
		L1InfotreeIndex: 1,
	}
	sut.addNewSequenceBatch(&l2Block)
	l2Block = datastream.L2Block{
		Number:          2,
		BatchNumber:     1,
		L1InfotreeIndex: 0,
	}
	sut.addNewBatchL2Block(&l2Block)
	data := sut.sequenceData[sut.wipBatch]
	// L1InfotreeIndex 0 is ignored
	require.Equal(t, uint32(1), data.batch.L1InfoTreeIndex(), "new block have index=0 and is ignored")

	l2Block = datastream.L2Block{
		Number:          2,
		BatchNumber:     1,
		L1InfotreeIndex: 5,
	}
	sut.addNewBatchL2Block(&l2Block)
	data = sut.sequenceData[sut.wipBatch]
	require.Equal(t, uint32(5), data.batch.L1InfoTreeIndex(), "new block have index=5 and is set")
}

func Test_marginTimeElapsed(t *testing.T) {
	t.Parallel()

	type args struct {
		l2BlockTimestamp uint64
		currentTime      uint64
		timeMargin       int64
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 int64
	}{
		{
			name: "time elapsed",
			args: args{
				l2BlockTimestamp: 100,
				currentTime:      200,
				timeMargin:       50,
			},
			want:  true,
			want1: 0,
		},
		{
			name: "time not elapsed",
			args: args{
				l2BlockTimestamp: 100,
				currentTime:      200,
				timeMargin:       150,
			},
			want:  false,
			want1: 50,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, got1 := marginTimeElapsed(tt.args.l2BlockTimestamp, tt.args.currentTime, tt.args.timeMargin)
			require.Equal(t, tt.want, got, "marginTimeElapsed() got = %v, want %v", got, tt.want)
			require.Equal(t, tt.want1, got1, "marginTimeElapsed() got1 = %v, want %v", got1, tt.want1)
		})
	}
}

func Test_purgeSequences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		seqSendingStopped  bool
		latestVirtualBatch uint64
		sequenceList       []uint64
		sequenceData       map[uint64]*sequenceData

		expectedSequenceList []uint64
		expectedSequenceData map[uint64]*sequenceData
	}{
		{
			name:               "sequences purged when seqSendingStopped",
			seqSendingStopped:  true,
			latestVirtualBatch: 2,
			sequenceList:       []uint64{1, 2},
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
			name:               "no sequences purged",
			seqSendingStopped:  false,
			latestVirtualBatch: 3,
			sequenceList:       []uint64{4, 5},
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
			name:               "sequences purged",
			seqSendingStopped:  false,
			latestVirtualBatch: 5,
			sequenceList:       []uint64{4, 5, 6},
			sequenceData: map[uint64]*sequenceData{
				4: {},
				5: {},
				6: {},
			},
			expectedSequenceList: []uint64{6},
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
				seqSendingStopped:  tt.seqSendingStopped,
				latestVirtualBatch: tt.latestVirtualBatch,
				sequenceList:       tt.sequenceList,
				sequenceData:       tt.sequenceData,
				logger:             log.GetDefaultLogger(),
			}

			ss.purgeSequences()

			require.Equal(t, tt.expectedSequenceList, ss.sequenceList)
			require.Equal(t, tt.expectedSequenceData, ss.sequenceData)
		})
	}
}

func Test_purgeEthTransactions(t *testing.T) {
	t.Parallel()

	firstTimestamp := time.Now().Add(-time.Hour)
	secondTimestamp := time.Now().Add(time.Hour)

	tests := []struct {
		name                    string
		seqSendingStopped       bool
		ethTransactions         map[common.Hash]*ethTxData
		ethTxData               map[common.Hash][]byte
		getEthTxManager         func(t *testing.T) *EthTxMngrMock
		sequenceList            []uint64
		expectedEthTransactions map[common.Hash]*ethTxData
		expectedEthTxData       map[common.Hash][]byte
	}{
		{
			name:              "sequence sender stopped",
			seqSendingStopped: true,
			ethTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {
					StatusTimestamp: firstTimestamp,
					OnMonitor:       true,
					Status:          ethtxmanager.MonitoredTxStatusFinalized.String(),
				},
			},
			ethTxData: map[common.Hash][]byte{
				common.HexToHash("0x1"): []byte{1, 2, 3},
			},
			getEthTxManager: func(t *testing.T) *EthTxMngrMock {
				return NewEthTxMngrMock(t)
			},
			sequenceList: []uint64{1, 2},
			expectedEthTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {
					StatusTimestamp: firstTimestamp,
					OnMonitor:       true,
					Status:          ethtxmanager.MonitoredTxStatusFinalized.String(),
				},
			},
			expectedEthTxData: map[common.Hash][]byte{
				common.HexToHash("0x1"): []byte{1, 2, 3},
			},
		},
		{
			name: "transactions purged",
			ethTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {
					StatusTimestamp: firstTimestamp,
					OnMonitor:       true,
					Status:          ethtxmanager.MonitoredTxStatusFinalized.String(),
				},
				common.HexToHash("0x2"): {
					StatusTimestamp: secondTimestamp,
				},
			},
			ethTxData: map[common.Hash][]byte{
				common.HexToHash("0x1"): []byte{1, 2, 3},
				common.HexToHash("0x2"): []byte{4, 5, 6},
			},
			getEthTxManager: func(t *testing.T) *EthTxMngrMock {
				mngr := NewEthTxMngrMock(t)
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
				common.HexToHash("0x2"): []byte{4, 5, 6},
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
		getEthTxManager func(t *testing.T) *EthTxMngrMock

		expectErr        error
		expectPendingTxs uint64
	}{
		{
			name: "sunccessfully synced",
			ethTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {
					StatusTimestamp: time.Now(),
					OnMonitor:       true,
					Status:          ethtxmanager.MonitoredTxStatusCreated.String(),
				},
			},
			getEthTxManager: func(t *testing.T) *EthTxMngrMock {
				mngr := NewEthTxMngrMock(t)
				mngr.On("Result", mock.Anything, common.HexToHash("0x1")).Return(ethtxmanager.MonitoredTxResult{
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
