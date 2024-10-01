package aggregator

import (
	"bytes"
	"context"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	mocks "github.com/0xPolygon/cdk/aggregator/mocks"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/state"
	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/0xPolygonHermez/zkevm-data-streamer/datastreamer"
	"github.com/0xPolygonHermez/zkevm-synchronizer-l1/synchronizer"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func Test_resetCurrentBatchData(t *testing.T) {
	agg := Aggregator{
		currentBatchStreamData: []byte("test"),
		currentStreamBatchRaw: state.BatchRawV2{
			Blocks: []state.L2BlockRaw{
				{
					BlockNumber:         1,
					ChangeL2BlockHeader: state.ChangeL2BlockHeader{},
					Transactions:        []state.L2TxRaw{},
				},
			},
		},
		currentStreamL2Block: state.L2BlockRaw{},
	}

	agg.resetCurrentBatchData()

	assert.Equal(t, []byte{}, agg.currentBatchStreamData)
	assert.Equal(t, state.BatchRawV2{Blocks: make([]state.L2BlockRaw, 0)}, agg.currentStreamBatchRaw)
	assert.Equal(t, state.L2BlockRaw{}, agg.currentStreamL2Block)
}

// func Test_retrieveWitnesst(t *testing.T) {
// 	mockState := new(mocks.StateInterfaceMock)

// 	witnessChan := make(chan state.DBBatch)
// 	agg := Aggregator{
// 		witnessRetrievalChan: witnessChan,
// 		state:                mockState,
// 		cfg: Config{
// 			RetryTime: types.Duration{Duration: 1 * time.Second},
// 		},
// 	}

// 	mockState.On("AddBatch", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
// 	// .On("getWitness", mock.Anything, mock.Anything, mock.Anything).Return([]byte("witness data"), nil)

// 	// Send a mock DBatch to the witness retrieval channel
// 	witnessChan <- state.DBBatch{
// 		Batch: state.Batch{BatchNumber: 1},
// 	}

// 	go agg.retrieveWitness()
// 	time.Sleep(100 * time.Millisecond)

// 	mockState.AssertExpectations(t)
// }

func Test_handleReorg(t *testing.T) {
	t.Parallel()

	mockL1Syncr := new(mocks.SynchronizerInterfaceMock)
	mockState := new(mocks.StateInterfaceMock)
	reorgData := synchronizer.ReorgExecutionResult{}

	agg := &Aggregator{
		l1Syncr: mockL1Syncr,
		state:   mockState,
		logger:  log.GetDefaultLogger(),
		halted:  atomic.Bool{},
		ctx:     context.Background(),
	}

	mockL1Syncr.On("GetLastestVirtualBatchNumber", mock.Anything).Return(uint64(100), nil)
	mockState.On("DeleteBatchesNewerThanBatchNumber", mock.Anything, uint64(100), mock.Anything).Return(nil)

	go agg.handleReorg(reorgData)
	time.Sleep(11 * time.Second)

	mockState.AssertExpectations(t)
	mockL1Syncr.AssertExpectations(t)
}

func Test_handleRollbackBatches(t *testing.T) {
	t.Parallel()

	mockStreamClient := new(mocks.StreamClientMock)
	mockEtherman := new(mocks.EthermanMock)
	mockState := new(mocks.StateInterfaceMock)

	// Test data
	rollbackData := synchronizer.RollbackBatchesData{
		LastBatchNumber: 100,
	}

	mockStreamClient.On("IsStarted").Return(true)
	mockStreamClient.On("ResetProcessEntryFunc").Return()
	mockStreamClient.On("SetProcessEntryFunc", mock.Anything).Return()
	mockStreamClient.On("ExecCommandStop").Return(nil)
	mockStreamClient.On("Start").Return(nil)
	mockStreamClient.On("ExecCommandStartBookmark", mock.Anything).Return(nil)
	mockEtherman.On("GetLatestVerifiedBatchNum").Return(uint64(90), nil)
	mockState.On("DeleteBatchesNewerThanBatchNumber", mock.Anything, rollbackData.LastBatchNumber, nil).Return(nil)
	mockState.On("DeleteBatchesOlderThanBatchNumber", mock.Anything, rollbackData.LastBatchNumber, nil).Return(nil)
	mockState.On("DeleteUngeneratedProofs", mock.Anything, nil).Return(nil)
	mockState.On("DeleteGeneratedProofs", mock.Anything, rollbackData.LastBatchNumber+1, maxDBBigIntValue, nil).Return(nil)

	agg := Aggregator{
		ctx:                    context.Background(),
		streamClient:           mockStreamClient,
		Etherman:               mockEtherman,
		state:                  mockState,
		logger:                 log.GetDefaultLogger(),
		halted:                 atomic.Bool{},
		streamClientMutex:      &sync.Mutex{},
		currentBatchStreamData: []byte{},
		currentStreamBatchRaw:  state.BatchRawV2{},
		currentStreamL2Block:   state.L2BlockRaw{},
	}

	agg.handleRollbackBatches(rollbackData)

	mockStreamClient.AssertExpectations(t)
	mockEtherman.AssertExpectations(t)
	mockState.AssertExpectations(t)
}

func Test_handleReceivedDataStream_BATCH_START(t *testing.T) {
	mockState := new(mocks.StateInterfaceMock)
	mockL1Syncr := new(mocks.SynchronizerInterfaceMock)
	agg := Aggregator{
		state:              mockState,
		l1Syncr:            mockL1Syncr,
		logger:             log.GetDefaultLogger(),
		halted:             atomic.Bool{},
		currentStreamBatch: state.Batch{},
	}

	// Prepare a FileEntry for Batch Start
	batchStartData, err := proto.Marshal(&datastream.BatchStart{
		Number:  1,
		ChainId: 2,
		ForkId:  3,
		Type:    datastream.BatchType_BATCH_TYPE_REGULAR,
	})
	assert.NoError(t, err)

	batchStartEntry := &datastreamer.FileEntry{
		Type: datastreamer.EntryType(datastream.EntryType_ENTRY_TYPE_BATCH_START),
		Data: batchStartData,
	}

	// Test the handleReceivedDataStream for Batch Start
	err = agg.handleReceivedDataStream(batchStartEntry, nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, agg.currentStreamBatch.BatchNumber, uint64(1))
	assert.Equal(t, agg.currentStreamBatch.ChainID, uint64(2))
	assert.Equal(t, agg.currentStreamBatch.ForkID, uint64(3))
	assert.Equal(t, agg.currentStreamBatch.Type, datastream.BatchType_BATCH_TYPE_REGULAR)
}

func Test_handleReceivedDataStream_BATCH_END(t *testing.T) {
	mockState := new(mocks.StateInterfaceMock)
	mockL1Syncr := new(mocks.SynchronizerInterfaceMock)
	agg := Aggregator{
		state:   mockState,
		l1Syncr: mockL1Syncr,
		logger:  log.GetDefaultLogger(),
		halted:  atomic.Bool{},
		currentStreamBatch: state.Batch{
			BatchNumber: uint64(2),
			Type:        datastream.BatchType_BATCH_TYPE_REGULAR,
			Coinbase:    common.Address{},
		},
		currentStreamL2Block: state.L2BlockRaw{
			BlockNumber: uint64(10),
		},
		currentStreamBatchRaw: state.BatchRawV2{
			Blocks: []state.L2BlockRaw{
				{
					BlockNumber:         uint64(9),
					ChangeL2BlockHeader: state.ChangeL2BlockHeader{},
					Transactions:        []state.L2TxRaw{},
				},
			},
		},
		cfg: Config{
			UseL1BatchData: false,
		},
	}

	batchEndData, err := proto.Marshal(&datastream.BatchEnd{
		Number:        1,
		LocalExitRoot: []byte{1, 2, 3},
		StateRoot:     []byte{4, 5, 6},
		Debug:         nil,
	})
	assert.NoError(t, err)

	batchEndEntry := &datastreamer.FileEntry{
		Type: datastreamer.EntryType(datastream.EntryType_ENTRY_TYPE_BATCH_END),
		Data: batchEndData,
	}

	mockState.On("GetBatch", mock.Anything, agg.currentStreamBatch.BatchNumber-1, nil).
		Return(&state.DBBatch{
			Batch: state.Batch{
				AccInputHash: common.Hash{},
			},
		}, nil).Once()
	mockState.On("GetBatch", mock.Anything, agg.currentStreamBatch.BatchNumber, nil).
		Return(&state.DBBatch{
			Witness: []byte("test_witness"),
		}, nil).Once()
	mockState.On("AddBatch", mock.Anything, mock.Anything, nil).Return(nil).Once()
	mockL1Syncr.On("GetVirtualBatchByBatchNumber", mock.Anything, agg.currentStreamBatch.BatchNumber).Return(&synchronizer.VirtualBatch{BatchL2Data: []byte{1, 2, 3}}, nil).Once()
	mockL1Syncr.On("GetSequenceByBatchNumber", mock.Anything, agg.currentStreamBatch.BatchNumber).
		Return(&synchronizer.SequencedBatches{
			L1InfoRoot: common.Hash{},
			Timestamp:  time.Now(),
		}, nil).Once()

	err = agg.handleReceivedDataStream(batchEndEntry, nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, agg.currentBatchStreamData, []byte{})
	assert.Equal(t, agg.currentStreamBatchRaw, state.BatchRawV2{Blocks: make([]state.L2BlockRaw, 0)})
	assert.Equal(t, agg.currentStreamL2Block, state.L2BlockRaw{})

	// Verify the mock expectations
	mockState.AssertExpectations(t)
	mockL1Syncr.AssertExpectations(t)
}

func Test_handleReceivedDataStream_L2_BLOCK(t *testing.T) {
	t.Parallel()

	agg := Aggregator{
		currentStreamL2Block: state.L2BlockRaw{
			BlockNumber: uint64(9),
		},
		currentStreamBatchRaw: state.BatchRawV2{
			Blocks: []state.L2BlockRaw{},
		},
		currentStreamBatch: state.Batch{},
	}

	// Mock data for L2Block
	l2Block := &datastream.L2Block{
		Number:          uint64(10),
		DeltaTimestamp:  uint32(5),
		L1InfotreeIndex: uint32(1),
		Coinbase:        []byte{0x01},
		GlobalExitRoot:  []byte{0x02},
	}

	l2BlockData, err := proto.Marshal(l2Block)
	assert.NoError(t, err)

	l2BlockEntry := &datastreamer.FileEntry{
		Type: datastreamer.EntryType(datastream.EntryType_ENTRY_TYPE_L2_BLOCK),
		Data: l2BlockData,
	}

	err = agg.handleReceivedDataStream(l2BlockEntry, nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, uint64(10), agg.currentStreamL2Block.BlockNumber)
	assert.Equal(t, uint32(5), agg.currentStreamL2Block.ChangeL2BlockHeader.DeltaTimestamp)
	assert.Equal(t, uint32(1), agg.currentStreamL2Block.ChangeL2BlockHeader.IndexL1InfoTree)
	assert.Equal(t, 0, len(agg.currentStreamL2Block.Transactions))
	assert.Equal(t, uint32(1), agg.currentStreamBatch.L1InfoTreeIndex)
	assert.Equal(t, common.BytesToAddress([]byte{0x01}), agg.currentStreamBatch.Coinbase)
	assert.Equal(t, common.BytesToHash([]byte{0x02}), agg.currentStreamBatch.GlobalExitRoot)
}

func Test_handleReceivedDataStream_TRANSACTION(t *testing.T) {
	t.Parallel()

	agg := Aggregator{
		currentStreamL2Block: state.L2BlockRaw{
			Transactions: []state.L2TxRaw{},
		},
		logger: log.GetDefaultLogger(),
	}

	tx := ethTypes.NewTransaction(
		0,
		common.HexToAddress("0x01"),
		big.NewInt(1000000000000000000),
		uint64(21000),
		big.NewInt(20000000000),
		nil,
	)

	// Encode transaction into RLP format
	var buf bytes.Buffer
	if err := tx.EncodeRLP(&buf); err != nil {
		t.Fatalf("Failed to encode transaction: %v", err)
	}

	transaction := &datastream.Transaction{
		L2BlockNumber:               uint64(10),
		Index:                       uint64(0),
		IsValid:                     true,
		Encoded:                     buf.Bytes(),
		EffectiveGasPricePercentage: uint32(90),
	}

	transactionData, err := proto.Marshal(transaction)
	assert.NoError(t, err)

	transactionEntry := &datastreamer.FileEntry{
		Type: datastreamer.EntryType(datastream.EntryType_ENTRY_TYPE_TRANSACTION),
		Data: transactionData,
	}

	err = agg.handleReceivedDataStream(transactionEntry, nil, nil)
	assert.NoError(t, err)

	assert.Len(t, agg.currentStreamL2Block.Transactions, 1)
	assert.Equal(t, uint8(90), agg.currentStreamL2Block.Transactions[0].EfficiencyPercentage)
	assert.False(t, agg.currentStreamL2Block.Transactions[0].TxAlreadyEncoded)
	assert.NotNil(t, agg.currentStreamL2Block.Transactions[0].Tx)
}
