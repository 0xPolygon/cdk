package aggregator

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	mocks "github.com/0xPolygon/cdk/aggregator/mocks"
	"github.com/0xPolygon/cdk/aggregator/prover"
	"github.com/0xPolygon/cdk/config/types"
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
	mockState.On("DeleteGeneratedProofs", mock.Anything, rollbackData.LastBatchNumber+1, mock.AnythingOfType("uint64"), nil).Return(nil)

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

func Test_final_GetBatchError(t *testing.T) {
	assert := assert.New(t)
	errTest := errors.New("test error")
	batchNum := uint64(23)
	batchNumFinal := uint64(42)

	proofID := "proofId"
	proverName := "proverName"
	proverID := "proverID"
	recursiveProof := &state.Proof{
		Prover:           &proverName,
		ProverID:         &proverID,
		ProofID:          &proofID,
		BatchNumber:      batchNum,
		BatchNumberFinal: batchNumFinal,
	}
	finalProof := &prover.FinalProof{}

	mockState := new(mocks.StateInterfaceMock)
	mockEtherman := new(mocks.EthermanMock)
	mockEthTxManager := new(mocks.EthTxManagerClientMock)

	agg := Aggregator{
		state:                   mockState,
		Etherman:                mockEtherman,
		ethTxManager:            mockEthTxManager,
		finalProof:              make(chan finalProofMsg),
		logger:                  log.GetDefaultLogger(),
		verifyingProof:          false,
		stateDBMutex:            &sync.Mutex{},
		timeSendFinalProofMutex: &sync.RWMutex{},
	}

	agg.ctx, agg.exit = context.WithCancel(context.Background())
	mockState.On("GetBatch", mock.Anything, batchNumFinal, nil).Run(func(args mock.Arguments) {
		// test is done, stop the sendFinalProof method
		fmt.Println("Stopping sendFinalProof")
		agg.exit()
	}).Return(nil, errTest).Once()

	go func() {
		finalMsg := finalProofMsg{
			proverID:       proverID,
			recursiveProof: recursiveProof,
			finalProof:     finalProof,
		}
		agg.finalProof <- finalMsg
	}()

	agg.sendFinalProof()
	assert.False(agg.verifyingProof)
}

func Test_final_SettleAgglayerErrorSendTx(t *testing.T) {
	assert := assert.New(t)
	errTest := errors.New("test error")
	batchNum := uint64(23)
	batchNumFinal := uint64(42)

	proofID := "proofId"
	proverName := "proverName"
	proverID := "proverID"
	recursiveProof := &state.Proof{
		Prover:           &proverName,
		ProverID:         &proverID,
		ProofID:          &proofID,
		BatchNumber:      batchNum,
		BatchNumberFinal: batchNumFinal,
	}
	finalProof := &prover.FinalProof{
		Proof: "test proof",
	}

	mockState := new(mocks.StateInterfaceMock)
	mockEtherman := new(mocks.EthermanMock)
	mockEthTxManager := new(mocks.EthTxManagerClientMock)
	mockAgglayerClientInterface := new(mocks.AgglayerClientInterfaceMock)

	cfg := Config{
		SettlementBackend: AggLayer,
	}

	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatal("Error generating key")
	}

	agg := Aggregator{
		state:                   mockState,
		Etherman:                mockEtherman,
		ethTxManager:            mockEthTxManager,
		aggLayerClient:          mockAgglayerClientInterface,
		finalProof:              make(chan finalProofMsg),
		logger:                  log.GetDefaultLogger(),
		verifyingProof:          false,
		stateDBMutex:            &sync.Mutex{},
		timeSendFinalProofMutex: &sync.RWMutex{},
		cfg:                     cfg,
		sequencerPrivateKey:     privateKey,
	}

	agg.ctx, agg.exit = context.WithCancel(context.Background())
	mockState.On("GetBatch", mock.Anything, batchNumFinal, nil).Run(func(args mock.Arguments) {
	}).Return(&state.DBBatch{
		Batch: state.Batch{
			LocalExitRoot: common.Hash{},
			StateRoot:     common.Hash{},
		},
	}, nil).Once()

	mockEtherman.On("GetRollupId").Return(uint32(1)).Once()
	mockAgglayerClientInterface.On("SendTx", mock.Anything).Run(func(args mock.Arguments) {
		// test is done, stop the sendFinalProof method
		fmt.Println("Stopping sendFinalProof")
		agg.exit()
	}).Return(nil, errTest).Once()
	mockState.On("UpdateGeneratedProof", mock.Anything, mock.Anything, nil).Return(nil).Once()

	go func() {
		finalMsg := finalProofMsg{
			proverID:       proverID,
			recursiveProof: recursiveProof,
			finalProof:     finalProof,
		}
		agg.finalProof <- finalMsg
	}()

	agg.sendFinalProof()
	assert.False(agg.verifyingProof)
}

func Test_final_SettleAgglayerErrorWaitTxToBeMined(t *testing.T) {
	assert := assert.New(t)
	errTest := errors.New("test error")
	batchNum := uint64(23)
	batchNumFinal := uint64(42)

	proofID := "proofId"
	proverName := "proverName"
	proverID := "proverID"
	recursiveProof := &state.Proof{
		Prover:           &proverName,
		ProverID:         &proverID,
		ProofID:          &proofID,
		BatchNumber:      batchNum,
		BatchNumberFinal: batchNumFinal,
	}
	finalProof := &prover.FinalProof{
		Proof: "test proof",
	}

	mockState := new(mocks.StateInterfaceMock)
	mockEtherman := new(mocks.EthermanMock)
	mockEthTxManager := new(mocks.EthTxManagerClientMock)
	mockAgglayerClientInterface := new(mocks.AgglayerClientInterfaceMock)

	cfg := Config{
		SettlementBackend: AggLayer,
		AggLayerTxTimeout: types.Duration{Duration: time.Millisecond * 1},
	}

	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatal("Error generating key")
	}

	agg := Aggregator{
		state:                   mockState,
		Etherman:                mockEtherman,
		ethTxManager:            mockEthTxManager,
		aggLayerClient:          mockAgglayerClientInterface,
		finalProof:              make(chan finalProofMsg),
		logger:                  log.GetDefaultLogger(),
		verifyingProof:          false,
		stateDBMutex:            &sync.Mutex{},
		timeSendFinalProofMutex: &sync.RWMutex{},
		cfg:                     cfg,
		sequencerPrivateKey:     privateKey,
	}

	agg.ctx, agg.exit = context.WithCancel(context.Background())
	mockState.On("GetBatch", mock.Anything, batchNumFinal, nil).Run(func(args mock.Arguments) {
	}).Return(&state.DBBatch{
		Batch: state.Batch{
			LocalExitRoot: common.Hash{},
			StateRoot:     common.Hash{},
		},
	}, nil).Once()

	mockEtherman.On("GetRollupId").Return(uint32(1)).Once()
	mockAgglayerClientInterface.On("SendTx", mock.Anything).Return(common.Hash{}, nil).Once()
	mockAgglayerClientInterface.On("WaitTxToBeMined", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		fmt.Println("Stopping sendFinalProof")
		agg.exit()
	}).Return(errTest)
	mockState.On("UpdateGeneratedProof", mock.Anything, mock.Anything, nil).Return(nil).Once()

	go func() {
		finalMsg := finalProofMsg{
			proverID:       proverID,
			recursiveProof: recursiveProof,
			finalProof:     finalProof,
		}
		agg.finalProof <- finalMsg
	}()

	agg.sendFinalProof()
	assert.False(agg.verifyingProof)
}

func Test_final_SettleAgglayerErrorSuccess(t *testing.T) {
	assert := assert.New(t)
	// errTest := errors.New("test error")
	batchNum := uint64(23)
	batchNumFinal := uint64(42)

	proofID := "proofId"
	proverName := "proverName"
	proverID := "proverID"
	recursiveProof := &state.Proof{
		Prover:           &proverName,
		ProverID:         &proverID,
		ProofID:          &proofID,
		BatchNumber:      batchNum,
		BatchNumberFinal: batchNumFinal,
	}
	finalProof := &prover.FinalProof{
		Proof: "test proof",
	}

	mockState := new(mocks.StateInterfaceMock)
	mockEtherman := new(mocks.EthermanMock)
	mockEthTxManager := new(mocks.EthTxManagerClientMock)
	mockAgglayerClientInterface := new(mocks.AgglayerClientInterfaceMock)

	cfg := Config{
		SettlementBackend: AggLayer,
		AggLayerTxTimeout: types.Duration{Duration: time.Millisecond * 1},
	}

	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatal("Error generating key")
	}

	agg := Aggregator{
		state:                   mockState,
		Etherman:                mockEtherman,
		ethTxManager:            mockEthTxManager,
		aggLayerClient:          mockAgglayerClientInterface,
		finalProof:              make(chan finalProofMsg),
		logger:                  log.GetDefaultLogger(),
		verifyingProof:          false,
		stateDBMutex:            &sync.Mutex{},
		timeSendFinalProofMutex: &sync.RWMutex{},
		cfg:                     cfg,
		sequencerPrivateKey:     privateKey,
	}

	agg.ctx, agg.exit = context.WithCancel(context.Background())
	mockState.On("GetBatch", mock.Anything, batchNumFinal, nil).Run(func(args mock.Arguments) {
	}).Return(&state.DBBatch{
		Batch: state.Batch{
			LocalExitRoot: common.Hash{},
			StateRoot:     common.Hash{},
		},
	}, nil).Once()

	mockEtherman.On("GetRollupId").Return(uint32(1)).Once()
	mockAgglayerClientInterface.On("SendTx", mock.Anything).Return(common.Hash{}, nil).Once()
	mockAgglayerClientInterface.On("WaitTxToBeMined", mock.Anything, mock.Anything).Return(nil)
	mockState.On("UpdateGeneratedProof", mock.Anything, mock.Anything, nil).Return(nil).Once()

	go func() {
		finalMsg := finalProofMsg{
			proverID:       proverID,
			recursiveProof: recursiveProof,
			finalProof:     finalProof,
		}
		agg.finalProof <- finalMsg
		time.Sleep(1 * time.Second)
		agg.exit()
	}()

	agg.sendFinalProof()
	assert.False(agg.verifyingProof)
}

func Test_final_SettleDirect_ErrorBuildTrustedVerifyBatchesTxData(t *testing.T) {
	assert := assert.New(t)
	errTest := errors.New("test error")
	batchNum := uint64(23)
	batchNumFinal := uint64(42)

	proofID := "proofId"
	proverName := "proverName"
	proverID := "proverID"
	recursiveProof := &state.Proof{
		Prover:           &proverName,
		ProverID:         &proverID,
		ProofID:          &proofID,
		BatchNumber:      batchNum,
		BatchNumberFinal: batchNumFinal,
	}
	finalProof := &prover.FinalProof{
		Proof: "test proof",
	}

	mockState := new(mocks.StateInterfaceMock)
	mockEtherman := new(mocks.EthermanMock)

	cfg := Config{
		SettlementBackend: L1,
		SenderAddress:     "0x6278712b352Ef1dB57a5f74B79a7da78a369A9b3",
	}

	agg := Aggregator{
		state:                   mockState,
		Etherman:                mockEtherman,
		finalProof:              make(chan finalProofMsg),
		logger:                  log.GetDefaultLogger(),
		verifyingProof:          false,
		stateDBMutex:            &sync.Mutex{},
		timeSendFinalProofMutex: &sync.RWMutex{},
		cfg:                     cfg,
	}

	agg.ctx, agg.exit = context.WithCancel(context.Background())
	mockState.On("GetBatch", mock.Anything, batchNumFinal, nil).Run(func(args mock.Arguments) {
	}).Return(&state.DBBatch{
		Batch: state.Batch{
			LocalExitRoot: common.Hash{},
			StateRoot:     common.Hash{},
		},
	}, nil).Once()

	mockEtherman.On("GetRollupId").Return(uint32(1)).Once()
	mockEtherman.On("BuildTrustedVerifyBatchesTxData", batchNum-1, batchNumFinal, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		fmt.Println("Stopping sendFinalProof")
		agg.exit()
	}).Return(nil, nil, errTest)
	mockState.On("UpdateGeneratedProof", mock.Anything, mock.Anything, nil).Return(nil).Once()

	go func() {
		finalMsg := finalProofMsg{
			proverID:       proverID,
			recursiveProof: recursiveProof,
			finalProof:     finalProof,
		}
		agg.finalProof <- finalMsg
	}()

	agg.sendFinalProof()
	assert.False(agg.verifyingProof)
}

func Test_final_SettleDirect_Success(t *testing.T) {
	assert := assert.New(t)
	errTest := errors.New("test error")
	batchNum := uint64(23)
	batchNumFinal := uint64(42)

	proofID := "proofId"
	proverName := "proverName"
	proverID := "proverID"
	recursiveProof := &state.Proof{
		Prover:           &proverName,
		ProverID:         &proverID,
		ProofID:          &proofID,
		BatchNumber:      batchNum,
		BatchNumberFinal: batchNumFinal,
	}
	finalProof := &prover.FinalProof{
		Proof: "test proof",
	}

	mockState := new(mocks.StateInterfaceMock)
	mockEtherman := new(mocks.EthermanMock)
	mockEthTxManager := new(mocks.EthTxManagerClientMock)

	cfg := Config{
		SettlementBackend: L1,
		SenderAddress:     "0x6278712b352Ef1dB57a5f74B79a7da78a369A9b3",
	}

	agg := Aggregator{
		state:                   mockState,
		Etherman:                mockEtherman,
		ethTxManager:            mockEthTxManager,
		finalProof:              make(chan finalProofMsg),
		logger:                  log.GetDefaultLogger(),
		verifyingProof:          false,
		stateDBMutex:            &sync.Mutex{},
		timeSendFinalProofMutex: &sync.RWMutex{},
		cfg:                     cfg,
	}

	agg.ctx, agg.exit = context.WithCancel(context.Background())
	mockState.On("GetBatch", mock.Anything, batchNumFinal, nil).Run(func(args mock.Arguments) {
	}).Return(&state.DBBatch{
		Batch: state.Batch{
			LocalExitRoot: common.Hash{},
			StateRoot:     common.Hash{},
		},
	}, nil).Once()

	mockEtherman.On("GetRollupId").Return(uint32(1)).Once()
	mockEtherman.On("BuildTrustedVerifyBatchesTxData", batchNum-1, batchNumFinal, mock.Anything, mock.Anything).Return(nil, nil, errTest).Once()
	mockEthTxManager.On("Add", mock.Anything, mock.Anything, nil, big.NewInt(0), mock.Anything, agg.cfg.GasOffset, nil).Return(nil, errTest).Once()
	mockState.On("UpdateGeneratedProof", mock.Anything, mock.Anything, nil).Return(nil).Once()

	go func() {
		finalMsg := finalProofMsg{
			proverID:       proverID,
			recursiveProof: recursiveProof,
			finalProof:     finalProof,
		}
		agg.finalProof <- finalMsg

		time.Sleep(time.Second)
		agg.exit()
	}()

	agg.sendFinalProof()
	assert.False(agg.verifyingProof)
}

func Test_buildFinalProof(t *testing.T) {
	t.Parallel()

	mockProver := new(mocks.ProverInterfaceMock)

	batchNum := uint64(23)
	batchNumFinal := uint64(42)
	proofID := "proofId"
	proverID := "proverID"
	recursiveProof := &state.Proof{
		ProverID:         &proverID,
		Proof:            "test proof",
		ProofID:          &proofID,
		BatchNumber:      batchNum,
		BatchNumberFinal: batchNumFinal,
	}

	agg := Aggregator{
		logger: log.GetDefaultLogger(),
		cfg: Config{
			SenderAddress: "0x6278712b352Ef1dB57a5f74B79a7da78a369A9b3",
		},
	}

	finalProofID := "finalProofID"
	finalProof := prover.FinalProof{
		Public: &prover.PublicInputsExtended{
			NewStateRoot:     []byte("StateRoot"),
			NewLocalExitRoot: []byte("LocalExitRoot"),
		},
	}

	mockProver.On("Name").Return("name").Once()
	mockProver.On("ID").Return("id").Once()
	mockProver.On("Addr").Return("addr").Once()
	mockProver.On("FinalProof", recursiveProof.Proof, agg.cfg.SenderAddress).Return(&finalProofID, nil).Once()
	mockProver.On("WaitFinalProof", mock.Anything, finalProofID).Return(&finalProof, nil).Once()

	fProof, err := agg.buildFinalProof(context.Background(), mockProver, recursiveProof)
	assert.NoError(t, err)
	assert.Equal(t, finalProof.Public.NewStateRoot, fProof.Public.NewStateRoot)
	assert.Equal(t, finalProof.Public.NewLocalExitRoot, fProof.Public.NewLocalExitRoot)
}

func Test_buildFinalProof_MockProver(t *testing.T) {
	t.Parallel()

	mockProver := new(mocks.ProverInterfaceMock)
	mockState := new(mocks.StateInterfaceMock)

	batchNum := uint64(23)
	batchNumFinal := uint64(42)
	proofID := "proofId"
	proverID := "proverID"
	recursiveProof := &state.Proof{
		ProverID:         &proverID,
		Proof:            "test proof",
		ProofID:          &proofID,
		BatchNumber:      batchNum,
		BatchNumberFinal: batchNumFinal,
	}

	agg := Aggregator{
		state:  mockState,
		logger: log.GetDefaultLogger(),
		cfg: Config{
			SenderAddress: "0x6278712b352Ef1dB57a5f74B79a7da78a369A9b3",
		},
	}

	finalProofID := "finalProofID"
	finalProof := prover.FinalProof{
		Public: &prover.PublicInputsExtended{
			NewStateRoot:     []byte(mockedStateRoot),
			NewLocalExitRoot: []byte(mockedLocalExitRoot),
		},
	}

	finalDBBatch := &state.DBBatch{
		Batch: state.Batch{
			StateRoot:     common.Hash{},
			LocalExitRoot: common.Hash{},
		},
	}

	mockProver.On("Name").Return("name").Once()
	mockProver.On("ID").Return("id").Once()
	mockProver.On("Addr").Return("addr").Once()
	mockProver.On("FinalProof", recursiveProof.Proof, agg.cfg.SenderAddress).Return(&finalProofID, nil).Once()
	mockProver.On("WaitFinalProof", mock.Anything, finalProofID).Return(&finalProof, nil).Once()

	mockState.On("GetBatch", mock.Anything, batchNumFinal, nil).Return(finalDBBatch, nil)

	fProof, err := agg.buildFinalProof(context.Background(), mockProver, recursiveProof)
	assert.NoError(t, err)
	assert.Equal(t, common.Hash{}.Bytes(), fProof.Public.NewStateRoot)
	assert.Equal(t, common.Hash{}.Bytes(), fProof.Public.NewLocalExitRoot)
}
