package sequencesender

import (
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/state"
	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/0xPolygonHermez/zkevm-data-streamer/datastreamer"
	"github.com/0xPolygonHermez/zkevm-ethtx-manager/ethtxmanager"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/proto"
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
			name: "successfully synced",
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
func Test_syncAllEthTxResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		ethTransactions map[common.Hash]*ethTxData
		getEthTxManager func(t *testing.T) *EthTxMngrMock

		expectErr        error
		expectPendingTxs uint64
	}{
		{
			name: "successfully synced",
			ethTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {
					StatusTimestamp: time.Now(),
					OnMonitor:       true,
					Status:          ethtxmanager.MonitoredTxStatusCreated.String(),
				},
			},
			getEthTxManager: func(t *testing.T) *EthTxMngrMock {
				mngr := NewEthTxMngrMock(t)
				mngr.On("ResultsByStatus", mock.Anything, []ethtxmanager.MonitoredTxStatus(nil)).Return([]ethtxmanager.MonitoredTxResult{
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
			getEthTxManager: func(t *testing.T) *EthTxMngrMock {
				mngr := NewEthTxMngrMock(t)
				mngr.On("ResultsByStatus", mock.Anything, []ethtxmanager.MonitoredTxStatus(nil)).Return([]ethtxmanager.MonitoredTxResult{
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
		txsResults              map[common.Hash]ethtxmanager.TxResult
		ethTxData               map[common.Hash][]byte
		ethTransactions         map[common.Hash]*ethTxData
		expectedRthTxData       map[common.Hash][]byte
		expectedEthTransactions map[common.Hash]*ethTxData
	}{
		{
			name:   "successfully copied",
			txHash: common.HexToHash("0x1"),
			txData: []byte{1, 2, 3},
			txsResults: map[common.Hash]ethtxmanager.TxResult{
				common.HexToHash("0x1"): {},
			},
			ethTxData: map[common.Hash][]byte{
				common.HexToHash("0x1"): []byte{0, 2, 3},
			},
			ethTransactions: map[common.Hash]*ethTxData{
				common.HexToHash("0x1"): {},
			},
			expectedRthTxData: map[common.Hash][]byte{
				common.HexToHash("0x1"): []byte{1, 2, 3},
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

func Test_getBatchFromRPC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		batch        uint64
		resp         string
		requestErr   error
		expectBlocks int
		expectData   string
		expectErr    error
	}{
		{
			name:         "successfully fetched",
			resp:         `{"jsonrpc":"2.0","id":1,"result":{"blocks":["1", "2", "3"],"batchL2Data":"test"}}`,
			batch:        0,
			expectBlocks: 3,
			expectData:   "test",
			expectErr:    nil,
		},
		{
			name:         "invalid json",
			resp:         `{"jsonrpc":"2.0","id":1,"result":{"blocks":invalid,"batchL2Data":"test"}}`,
			batch:        0,
			expectBlocks: 3,
			expectData:   "test",
			expectErr:    errors.New("invalid character 'i' looking for beginning of value"),
		},
		{
			name:         "wrong json",
			resp:         `{"jsonrpc":"2.0","id":1,"result":{"blocks":"invalid","batchL2Data":"test"}}`,
			batch:        0,
			expectBlocks: 3,
			expectData:   "test",
			expectErr:    errors.New("error unmarshalling the batch number from the response calling zkevm_getBatchByNumber: json: cannot unmarshal string into Go struct field zkEVMBatch.Blocks of type []string"),
		},
		{
			name:         "error in the response",
			resp:         `{"jsonrpc":"2.0","id":1,"result":null,"error":{"code":-32602,"message":"Invalid params"}}`,
			batch:        0,
			expectBlocks: 0,
			expectData:   "",
			expectErr:    errors.New("error in the response calling zkevm_getBatchByNumber: &{-32602 Invalid params <nil>}"),
		},
		{
			name:         "http failed",
			requestErr:   errors.New("failed to fetch"),
			batch:        0,
			expectBlocks: 0,
			expectData:   "",
			expectErr:    errors.New("invalid status code, expected: 200, found: 500"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.requestErr != nil {
					http.Error(w, tt.requestErr.Error(), http.StatusInternalServerError)
					return
				}

				_, _ = w.Write([]byte(tt.resp))
			}))
			defer srv.Close()

			blocks, data, err := getBatchFromRPC(srv.URL, tt.batch)
			if tt.expectErr != nil {
				require.Equal(t, tt.expectErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectBlocks, blocks)
				require.Equal(t, tt.expectData, data)
			}
		})
	}
}

func Test_addNewBatchL2Block(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		wipBatch             uint64
		l2Block              *datastream.L2Block
		sequenceData         map[uint64]*sequenceData
		expectedSequenceData map[uint64]*sequenceData
	}{
		{
			name:     "successfully added",
			wipBatch: 1,
			l2Block: &datastream.L2Block{
				Number:          34,
				BatchNumber:     1,
				Timestamp:       uint64(123),
				DeltaTimestamp:  0,
				MinTimestamp:    0,
				L1Blockhash:     nil,
				L1InfotreeIndex: 14,
				Hash:            []byte{123, 123, 123},
				StateRoot:       nil,
				GlobalExitRoot:  nil,
				Coinbase:        []byte{5, 6, 7},
				BlockGasLimit:   0,
				BlockInfoRoot:   nil,
				Debug:           nil,
			},
			sequenceData: map[uint64]*sequenceData{
				1: {
					batchClosed: true,
					batch: txbuilder.NewBananaBatch(&etherman.Batch{
						L2Data:               nil,
						LastCoinbase:         common.BytesToAddress([]byte{5, 6, 7}),
						ForcedGlobalExitRoot: common.Hash{},
						ForcedBlockHashL1:    common.Hash{},
						ForcedBatchTimestamp: 0,
						BatchNumber:          0,
						L1InfoTreeIndex:      0,
						LastL2BLockTimestamp: 0,
						GlobalExitRoot:       common.Hash{},
					}),
					batchRaw: &state.BatchRawV2{
						Blocks: nil,
					},
					batchType: 0,
				},
			},
			expectedSequenceData: map[uint64]*sequenceData{
				1: {
					batchClosed: true,
					batch: txbuilder.NewBananaBatch(&etherman.Batch{
						L2Data:               nil,
						LastCoinbase:         common.BytesToAddress([]byte{5, 6, 7}),
						ForcedGlobalExitRoot: common.Hash{},
						ForcedBlockHashL1:    common.Hash{},
						ForcedBatchTimestamp: 0,
						BatchNumber:          0,
						L1InfoTreeIndex:      14,
						LastL2BLockTimestamp: 123,
						GlobalExitRoot:       common.Hash{},
					}),
					batchRaw: &state.BatchRawV2{
						Blocks: []state.L2BlockRaw{
							{
								BlockNumber: 0,
								ChangeL2BlockHeader: state.ChangeL2BlockHeader{
									IndexL1InfoTree: 14,
								},
								Transactions: nil,
							},
						},
					},
					batchType: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ss := SequenceSender{
				wipBatch:     tt.wipBatch,
				sequenceData: tt.sequenceData,
				logger:       log.GetDefaultLogger(),
			}

			ss.addNewBatchL2Block(tt.l2Block)
			require.Equal(t, tt.expectedSequenceData, ss.sequenceData)
		})
	}
}

func Test_addInfoSequenceBatchStart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		batch                *datastream.BatchStart
		wipBatch             uint64
		sequenceData         map[uint64]*sequenceData
		expectedSequenceData map[uint64]*sequenceData
	}{
		{
			name: "successfully added",
			batch: &datastream.BatchStart{
				Type:   datastream.BatchType_BATCH_TYPE_FORCED,
				Number: 2,
			},
			wipBatch: 1,
			sequenceData: map[uint64]*sequenceData{
				1: {
					batchType: 1,
					batch: txbuilder.NewBananaBatch(&etherman.Batch{
						BatchNumber: 1,
					}),
				},
			},
			expectedSequenceData: map[uint64]*sequenceData{
				1: {
					batchType: 2,
					batch: txbuilder.NewBananaBatch(&etherman.Batch{
						BatchNumber: 1,
					}),
				},
			},
		},
		{
			name: "batch does not exist",
			batch: &datastream.BatchStart{
				Type:   datastream.BatchType_BATCH_TYPE_FORCED,
				Number: 2,
			},
			wipBatch: 1,
			sequenceData: map[uint64]*sequenceData{
				10: {},
			},
			expectedSequenceData: map[uint64]*sequenceData{
				10: {},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ss := SequenceSender{
				sequenceData: tt.sequenceData,
				wipBatch:     tt.wipBatch,
				logger:       log.GetDefaultLogger(),
			}

			ss.addInfoSequenceBatchStart(tt.batch)
			require.Equal(t, tt.expectedSequenceData, ss.sequenceData)
		})
	}
}

func Test_addNewSequenceBatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		l2Block              *datastream.L2Block
		wipBatch             uint64
		sequenceList         []uint64
		sequenceData         map[uint64]*sequenceData
		getTxBuilder         func(t *testing.T) *TxBuilderMock
		expectedSequenceList []uint64
		expectedSequenceData map[uint64]*sequenceData
	}{
		{
			name: "successfully added new batch",
			l2Block: &datastream.L2Block{
				Number:      1,
				BatchNumber: 2,
			},
			wipBatch:     1,
			sequenceList: []uint64{1},
			sequenceData: map[uint64]*sequenceData{
				1: {},
			},
			getTxBuilder: func(t *testing.T) *TxBuilderMock {
				mngr := NewTxBuilderMock(t)
				mngr.On("NewBatchFromL2Block", mock.Anything).Return(txbuilder.NewBananaBatch(&etherman.Batch{
					BatchNumber: 2,
				}), nil)
				return mngr
			},
			expectedSequenceList: []uint64{1, 2},
			expectedSequenceData: map[uint64]*sequenceData{
				1: {},
				2: {
					batchClosed: false,
					batch: txbuilder.NewBananaBatch(&etherman.Batch{
						BatchNumber: 2,
					}),
					batchRaw: &state.BatchRawV2{},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ss := SequenceSender{
				sequenceData: tt.sequenceData,
				wipBatch:     tt.wipBatch,
				TxBuilder:    tt.getTxBuilder(t),
				logger:       log.GetDefaultLogger(),
			}

			ss.addNewSequenceBatch(tt.l2Block)
			require.Equal(t, tt.expectedSequenceData, ss.sequenceData)
		})
	}
}

func Test_closeSequenceBatch(t *testing.T) {
	t.Parallel()

	batchRaw := &state.BatchRawV2{
		Blocks: []state.L2BlockRaw{{
			BlockNumber:         1,
			ChangeL2BlockHeader: state.ChangeL2BlockHeader{},
			Transactions:        nil,
		}},
	}
	lsData, err := state.EncodeBatchV2(batchRaw)
	require.NoError(t, err)

	tests := []struct {
		name                 string
		wipBatch             uint64
		sequenceData         map[uint64]*sequenceData
		expectedSequenceData map[uint64]*sequenceData
		expectedErr          error
	}{
		{
			name:     "successfully closed",
			wipBatch: 1,
			sequenceData: map[uint64]*sequenceData{
				1: {
					batch: txbuilder.NewBananaBatch(&etherman.Batch{}),
					batchRaw: &state.BatchRawV2{
						Blocks: []state.L2BlockRaw{{
							BlockNumber:         1,
							ChangeL2BlockHeader: state.ChangeL2BlockHeader{},
							Transactions:        nil,
						}},
					},
				},
			},
			expectedSequenceData: map[uint64]*sequenceData{
				1: {
					batchClosed: true,
					batch: txbuilder.NewBananaBatch(&etherman.Batch{
						L2Data: lsData,
					}),
					batchRaw: batchRaw,
				},
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ss := SequenceSender{
				wipBatch:     tt.wipBatch,
				sequenceData: tt.sequenceData,
				logger:       log.GetDefaultLogger(),
			}

			err := ss.closeSequenceBatch()
			if tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedSequenceData, ss.sequenceData)
			}
		})
	}
}

func Test_sendTx(t *testing.T) {
	t.Parallel()

	addr := common.BytesToAddress([]byte{1, 2, 3})
	hash := common.HexToHash("0x1")
	now := time.Now()

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
		getEthTxManager func(t *testing.T) *EthTxMngrMock
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
			getEthTxManager: func(t *testing.T) *EthTxMngrMock {
				mngr := NewEthTxMngrMock(t)
				nonce := uint64(10)
				mngr.On("AddWithGas", mock.Anything, &addr, &nonce, big.NewInt(0), []byte("test"), uint64(0), mock.Anything, uint64(100500)).Return(hash, nil)
				mngr.On("Result", mock.Anything, hash).Return(ethtxmanager.MonitoredTxResult{
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
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpFile, err := os.CreateTemp(os.TempDir(), tt.name)
			require.NoError(t, err)
			defer os.RemoveAll(tmpFile.Name() + ".tmp")

			ss := SequenceSender{
				currentNonce:        tt.state.currentNonce,
				ethTxData:           tt.state.ethTxData,
				ethTransactions:     tt.state.ethTransactions,
				ethTxManager:        tt.getEthTxManager(t),
				latestSentToL1Batch: tt.state.latestSentToL1Batch,
				cfg: Config{
					SequencesTxFileName: tmpFile.Name() + ".tmp",
				},
				logger: log.GetDefaultLogger(),
			}

			getNow = func() time.Time { return now }
			err = ss.sendTx(context.Background(), tt.args.resend, tt.args.txOldHash, tt.args.to, tt.args.fromBatch, tt.args.toBatch, tt.args.data, tt.args.gas)
			getNow = time.Now
			if tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedState.currentNonce, ss.currentNonce)
				require.Equal(t, tt.expectedState.ethTxData, ss.ethTxData)
				require.Equal(t, tt.expectedState.ethTransactions, ss.ethTransactions)
				require.Equal(t, tt.expectedState.latestSentToL1Batch, ss.latestSentToL1Batch)
			}
		})
	}
}

func Test_entryTypeToString(t *testing.T) {
	type args struct {
		entryType datastream.EntryType
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "entry type 1",
			args: args{
				entryType: datastream.EntryType_ENTRY_TYPE_BATCH_START,
			},
			want: "BatchStart",
		},
		{
			name: "entry type 2",
			args: args{
				entryType: datastream.EntryType_ENTRY_TYPE_L2_BLOCK,
			},
			want: "L2Block",
		},
		{
			name: "entry type 3",
			args: args{
				entryType: datastream.EntryType_ENTRY_TYPE_TRANSACTION,
			},
			want: "Transaction",
		},
		{
			name: "entry type 4",
			args: args{
				entryType: datastream.EntryType_ENTRY_TYPE_BATCH_END,
			},
			want: "BatchEnd",
		},
		{
			name: "entry type unexpected",
			args: args{
				entryType: 10,
			},
			want: "10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := entryTypeToString(tt.args.entryType)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_updateLatestVirtualBatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                       string
		getEtherman                func(t *testing.T) *EthermanMock
		latestVirtualBatch         uint64
		expectedLatestVirtualBatch uint64
		expectedErr                error
	}{
		{
			name: "successfully updated",
			getEtherman: func(t *testing.T) *EthermanMock {
				mngr := NewEthermanMock(t)
				mngr.On("GetLatestBatchNumber").Return(uint64(3), nil)
				return mngr
			},
			expectedLatestVirtualBatch: 3,
		},
		{
			name: "etherman returns error",
			getEtherman: func(t *testing.T) *EthermanMock {
				mngr := NewEthermanMock(t)
				mngr.On("GetLatestBatchNumber").Return(uint64(0), errors.New("test error"))
				return mngr
			},
			expectedErr: errors.New("fail to get latest virtual batch"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ss := SequenceSender{
				etherman:           tt.getEtherman(t),
				latestVirtualBatch: tt.latestVirtualBatch,
				logger:             log.GetDefaultLogger(),
			}

			err := ss.updateLatestVirtualBatch()
			if tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedLatestVirtualBatch, ss.latestVirtualBatch)
			}
		})
	}
}

func Test_addNewBlockTx(t *testing.T) {
	t.Parallel()

	tx1, err := state.DecodeTx(txStreamEncoded1)
	require.NoError(t, err)

	tests := []struct {
		name                 string
		l2Tx                 *datastream.Transaction
		wipBatch             uint64
		sequenceData         map[uint64]*sequenceData
		expectedSequenceData map[uint64]*sequenceData
	}{
		{
			name: "successfully added",
			l2Tx: &datastream.Transaction{
				L2BlockNumber: 2,
				ImStateRoot:   []byte{1, 2, 3, 5, 6, 7, 8, 9, 0},
				Encoded:       hexutils.HexToBytes(txStreamEncoded1),
			},
			wipBatch: 1,
			sequenceData: map[uint64]*sequenceData{
				1: {
					batch: txbuilder.NewBananaBatch(&etherman.Batch{
						BatchNumber: 2,
					}),
					batchRaw: &state.BatchRawV2{
						Blocks: []state.L2BlockRaw{{
							BlockNumber:         1,
							ChangeL2BlockHeader: state.ChangeL2BlockHeader{},
							Transactions:        nil,
						}},
					},
				},
			},
			expectedSequenceData: map[uint64]*sequenceData{
				1: {
					batch: txbuilder.NewBananaBatch(&etherman.Batch{
						BatchNumber: 2,
					}),
					batchRaw: &state.BatchRawV2{
						Blocks: []state.L2BlockRaw{{
							BlockNumber:         1,
							ChangeL2BlockHeader: state.ChangeL2BlockHeader{},
							Transactions: []state.L2TxRaw{{
								Tx: tx1,
							}},
						}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ss := SequenceSender{
				wipBatch:     tt.wipBatch,
				sequenceData: tt.sequenceData,
				logger:       log.GetDefaultLogger(),
			}

			ss.addNewBlockTx(tt.l2Tx)
			require.Equal(t, tt.expectedSequenceData[tt.wipBatch].batchClosed, ss.sequenceData[tt.wipBatch].batchClosed)
			require.Equal(t, tt.expectedSequenceData[tt.wipBatch].batchType, ss.sequenceData[tt.wipBatch].batchType)
			require.Equal(t, tt.expectedSequenceData[tt.wipBatch].batch, ss.sequenceData[tt.wipBatch].batch)
		})
	}
}

func Test_handleReceivedDataStream(t *testing.T) {
	t.Parallel()

	l2Block, err := proto.Marshal(&datastream.L2Block{
		Number: 2,
	})
	require.NoError(t, err)

	prevL2Block, err := proto.Marshal(&datastream.L2Block{
		Number: 1,
	})
	require.NoError(t, err)

	l2TxRaw, err := proto.Marshal(&datastream.Transaction{
		Encoded:     hexutils.HexToBytes(txStreamEncoded1),
		ImStateRoot: []byte{1, 2, 3, 5, 6, 7, 8, 9, 0},
	})
	require.NoError(t, err)

	batchEndRaw, err := proto.Marshal(&datastream.BatchEnd{
		Number: 2,
	})
	require.NoError(t, err)

	tests := []struct {
		name              string
		entry             *datastreamer.FileEntry
		prevStreamEntry   *datastreamer.FileEntry
		fromStreamBatch   uint64
		latestStreamBatch uint64
		validStream       bool
		wipBatch          uint64
		sequenceData      map[uint64]*sequenceData
		expectedErr       error
	}{
		{
			name: "successfully handled L2 block",
			entry: &datastreamer.FileEntry{
				Type: 2,
			},
			prevStreamEntry: &datastreamer.FileEntry{
				Type: 1,
			},
			fromStreamBatch:   1,
			latestStreamBatch: 2,
			validStream:       true,
		},
		{
			name: "successfully handled L2 block with the same prev block",
			entry: &datastreamer.FileEntry{
				Type: 2,
				Data: l2Block,
			},
			prevStreamEntry: &datastreamer.FileEntry{
				Type: 2,
				Data: prevL2Block,
			},
			fromStreamBatch:   1,
			latestStreamBatch: 2,
			validStream:       true,
		},
		{
			name: "successfully handled transaction",
			entry: &datastreamer.FileEntry{
				Type: 3,
				Data: l2TxRaw,
			},
			prevStreamEntry: &datastreamer.FileEntry{
				Type: 2,
			},
			fromStreamBatch:   1,
			latestStreamBatch: 2,
			validStream:       true,
			wipBatch:          2,
			sequenceData: map[uint64]*sequenceData{
				2: {
					batchRaw: &state.BatchRawV2{
						Blocks: []state.L2BlockRaw{{
							BlockNumber: 1,
						}},
					},
				},
			},
		},
		{
			name: "successfully handled batch start",
			entry: &datastreamer.FileEntry{
				Type: 1,
			},
			prevStreamEntry: &datastreamer.FileEntry{
				Type: 1,
			},
			fromStreamBatch:   1,
			latestStreamBatch: 2,
			validStream:       true,
		},
		{
			name: "successfully handled batch end",
			entry: &datastreamer.FileEntry{
				Type: 4,
				Data: batchEndRaw,
			},
			prevStreamEntry: &datastreamer.FileEntry{
				Type: 2,
			},
			fromStreamBatch:   1,
			latestStreamBatch: 2,
			validStream:       true,
			wipBatch:          2,
			sequenceData: map[uint64]*sequenceData{
				2: {
					batch: txbuilder.NewBananaBatch(&etherman.Batch{
						BatchNumber: 2,
					}),
					batchRaw: &state.BatchRawV2{
						Blocks: []state.L2BlockRaw{{
							BlockNumber: 1,
						}},
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
				prevStreamEntry:   tt.prevStreamEntry,
				fromStreamBatch:   tt.fromStreamBatch,
				latestStreamBatch: tt.latestStreamBatch,
				validStream:       tt.validStream,
				sequenceData:      tt.sequenceData,
				wipBatch:          tt.wipBatch,
				logger:            log.GetDefaultLogger(),
			}

			err := s.handleReceivedDataStream(tt.entry, nil, nil)
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
