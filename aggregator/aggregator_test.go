package aggregator

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
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
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

var (
	proofID      = "proofId"
	proof        = "proof"
	finalProofID = "finalProofID"
	proverName   = "proverName"
	proverID     = "proverID"
)

type mox struct {
	stateMock          *mocks.StateInterfaceMock
	ethTxManager       *mocks.EthTxManagerClientMock
	etherman           *mocks.EthermanMock
	proverMock         *mocks.ProverInterfaceMock
	aggLayerClientMock *mocks.AgglayerClientInterfaceMock
}

func Test_resetCurrentBatchData(t *testing.T) {
	t.Parallel()

	a := Aggregator{
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

	a.resetCurrentBatchData()

	assert.Equal(t, []byte{}, a.currentBatchStreamData)
	assert.Equal(t, state.BatchRawV2{Blocks: make([]state.L2BlockRaw, 0)}, a.currentStreamBatchRaw)
	assert.Equal(t, state.L2BlockRaw{}, a.currentStreamL2Block)
}

func Test_handleReorg(t *testing.T) {
	t.Parallel()

	mockL1Syncr := new(mocks.SynchronizerInterfaceMock)
	mockState := new(mocks.StateInterfaceMock)
	reorgData := synchronizer.ReorgExecutionResult{}

	a := &Aggregator{
		l1Syncr: mockL1Syncr,
		state:   mockState,
		logger:  log.GetDefaultLogger(),
		halted:  atomic.Bool{},
		ctx:     context.Background(),
	}

	mockL1Syncr.On("GetLastestVirtualBatchNumber", mock.Anything).Return(uint64(100), nil).Once()
	mockState.On("DeleteBatchesNewerThanBatchNumber", mock.Anything, uint64(100), mock.Anything).Return(nil).Once()

	go a.handleReorg(reorgData)
	time.Sleep(3 * time.Second)

	assert.True(t, a.halted.Load())
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

	mockStreamClient.On("IsStarted").Return(true).Once()
	mockStreamClient.On("ResetProcessEntryFunc").Return().Once()
	mockStreamClient.On("SetProcessEntryFunc", mock.Anything).Return().Once()
	mockStreamClient.On("ExecCommandStop").Return(nil).Once()
	mockStreamClient.On("Start").Return(nil).Once()
	mockStreamClient.On("ExecCommandStartBookmark", mock.Anything).Return(nil).Once()
	mockEtherman.On("GetLatestVerifiedBatchNum").Return(uint64(90), nil).Once()
	mockState.On("DeleteBatchesNewerThanBatchNumber", mock.Anything, rollbackData.LastBatchNumber, nil).Return(nil).Once()
	mockState.On("DeleteBatchesOlderThanBatchNumber", mock.Anything, rollbackData.LastBatchNumber, nil).Return(nil).Once()
	mockState.On("DeleteUngeneratedProofs", mock.Anything, nil).Return(nil).Once()
	mockState.On("DeleteGeneratedProofs", mock.Anything, rollbackData.LastBatchNumber+1, mock.AnythingOfType("uint64"), nil).Return(nil).Once()

	a := Aggregator{
		ctx:                    context.Background(),
		streamClient:           mockStreamClient,
		etherman:               mockEtherman,
		state:                  mockState,
		logger:                 log.GetDefaultLogger(),
		halted:                 atomic.Bool{},
		streamClientMutex:      &sync.Mutex{},
		currentBatchStreamData: []byte{},
		currentStreamBatchRaw:  state.BatchRawV2{},
		currentStreamL2Block:   state.L2BlockRaw{},
	}

	a.halted.Store(false)
	a.handleRollbackBatches(rollbackData)

	assert.False(t, a.halted.Load())
	mockStreamClient.AssertExpectations(t)
	mockEtherman.AssertExpectations(t)
	mockState.AssertExpectations(t)
}

func Test_handleReceivedDataStream_BATCH_START(t *testing.T) {
	t.Parallel()

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

func Test_handleReceivedDataStreamBatchEnd(t *testing.T) {
	t.Parallel()

	mockState := new(mocks.StateInterfaceMock)
	mockL1Syncr := new(mocks.SynchronizerInterfaceMock)
	a := Aggregator{
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

	mockState.On("GetBatch", mock.Anything, a.currentStreamBatch.BatchNumber-1, nil).
		Return(&state.DBBatch{
			Batch: state.Batch{
				AccInputHash: common.Hash{},
			},
		}, nil).Once()
	mockState.On("GetBatch", mock.Anything, a.currentStreamBatch.BatchNumber, nil).
		Return(&state.DBBatch{
			Witness: []byte("test_witness"),
		}, nil).Once()
	mockState.On("AddBatch", mock.Anything, mock.Anything, nil).Return(nil).Once()
	mockL1Syncr.On("GetVirtualBatchByBatchNumber", mock.Anything, a.currentStreamBatch.BatchNumber).
		Return(&synchronizer.VirtualBatch{BatchL2Data: []byte{1, 2, 3}}, nil).Once()
	mockL1Syncr.On("GetSequenceByBatchNumber", mock.Anything, a.currentStreamBatch.BatchNumber).
		Return(&synchronizer.SequencedBatches{
			L1InfoRoot: common.Hash{},
			Timestamp:  time.Now(),
		}, nil).Once()

	err = a.handleReceivedDataStream(batchEndEntry, nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, a.currentBatchStreamData, []byte{})
	assert.Equal(t, a.currentStreamBatchRaw, state.BatchRawV2{Blocks: make([]state.L2BlockRaw, 0)})
	assert.Equal(t, a.currentStreamL2Block, state.L2BlockRaw{})

	mockState.AssertExpectations(t)
	mockL1Syncr.AssertExpectations(t)
}

func Test_handleReceivedDataStreamL2Block(t *testing.T) {
	t.Parallel()

	a := Aggregator{
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

	err = a.handleReceivedDataStream(l2BlockEntry, nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, uint64(10), a.currentStreamL2Block.BlockNumber)
	assert.Equal(t, uint32(5), a.currentStreamL2Block.ChangeL2BlockHeader.DeltaTimestamp)
	assert.Equal(t, uint32(1), a.currentStreamL2Block.ChangeL2BlockHeader.IndexL1InfoTree)
	assert.Equal(t, 0, len(a.currentStreamL2Block.Transactions))
	assert.Equal(t, uint32(1), a.currentStreamBatch.L1InfoTreeIndex)
	assert.Equal(t, common.BytesToAddress([]byte{0x01}), a.currentStreamBatch.Coinbase)
	assert.Equal(t, common.BytesToHash([]byte{0x02}), a.currentStreamBatch.GlobalExitRoot)
}

func Test_handleReceivedDataStreamTransaction(t *testing.T) {
	t.Parallel()

	a := Aggregator{
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

	err = a.handleReceivedDataStream(transactionEntry, nil, nil)
	assert.NoError(t, err)

	assert.Len(t, a.currentStreamL2Block.Transactions, 1)
	assert.Equal(t, uint8(90), a.currentStreamL2Block.Transactions[0].EfficiencyPercentage)
	assert.False(t, a.currentStreamL2Block.Transactions[0].TxAlreadyEncoded)
	assert.NotNil(t, a.currentStreamL2Block.Transactions[0].Tx)
}

func Test_sendFinalProofSuccess(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	batchNum := uint64(23)
	batchNumFinal := uint64(42)

	recursiveProof := &state.Proof{
		Prover:           &proverName,
		ProverID:         &proverID,
		ProofID:          &proofID,
		BatchNumber:      batchNum,
		BatchNumberFinal: batchNumFinal,
	}
	finalProof := &prover.FinalProof{}

	testCases := []struct {
		name    string
		setup   func(m mox, a *Aggregator)
		asserts func(a *Aggregator)
	}{
		{
			name: "Successfully settled on Agglayer",
			setup: func(m mox, a *Aggregator) {
				cfg := Config{
					SettlementBackend: AggLayer,
					AggLayerTxTimeout: types.Duration{Duration: time.Millisecond * 1},
				}
				a.cfg = cfg

				m.stateMock.On("GetBatch", mock.Anything, batchNumFinal, nil).Run(func(args mock.Arguments) {
				}).Return(&state.DBBatch{
					Batch: state.Batch{
						LocalExitRoot: common.Hash{},
						StateRoot:     common.Hash{},
					},
				}, nil).Once()

				m.etherman.On("GetRollupId").Return(uint32(1)).Once()
				testHash := common.BytesToHash([]byte("test hash"))
				m.aggLayerClientMock.On("SendTx", mock.Anything).Return(testHash, nil).Once()
				m.aggLayerClientMock.On("WaitTxToBeMined", testHash, mock.Anything).Return(nil).Once()
			},
			asserts: func(a *Aggregator) {
				assert.False(a.verifyingProof)
			},
		},
		{
			name: "Successfully settled on L1 (Direct)",
			setup: func(m mox, a *Aggregator) {
				senderAddr := common.BytesToAddress([]byte("sender address")).Hex()
				toAddr := common.BytesToAddress([]byte("to address"))
				data := []byte("data")
				cfg := Config{
					SettlementBackend: L1,
					SenderAddress:     senderAddr,
					GasOffset:         uint64(10),
				}
				a.cfg = cfg

				m.stateMock.On("GetBatch", mock.Anything, batchNumFinal, nil).Run(func(args mock.Arguments) {
				}).Return(&state.DBBatch{
					Batch: state.Batch{
						LocalExitRoot: common.Hash{},
						StateRoot:     common.Hash{},
					},
				}, nil).Once()

				m.etherman.On("BuildTrustedVerifyBatchesTxData", batchNum-1, batchNumFinal, mock.Anything, common.HexToAddress(senderAddr)).Return(&toAddr, data, nil).Once()
				m.ethTxManager.On("Add", mock.Anything, &toAddr, big.NewInt(0), data, a.cfg.GasOffset, (*ethTypes.BlobTxSidecar)(nil)).Return(nil, nil).Once()
				m.ethTxManager.On("ProcessPendingMonitoredTxs", mock.Anything, mock.Anything).Once()
			},
			asserts: func(a *Aggregator) {
				assert.False(a.verifyingProof)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stateMock := mocks.NewStateInterfaceMock(t)
			ethTxManager := mocks.NewEthTxManagerClientMock(t)
			etherman := mocks.NewEthermanMock(t)
			aggLayerClient := mocks.NewAgglayerClientInterfaceMock(t)

			curve := elliptic.P256()
			privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
			if err != nil {
				t.Fatal("Error generating key")
			}

			a := Aggregator{
				state:                   stateMock,
				etherman:                etherman,
				ethTxManager:            ethTxManager,
				aggLayerClient:          aggLayerClient,
				finalProof:              make(chan finalProofMsg),
				logger:                  log.GetDefaultLogger(),
				verifyingProof:          false,
				stateDBMutex:            &sync.Mutex{},
				timeSendFinalProofMutex: &sync.RWMutex{},
				sequencerPrivateKey:     privateKey,
			}
			a.ctx, a.exit = context.WithCancel(context.Background())

			m := mox{
				stateMock:          stateMock,
				ethTxManager:       ethTxManager,
				etherman:           etherman,
				aggLayerClientMock: aggLayerClient,
			}
			if tc.setup != nil {
				tc.setup(m, &a)
			}
			// send a final proof over the channel
			go func() {
				finalMsg := finalProofMsg{
					proverID:       proverID,
					recursiveProof: recursiveProof,
					finalProof:     finalProof,
				}
				a.finalProof <- finalMsg
				time.Sleep(1 * time.Second)
				a.exit()
			}()

			a.sendFinalProof()
			if tc.asserts != nil {
				tc.asserts(&a)
			}
		})
	}
}

func Test_sendFinalProofError(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	errTest := errors.New("test error")
	batchNum := uint64(23)
	batchNumFinal := uint64(42)
	sender := common.BytesToAddress([]byte("SenderAddress"))
	senderAddr := sender.Hex()

	recursiveProof := &state.Proof{
		Prover:           &proverName,
		ProverID:         &proverID,
		ProofID:          &proofID,
		BatchNumber:      batchNum,
		BatchNumberFinal: batchNumFinal,
	}
	finalProof := &prover.FinalProof{}

	testCases := []struct {
		name    string
		setup   func(mox, *Aggregator)
		asserts func(*Aggregator)
	}{
		{
			name: "Failed to settle on Agglayer: GetBatch error",
			setup: func(m mox, a *Aggregator) {
				m.stateMock.On("GetBatch", mock.Anything, batchNumFinal, nil).Run(func(args mock.Arguments) {
					// test is done, stop the sendFinalProof method
					fmt.Println("Stopping sendFinalProof")
					a.exit()
				}).Return(nil, errTest).Once()
			},
			asserts: func(a *Aggregator) {
				assert.False(a.verifyingProof)
			},
		},
		{
			name: "Failed to settle on Agglayer: SendTx error",
			setup: func(m mox, a *Aggregator) {
				cfg := Config{
					SettlementBackend: AggLayer,
				}
				a.cfg = cfg

				m.stateMock.On("GetBatch", mock.Anything, batchNumFinal, nil).Run(func(args mock.Arguments) {
				}).Return(&state.DBBatch{
					Batch: state.Batch{
						LocalExitRoot: common.Hash{},
						StateRoot:     common.Hash{},
					},
				}, nil).Once()

				m.etherman.On("GetRollupId").Return(uint32(1)).Once()
				m.aggLayerClientMock.On("SendTx", mock.Anything).Run(func(args mock.Arguments) {
					// test is done, stop the sendFinalProof method
					fmt.Println("Stopping sendFinalProof")
					a.exit()
				}).Return(nil, errTest).Once()
				m.stateMock.On("UpdateGeneratedProof", mock.Anything, mock.Anything, nil).Return(nil).Once()
			},
			asserts: func(a *Aggregator) {
				assert.False(a.verifyingProof)
			},
		},
		{
			name: "Failed to settle on Agglayer: WaitTxToBeMined error",
			setup: func(m mox, a *Aggregator) {
				cfg := Config{
					SettlementBackend: AggLayer,
					AggLayerTxTimeout: types.Duration{Duration: time.Millisecond * 1},
				}
				a.cfg = cfg

				m.stateMock.On("GetBatch", mock.Anything, batchNumFinal, nil).Run(func(args mock.Arguments) {
				}).Return(&state.DBBatch{
					Batch: state.Batch{
						LocalExitRoot: common.Hash{},
						StateRoot:     common.Hash{},
					},
				}, nil).Once()

				m.etherman.On("GetRollupId").Return(uint32(1)).Once()
				m.aggLayerClientMock.On("SendTx", mock.Anything).Return(common.Hash{}, nil).Once()
				m.aggLayerClientMock.On("WaitTxToBeMined", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fmt.Println("Stopping sendFinalProof")
					a.exit()
				}).Return(errTest)
				m.stateMock.On("UpdateGeneratedProof", mock.Anything, mock.Anything, nil).Return(nil).Once()
			},
			asserts: func(a *Aggregator) {
				assert.False(a.verifyingProof)
			},
		},
		{
			name: "Failed to settle on L1 (Direct): BuildTrustedVerifyBatchesTxData error",
			setup: func(m mox, a *Aggregator) {
				cfg := Config{
					SettlementBackend: L1,
					SenderAddress:     senderAddr,
				}
				a.cfg = cfg

				m.stateMock.On("GetBatch", mock.Anything, batchNumFinal, nil).Run(func(args mock.Arguments) {
				}).Return(&state.DBBatch{
					Batch: state.Batch{
						LocalExitRoot: common.Hash{},
						StateRoot:     common.Hash{},
					},
				}, nil).Once()

				m.etherman.On("BuildTrustedVerifyBatchesTxData", batchNum-1, batchNumFinal, mock.Anything, sender).Run(func(args mock.Arguments) {
					fmt.Println("Stopping sendFinalProof")
					a.exit()
				}).Return(nil, nil, errTest)
				m.stateMock.On("UpdateGeneratedProof", mock.Anything, recursiveProof, nil).Return(nil).Once()
			},
			asserts: func(a *Aggregator) {
				assert.False(a.verifyingProof)
			},
		},
		{
			name: "Failed to settle on L1 (Direct): Error Adding TX to ethTxManager",
			setup: func(m mox, a *Aggregator) {
				cfg := Config{
					SettlementBackend: L1,
					SenderAddress:     senderAddr,
					GasOffset:         uint64(10),
				}
				a.cfg = cfg

				m.stateMock.On("GetBatch", mock.Anything, batchNumFinal, nil).Run(func(args mock.Arguments) {
				}).Return(&state.DBBatch{
					Batch: state.Batch{
						LocalExitRoot: common.Hash{},
						StateRoot:     common.Hash{},
					},
				}, nil).Once()

				m.etherman.On("BuildTrustedVerifyBatchesTxData", batchNum-1, batchNumFinal, mock.Anything, sender).Return(nil, nil, nil).Once()
				m.ethTxManager.On("Add", mock.Anything, mock.Anything, big.NewInt(0), mock.Anything, a.cfg.GasOffset, (*ethTypes.BlobTxSidecar)(nil)).Run(func(args mock.Arguments) {
					fmt.Println("Stopping sendFinalProof")
					a.exit()
				}).Return(nil, errTest).Once()
				m.stateMock.On("UpdateGeneratedProof", mock.Anything, recursiveProof, nil).Return(nil).Once()
			},
			asserts: func(a *Aggregator) {
				assert.False(a.verifyingProof)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stateMock := mocks.NewStateInterfaceMock(t)
			ethTxManager := mocks.NewEthTxManagerClientMock(t)
			etherman := mocks.NewEthermanMock(t)
			aggLayerClient := mocks.NewAgglayerClientInterfaceMock(t)

			curve := elliptic.P256()
			privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
			if err != nil {
				t.Fatal("Error generating key")
			}

			a := Aggregator{
				state:                   stateMock,
				etherman:                etherman,
				ethTxManager:            ethTxManager,
				aggLayerClient:          aggLayerClient,
				finalProof:              make(chan finalProofMsg),
				logger:                  log.GetDefaultLogger(),
				verifyingProof:          false,
				stateDBMutex:            &sync.Mutex{},
				timeSendFinalProofMutex: &sync.RWMutex{},
				sequencerPrivateKey:     privateKey,
			}
			a.ctx, a.exit = context.WithCancel(context.Background())

			m := mox{
				stateMock:          stateMock,
				ethTxManager:       ethTxManager,
				etherman:           etherman,
				aggLayerClientMock: aggLayerClient,
			}
			if tc.setup != nil {
				tc.setup(m, &a)
			}
			// send a final proof over the channel
			go func() {
				finalMsg := finalProofMsg{
					proverID:       proverID,
					recursiveProof: recursiveProof,
					finalProof:     finalProof,
				}
				a.finalProof <- finalMsg
			}()

			a.sendFinalProof()
			if tc.asserts != nil {
				tc.asserts(&a)
			}
		})
	}
}

func Test_buildFinalProof(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	batchNum := uint64(23)
	batchNumFinal := uint64(42)
	recursiveProof := &state.Proof{
		ProverID:         &proverID,
		Proof:            "test proof",
		ProofID:          &proofID,
		BatchNumber:      batchNum,
		BatchNumberFinal: batchNumFinal,
	}

	testCases := []struct {
		name    string
		setup   func(mox, *Aggregator)
		asserts func(err error, fProof *prover.FinalProof)
	}{
		{
			name: "using real prover",
			setup: func(m mox, a *Aggregator) {
				finalProof := prover.FinalProof{
					Public: &prover.PublicInputsExtended{
						NewStateRoot:     []byte("StateRoot"),
						NewLocalExitRoot: []byte("LocalExitRoot"),
					},
				}

				m.proverMock.On("Name").Return("name").Once()
				m.proverMock.On("ID").Return("id").Once()
				m.proverMock.On("Addr").Return("addr").Once()
				m.proverMock.On("FinalProof", recursiveProof.Proof, a.cfg.SenderAddress).Return(&finalProofID, nil).Once()
				m.proverMock.On("WaitFinalProof", mock.Anything, finalProofID).Return(&finalProof, nil).Once()
			},
			asserts: func(err error, fProof *prover.FinalProof) {
				assert.NoError(err)
				assert.True(bytes.Equal([]byte("StateRoot"), fProof.Public.NewStateRoot), "State roots should be equal")
				assert.True(bytes.Equal([]byte("LocalExitRoot"), fProof.Public.NewLocalExitRoot), "LocalExit roots should be equal")
			},
		},
		{
			name: "using mock prover",
			setup: func(m mox, a *Aggregator) {
				finalProof := prover.FinalProof{
					Public: &prover.PublicInputsExtended{
						NewStateRoot:     []byte(mockedStateRoot),
						NewLocalExitRoot: []byte(mockedLocalExitRoot),
					},
				}

				finalDBBatch := &state.DBBatch{
					Batch: state.Batch{
						StateRoot:     common.BytesToHash([]byte("mock StateRoot")),
						LocalExitRoot: common.BytesToHash([]byte("mock LocalExitRoot")),
					},
				}

				m.proverMock.On("Name").Return("name").Once()
				m.proverMock.On("ID").Return("id").Once()
				m.proverMock.On("Addr").Return("addr").Once()
				m.proverMock.On("FinalProof", recursiveProof.Proof, a.cfg.SenderAddress).Return(&finalProofID, nil).Once()
				m.proverMock.On("WaitFinalProof", mock.Anything, finalProofID).Return(&finalProof, nil).Once()
				m.stateMock.On("GetBatch", mock.Anything, batchNumFinal, nil).Return(finalDBBatch, nil).Once()
			},
			asserts: func(err error, fProof *prover.FinalProof) {
				assert.NoError(err)
				expStateRoot := common.BytesToHash([]byte("mock StateRoot"))
				expLocalExitRoot := common.BytesToHash([]byte("mock LocalExitRoot"))
				assert.True(bytes.Equal(expStateRoot.Bytes(), fProof.Public.NewStateRoot), "State roots should be equal")
				assert.True(bytes.Equal(expLocalExitRoot.Bytes(), fProof.Public.NewLocalExitRoot), "LocalExit roots should be equal")
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			proverMock := mocks.NewProverInterfaceMock(t)
			stateMock := mocks.NewStateInterfaceMock(t)
			m := mox{
				proverMock: proverMock,
				stateMock:  stateMock,
			}
			a := Aggregator{
				state:  stateMock,
				logger: log.GetDefaultLogger(),
				cfg: Config{
					SenderAddress: common.BytesToAddress([]byte("from")).Hex(),
				},
			}

			tc.setup(m, &a)
			fProof, err := a.buildFinalProof(context.Background(), proverMock, recursiveProof)
			tc.asserts(err, fProof)
		})
	}
}

func Test_tryBuildFinalProof(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	errTest := errors.New("test error")
	from := common.BytesToAddress([]byte("from"))
	cfg := Config{
		VerifyProofInterval:        types.Duration{Duration: time.Millisecond * 1},
		TxProfitabilityCheckerType: ProfitabilityAcceptAll,
		SenderAddress:              from.Hex(),
	}
	latestVerifiedBatchNum := uint64(22)
	batchNum := uint64(23)
	batchNumFinal := uint64(42)
	finalProof := prover.FinalProof{
		Proof: "",
		Public: &prover.PublicInputsExtended{
			NewStateRoot:     []byte("newStateRoot"),
			NewLocalExitRoot: []byte("newLocalExitRoot"),
		},
	}
	proofToVerify := state.Proof{
		ProofID:          &proofID,
		Proof:            proof,
		BatchNumber:      batchNum,
		BatchNumberFinal: batchNumFinal,
	}
	invalidProof := state.Proof{
		ProofID:          &proofID,
		Proof:            proof,
		BatchNumber:      uint64(123),
		BatchNumberFinal: uint64(456),
	}

	proverCtx := context.WithValue(context.Background(), "owner", "prover") //nolint:staticcheck
	matchProverCtxFn := func(ctx context.Context) bool { return ctx.Value("owner") == "prover" }
	matchAggregatorCtxFn := func(ctx context.Context) bool { return ctx.Value("owner") == "aggregator" }
	testCases := []struct {
		name           string
		proof          *state.Proof
		setup          func(mox, *Aggregator)
		asserts        func(bool, *Aggregator, error)
		assertFinalMsg func(*finalProofMsg)
	}{
		{
			name: "can't verify proof (verifyingProof = true)",
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Once()
				m.proverMock.On("ID").Return(proverID).Once()
				m.proverMock.On("Addr").Return("addr").Once()
				a.verifyingProof = true
			},
			asserts: func(result bool, a *Aggregator, err error) {
				a.verifyingProof = false // reset
				assert.False(result)
				assert.NoError(err)
			},
		},
		{
			name: "can't verify proof (veryfy time not reached yet)",
			setup: func(m mox, a *Aggregator) {
				a.timeSendFinalProof = time.Now().Add(10 * time.Second)
				m.proverMock.On("Name").Return(proverName).Once()
				m.proverMock.On("ID").Return(proverID).Once()
				m.proverMock.On("Addr").Return("addr").Once()
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.NoError(err)
			},
		},
		{
			name: "nil proof, error requesting the proof triggers defer",
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Twice()
				m.proverMock.On("ID").Return(proverID).Twice()
				m.proverMock.On("Addr").Return("addr").Twice()
				m.etherman.On("GetLatestVerifiedBatchNum").Return(latestVerifiedBatchNum, nil).Once()
				m.stateMock.On("GetProofReadyToVerify", mock.MatchedBy(matchProverCtxFn), latestVerifiedBatchNum, nil).Return(&proofToVerify, nil).Once()
				proofGeneratingTrueCall := m.stateMock.On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proofToVerify, nil).Return(nil).Once()
				m.proverMock.On("FinalProof", proofToVerify.Proof, from.String()).Return(nil, errTest).Once()
				m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchAggregatorCtxFn), &proofToVerify, nil).
					Return(nil).
					Once().
					NotBefore(proofGeneratingTrueCall)
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.ErrorIs(err, errTest)
			},
		},
		{
			name: "nil proof, error building the proof triggers defer",
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Twice()
				m.proverMock.On("ID").Return(proverID).Twice()
				m.proverMock.On("Addr").Return("addr").Twice()
				m.etherman.On("GetLatestVerifiedBatchNum").Return(latestVerifiedBatchNum, nil).Once()
				m.stateMock.On("GetProofReadyToVerify", mock.MatchedBy(matchProverCtxFn), latestVerifiedBatchNum, nil).Return(&proofToVerify, nil).Once()
				proofGeneratingTrueCall := m.stateMock.On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proofToVerify, nil).Return(nil).Once()
				m.proverMock.On("FinalProof", proofToVerify.Proof, from.String()).Return(&finalProofID, nil).Once()
				m.proverMock.On("WaitFinalProof", mock.MatchedBy(matchProverCtxFn), finalProofID).Return(nil, errTest).Once()
				m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchAggregatorCtxFn), &proofToVerify, nil).
					Return(nil).
					Once().
					NotBefore(proofGeneratingTrueCall)
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.ErrorIs(err, errTest)
			},
		},
		{
			name: "nil proof, generic error from GetProofReadyToVerify",
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Once()
				m.proverMock.On("ID").Return(proverID).Once()
				m.proverMock.On("Addr").Return(proverID).Once()
				m.etherman.On("GetLatestVerifiedBatchNum").Return(latestVerifiedBatchNum, nil).Once()
				m.stateMock.On("GetProofReadyToVerify", mock.MatchedBy(matchProverCtxFn), latestVerifiedBatchNum, nil).Return(nil, errTest).Once()
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.ErrorIs(err, errTest)
			},
		},
		{
			name: "nil proof, ErrNotFound from GetProofReadyToVerify",
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Once()
				m.proverMock.On("ID").Return(proverID).Once()
				m.proverMock.On("Addr").Return(proverID).Once()
				m.etherman.On("GetLatestVerifiedBatchNum").Return(latestVerifiedBatchNum, nil).Once()
				m.stateMock.On("GetProofReadyToVerify", mock.MatchedBy(matchProverCtxFn), latestVerifiedBatchNum, nil).Return(nil, state.ErrNotFound).Once()
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.NoError(err)
			},
		},
		{
			name: "nil proof gets a proof ready to verify",
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Twice()
				m.proverMock.On("ID").Return(proverID).Twice()
				m.proverMock.On("Addr").Return(proverID).Twice()
				m.etherman.On("GetLatestVerifiedBatchNum").Return(latestVerifiedBatchNum, nil).Once()
				m.stateMock.On("GetProofReadyToVerify", mock.MatchedBy(matchProverCtxFn), latestVerifiedBatchNum, nil).Return(&proofToVerify, nil).Once()
				m.stateMock.On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proofToVerify, nil).Return(nil).Once()
				m.proverMock.On("FinalProof", proofToVerify.Proof, from.String()).Return(&finalProofID, nil).Once()
				m.proverMock.On("WaitFinalProof", mock.MatchedBy(matchProverCtxFn), finalProofID).Return(&finalProof, nil).Once()
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.True(result)
				assert.NoError(err)
			},
			assertFinalMsg: func(msg *finalProofMsg) {
				assert.Equal(finalProof.Proof, msg.finalProof.Proof)
				assert.Equal(finalProof.Public.NewStateRoot, msg.finalProof.Public.NewStateRoot)
				assert.Equal(finalProof.Public.NewLocalExitRoot, msg.finalProof.Public.NewLocalExitRoot)
			},
		},
		{
			name:  "error checking if proof is a complete sequence",
			proof: &proofToVerify,
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Once()
				m.proverMock.On("ID").Return(proverID).Once()
				m.proverMock.On("Addr").Return(proverID).Once()
				m.etherman.On("GetLatestVerifiedBatchNum").Return(latestVerifiedBatchNum, nil).Once()
				m.stateMock.On("CheckProofContainsCompleteSequences", mock.MatchedBy(matchProverCtxFn), &proofToVerify, nil).Return(false, errTest).Once()
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.ErrorIs(err, errTest)
			},
		},
		{
			name:  "invalid proof (not consecutive to latest verified batch) rejected",
			proof: &invalidProof,
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Once()
				m.proverMock.On("ID").Return(proverID).Once()
				m.proverMock.On("Addr").Return(proverID).Once()
				m.etherman.On("GetLatestVerifiedBatchNum").Return(latestVerifiedBatchNum, nil).Once()
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.NoError(err)
			},
		},
		{
			name:  "invalid proof (not a complete sequence) rejected",
			proof: &proofToVerify,
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Once()
				m.proverMock.On("ID").Return(proverID).Once()
				m.proverMock.On("Addr").Return(proverID).Once()
				m.etherman.On("GetLatestVerifiedBatchNum").Return(latestVerifiedBatchNum, nil).Once()
				m.stateMock.On("CheckProofContainsCompleteSequences", mock.MatchedBy(matchProverCtxFn), &proofToVerify, nil).Return(false, nil).Once()
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.NoError(err)
			},
		},
		{
			name:  "valid proof ok",
			proof: &proofToVerify,
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Twice()
				m.proverMock.On("ID").Return(proverID).Twice()
				m.proverMock.On("Addr").Return(proverID).Twice()
				m.etherman.On("GetLatestVerifiedBatchNum").Return(latestVerifiedBatchNum, nil).Once()
				m.stateMock.On("CheckProofContainsCompleteSequences", mock.MatchedBy(matchProverCtxFn), &proofToVerify, nil).Return(true, nil).Once()
				m.proverMock.On("FinalProof", proofToVerify.Proof, from.String()).Return(&finalProofID, nil).Once()
				m.proverMock.On("WaitFinalProof", mock.MatchedBy(matchProverCtxFn), finalProofID).Return(&finalProof, nil).Once()
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.True(result)
				assert.NoError(err)
			},
			assertFinalMsg: func(msg *finalProofMsg) {
				assert.Equal(finalProof.Proof, msg.finalProof.Proof)
				assert.Equal(finalProof.Public.NewStateRoot, msg.finalProof.Public.NewStateRoot)
				assert.Equal(finalProof.Public.NewLocalExitRoot, msg.finalProof.Public.NewLocalExitRoot)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stateMock := mocks.NewStateInterfaceMock(t)
			ethTxManager := mocks.NewEthTxManagerClientMock(t)
			etherman := mocks.NewEthermanMock(t)
			proverMock := mocks.NewProverInterfaceMock(t)

			a := Aggregator{
				cfg:                     cfg,
				state:                   stateMock,
				etherman:                etherman,
				ethTxManager:            ethTxManager,
				logger:                  log.GetDefaultLogger(),
				stateDBMutex:            &sync.Mutex{},
				timeSendFinalProofMutex: &sync.RWMutex{},
				timeCleanupLockedProofs: cfg.CleanupLockedProofsInterval,
				finalProof:              make(chan finalProofMsg),
			}

			aggregatorCtx := context.WithValue(context.Background(), "owner", "aggregator") //nolint:staticcheck
			a.ctx, a.exit = context.WithCancel(aggregatorCtx)
			m := mox{
				stateMock:    stateMock,
				ethTxManager: ethTxManager,
				etherman:     etherman,
				proverMock:   proverMock,
			}
			if tc.setup != nil {
				tc.setup(m, &a)
			}

			var finalProofReceived bool
			if tc.assertFinalMsg != nil {
				// Start a goroutine to listen for the final proof and trigger the assertion
				finalProofReceived = false
				go func() {
					msg := <-a.finalProof
					tc.assertFinalMsg(&msg)
					finalProofReceived = true
				}()
			}

			result, err := a.tryBuildFinalProof(proverCtx, proverMock, tc.proof)

			if tc.asserts != nil {
				tc.asserts(result, &a, err)
			}

			if tc.assertFinalMsg != nil {
				// Wait for final proof to be received with a timeout
				timeout := time.After(time.Second)
				for !finalProofReceived {
					select {
					case <-timeout:
						t.Fatal("Final proof not received before timeout")
					default:
						time.Sleep(10 * time.Millisecond) // Sleep briefly to prevent busy waiting
					}
				}
			}
		})
	}
}

func Test_tryAggregateProofs(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	errTest := errors.New("test error")
	cfg := Config{
		VerifyProofInterval: types.Duration{Duration: time.Millisecond * 1},
	}

	recursiveProof := "recursiveProof"
	proverCtx := context.WithValue(context.Background(), "owner", "prover") //nolint:staticcheck
	matchProverCtxFn := func(ctx context.Context) bool { return ctx.Value("owner") == "prover" }
	matchAggregatorCtxFn := func(ctx context.Context) bool { return ctx.Value("owner") == "aggregator" }
	batchNum := uint64(23)
	batchNumFinal := uint64(42)
	proof1 := state.Proof{
		Proof:       "proof1",
		BatchNumber: batchNum,
	}
	proof2 := state.Proof{
		Proof:            "proof2",
		BatchNumberFinal: batchNumFinal,
	}
	testCases := []struct {
		name    string
		setup   func(mox, *Aggregator)
		asserts func(bool, *Aggregator, error)
	}{
		{
			name: "getAndLockProofsToAggregate returns generic error",
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Twice()
				m.proverMock.On("ID").Return(proverID).Twice()
				m.proverMock.On("Addr").Return("addr")
				m.stateMock.On("GetProofsToAggregate", mock.MatchedBy(matchProverCtxFn), nil).Return(nil, nil, errTest).Once()
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.ErrorIs(err, errTest)
			},
		},
		{
			name: "getAndLockProofsToAggregate returns ErrNotFound",
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Twice()
				m.proverMock.On("ID").Return(proverID).Twice()
				m.proverMock.On("Addr").Return("addr")
				m.stateMock.On("GetProofsToAggregate", mock.MatchedBy(matchProverCtxFn), nil).Return(nil, nil, state.ErrNotFound).Once()
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.NoError(err)
			},
		},
		{
			name: "getAndLockProofsToAggregate error updating proofs",
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Twice()
				m.proverMock.On("ID").Return(proverID).Twice()
				m.proverMock.On("Addr").Return("addr")
				dbTx := &mocks.DbTxMock{}
				dbTx.On("Rollback", mock.MatchedBy(matchProverCtxFn)).Return(nil).Once()
				m.stateMock.On("BeginStateTransaction", mock.MatchedBy(matchProverCtxFn)).Return(dbTx, nil).Once()
				m.stateMock.On("GetProofsToAggregate", mock.MatchedBy(matchProverCtxFn), nil).Return(&proof1, &proof2, nil).Once()
				m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proof1, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.NotNil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(errTest).
					Once()
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.ErrorIs(err, errTest)
			},
		},
		{
			name: "AggregatedProof error",
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Twice()
				m.proverMock.On("ID").Return(proverID).Twice()
				m.proverMock.On("Addr").Return("addr")
				dbTx := &mocks.DbTxMock{}
				lockProofsTxBegin := m.stateMock.On("BeginStateTransaction", mock.MatchedBy(matchProverCtxFn)).Return(dbTx, nil).Once()
				lockProofsTxCommit := dbTx.On("Commit", mock.MatchedBy(matchProverCtxFn)).Return(nil).Once()
				m.stateMock.On("GetProofsToAggregate", mock.MatchedBy(matchProverCtxFn), nil).Return(&proof1, &proof2, nil).Once()
				proof1GeneratingTrueCall := m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proof1, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.NotNil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once()
				proof2GeneratingTrueCall := m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proof2, dbTx).
					Run(func(args mock.Arguments) {
						// Use a type assertion with a check
						proofArg, ok := args[1].(*state.Proof)
						if !ok {
							assert.Fail("Expected argument of type *state.Proof")
						}
						assert.NotNil(proofArg.GeneratingSince)
					}).
					Return(nil).
					Once()
				m.proverMock.On("AggregatedProof", proof1.Proof, proof2.Proof).Return(nil, errTest).Once()
				m.stateMock.On("BeginStateTransaction", mock.MatchedBy(matchAggregatorCtxFn)).Return(dbTx, nil).Once().NotBefore(lockProofsTxBegin)
				m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchAggregatorCtxFn), &proof1, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						if !ok {
							assert.Fail("Expected argument of type *state.Proof")
						}
						assert.Nil(proofArg.GeneratingSince)
					}).
					Return(nil).
					Once().
					NotBefore(proof1GeneratingTrueCall)
				m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchAggregatorCtxFn), &proof2, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						if !ok {
							assert.Fail("Expected argument of type *state.Proof")
						}
						assert.Nil(proofArg.GeneratingSince, "Expected GeneratingSince to be nil")
					}).
					Return(nil).
					Once().
					NotBefore(proof2GeneratingTrueCall)
				dbTx.On("Commit", mock.MatchedBy(matchAggregatorCtxFn)).Return(nil).Once().NotBefore(lockProofsTxCommit)
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.ErrorIs(err, errTest)
			},
		},
		{
			name: "WaitRecursiveProof prover error",
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Twice()
				m.proverMock.On("ID").Return(proverID).Twice()
				m.proverMock.On("Addr").Return("addr")
				dbTx := &mocks.DbTxMock{}
				lockProofsTxBegin := m.stateMock.On("BeginStateTransaction", mock.MatchedBy(matchProverCtxFn)).Return(dbTx, nil).Once()
				lockProofsTxCommit := dbTx.On("Commit", mock.MatchedBy(matchProverCtxFn)).Return(nil).Once()
				m.stateMock.On("GetProofsToAggregate", mock.MatchedBy(matchProverCtxFn), nil).Return(&proof1, &proof2, nil).Once()
				proof1GeneratingTrueCall := m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proof1, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						if !ok {
							assert.Fail("Expected argument of type *state.Proof")
						}
						assert.NotNil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once()
				proof2GeneratingTrueCall := m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proof2, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.NotNil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once()
				m.proverMock.On("AggregatedProof", proof1.Proof, proof2.Proof).Return(&proofID, nil).Once()
				m.proverMock.On("WaitRecursiveProof", mock.MatchedBy(matchProverCtxFn), proofID).Return("", common.Hash{}, errTest).Once()
				m.stateMock.On("BeginStateTransaction", mock.MatchedBy(matchAggregatorCtxFn)).Return(dbTx, nil).Once().NotBefore(lockProofsTxBegin)
				m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchAggregatorCtxFn), &proof1, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.Nil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once().
					NotBefore(proof1GeneratingTrueCall)
				m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchAggregatorCtxFn), &proof2, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.Nil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once().
					NotBefore(proof2GeneratingTrueCall)
				dbTx.On("Commit", mock.MatchedBy(matchAggregatorCtxFn)).Return(nil).Once().NotBefore(lockProofsTxCommit)
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.ErrorIs(err, errTest)
			},
		},
		{
			name: "unlockProofsToAggregate error after WaitRecursiveProof prover error",
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Twice()
				m.proverMock.On("ID").Return(proverID).Twice()
				m.proverMock.On("Addr").Return(proverID)
				dbTx := &mocks.DbTxMock{}
				lockProofsTxBegin := m.stateMock.On("BeginStateTransaction", mock.MatchedBy(matchProverCtxFn)).Return(dbTx, nil).Once()
				dbTx.On("Commit", mock.MatchedBy(matchProverCtxFn)).Return(nil).Once()
				m.stateMock.On("GetProofsToAggregate", mock.MatchedBy(matchProverCtxFn), nil).Return(&proof1, &proof2, nil).Once()
				proof1GeneratingTrueCall := m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proof1, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.NotNil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once()
				m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proof2, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.NotNil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once()
				m.proverMock.On("AggregatedProof", proof1.Proof, proof2.Proof).Return(&proofID, nil).Once()
				m.proverMock.On("WaitRecursiveProof", mock.MatchedBy(matchProverCtxFn), proofID).Return("", common.Hash{}, errTest).Once()
				m.stateMock.On("BeginStateTransaction", mock.MatchedBy(matchAggregatorCtxFn)).Return(dbTx, nil).Once().NotBefore(lockProofsTxBegin)
				m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchAggregatorCtxFn), &proof1, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.Nil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(errTest).
					Once().
					NotBefore(proof1GeneratingTrueCall)
				dbTx.On("Rollback", mock.MatchedBy(matchAggregatorCtxFn)).Return(nil).Once()
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.ErrorIs(err, errTest)
			},
		},
		{
			name: "rollback after DeleteGeneratedProofs error in db transaction",
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Twice()
				m.proverMock.On("ID").Return(proverID).Twice()
				m.proverMock.On("Addr").Return("addr")
				dbTx := &mocks.DbTxMock{}
				lockProofsTxBegin := m.stateMock.On("BeginStateTransaction", mock.MatchedBy(matchProverCtxFn)).Return(dbTx, nil).Twice()
				lockProofsTxCommit := dbTx.On("Commit", mock.MatchedBy(matchProverCtxFn)).Return(nil).Once()
				m.stateMock.On("GetProofsToAggregate", mock.MatchedBy(matchProverCtxFn), nil).Return(&proof1, &proof2, nil).Once()
				proof1GeneratingTrueCall := m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proof1, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.NotNil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once()
				proof2GeneratingTrueCall := m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proof2, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.NotNil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once()
				m.proverMock.On("AggregatedProof", proof1.Proof, proof2.Proof).Return(&proofID, nil).Once()
				m.proverMock.On("WaitRecursiveProof", mock.MatchedBy(matchProverCtxFn), proofID).Return(recursiveProof, common.Hash{}, nil).Once()
				m.stateMock.On("DeleteGeneratedProofs", mock.MatchedBy(matchProverCtxFn), proof1.BatchNumber, proof2.BatchNumberFinal, dbTx).Return(errTest).Once()
				dbTx.On("Rollback", mock.MatchedBy(matchProverCtxFn)).Return(nil).Once()
				m.stateMock.On("BeginStateTransaction", mock.MatchedBy(matchAggregatorCtxFn)).Return(dbTx, nil).Once().NotBefore(lockProofsTxBegin)
				m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchAggregatorCtxFn), &proof1, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.Nil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once().
					NotBefore(proof1GeneratingTrueCall)
				m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchAggregatorCtxFn), &proof2, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.Nil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once().
					NotBefore(proof2GeneratingTrueCall)
				dbTx.On("Commit", mock.MatchedBy(matchAggregatorCtxFn)).Return(nil).Once().NotBefore(lockProofsTxCommit)
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.ErrorIs(err, errTest)
			},
		},
		{
			name: "rollback after AddGeneratedProof error in db transaction",
			setup: func(m mox, a *Aggregator) {
				m.proverMock.On("Name").Return(proverName).Twice()
				m.proverMock.On("ID").Return(proverID).Twice()
				m.proverMock.On("Addr").Return("addr")
				dbTx := &mocks.DbTxMock{}
				lockProofsTxBegin := m.stateMock.On("BeginStateTransaction", mock.MatchedBy(matchProverCtxFn)).Return(dbTx, nil).Twice()
				lockProofsTxCommit := dbTx.On("Commit", mock.MatchedBy(matchProverCtxFn)).Return(nil).Once()
				m.stateMock.On("GetProofsToAggregate", mock.MatchedBy(matchProverCtxFn), nil).Return(&proof1, &proof2, nil).Once()
				proof1GeneratingTrueCall := m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proof1, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.NotNil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once()
				proof2GeneratingTrueCall := m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proof2, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.NotNil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once()
				m.proverMock.On("AggregatedProof", proof1.Proof, proof2.Proof).Return(&proofID, nil).Once()
				m.proverMock.On("WaitRecursiveProof", mock.MatchedBy(matchProverCtxFn), proofID).Return(recursiveProof, common.Hash{}, nil).Once()
				m.stateMock.On("DeleteGeneratedProofs", mock.MatchedBy(matchProverCtxFn), proof1.BatchNumber, proof2.BatchNumberFinal, dbTx).Return(nil).Once()
				m.stateMock.On("AddGeneratedProof", mock.MatchedBy(matchProverCtxFn), mock.Anything, dbTx).Return(errTest).Once()
				dbTx.On("Rollback", mock.MatchedBy(matchProverCtxFn)).Return(nil).Once()
				m.stateMock.On("BeginStateTransaction", mock.MatchedBy(matchAggregatorCtxFn)).Return(dbTx, nil).Once().NotBefore(lockProofsTxBegin)
				m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchAggregatorCtxFn), &proof1, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.Nil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once().
					NotBefore(proof1GeneratingTrueCall)
				m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchAggregatorCtxFn), &proof2, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.Nil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once().
					NotBefore(proof2GeneratingTrueCall)
				dbTx.On("Commit", mock.MatchedBy(matchAggregatorCtxFn)).Return(nil).Once().NotBefore(lockProofsTxCommit)
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.False(result)
				assert.ErrorIs(err, errTest)
			},
		},
		{
			name: "time to send final, state error",
			setup: func(m mox, a *Aggregator) {
				a.cfg.VerifyProofInterval = types.Duration{Duration: time.Nanosecond}
				m.proverMock.On("Name").Return(proverName).Times(3)
				m.proverMock.On("ID").Return(proverID).Times(3)
				m.proverMock.On("Addr").Return("addr")
				dbTx := &mocks.DbTxMock{}
				m.stateMock.On("BeginStateTransaction", mock.MatchedBy(matchProverCtxFn)).Return(dbTx, nil).Twice()
				dbTx.On("Commit", mock.MatchedBy(matchProverCtxFn)).Return(nil).Twice()
				m.stateMock.On("GetProofsToAggregate", mock.MatchedBy(matchProverCtxFn), nil).Return(&proof1, &proof2, nil).Once()
				m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proof1, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.NotNil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once()
				m.stateMock.
					On("UpdateGeneratedProof", mock.MatchedBy(matchProverCtxFn), &proof2, dbTx).
					Run(func(args mock.Arguments) {
						proofArg, ok := args[1].(*state.Proof)
						assert.True(ok, "Expected argument of type *state.Proof")
						assert.NotNil(proofArg.GeneratingSince, "Expected GeneratingSince to be not nil")
					}).
					Return(nil).
					Once()

				m.proverMock.On("AggregatedProof", proof1.Proof, proof2.Proof).Return(&proofID, nil).Once()
				m.proverMock.On("WaitRecursiveProof", mock.MatchedBy(matchProverCtxFn), proofID).Return(recursiveProof, common.Hash{}, nil).Once()
				m.stateMock.On("DeleteGeneratedProofs", mock.MatchedBy(matchProverCtxFn), proof1.BatchNumber, proof2.BatchNumberFinal, dbTx).Return(nil).Once()
				expectedInputProver := map[string]interface{}{
					"recursive_proof_1": proof1.Proof,
					"recursive_proof_2": proof2.Proof,
				}
				b, err := json.Marshal(expectedInputProver)
				require.NoError(err)
				m.stateMock.On("AddGeneratedProof", mock.MatchedBy(matchProverCtxFn), mock.Anything, dbTx).Run(
					func(args mock.Arguments) {
						proof, ok := args[1].(*state.Proof)
						if !ok {
							t.Fatalf("expected args[1] to be of type *state.Proof, got %T", args[1])
						}
						assert.Equal(proof1.BatchNumber, proof.BatchNumber)
						assert.Equal(proof2.BatchNumberFinal, proof.BatchNumberFinal)
						assert.Equal(&proverName, proof.Prover)
						assert.Equal(&proverID, proof.ProverID)
						assert.Equal(string(b), proof.InputProver)
						assert.Equal(recursiveProof, proof.Proof)
						assert.InDelta(time.Now().Unix(), proof.GeneratingSince.Unix(), float64(time.Second))
					},
				).Return(nil).Once()

				m.etherman.On("GetLatestVerifiedBatchNum").Return(uint64(42), errTest).Once()
				m.stateMock.On("UpdateGeneratedProof", mock.MatchedBy(matchAggregatorCtxFn), mock.Anything, nil).Run(
					func(args mock.Arguments) {
						proof, ok := args[1].(*state.Proof)
						if !ok {
							t.Fatalf("expected args[1] to be of type *state.Proof, got %T", args[1])
						}
						assert.Equal(proof1.BatchNumber, proof.BatchNumber)
						assert.Equal(proof2.BatchNumberFinal, proof.BatchNumberFinal)
						assert.Equal(&proverName, proof.Prover)
						assert.Equal(&proverID, proof.ProverID)
						assert.Equal(string(b), proof.InputProver)
						assert.Equal(recursiveProof, proof.Proof)
						assert.Nil(proof.GeneratingSince)
					},
				).Return(nil).Once()
			},
			asserts: func(result bool, a *Aggregator, err error) {
				assert.True(result)
				assert.NoError(err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stateMock := mocks.NewStateInterfaceMock(t)
			ethTxManager := mocks.NewEthTxManagerClientMock(t)
			etherman := mocks.NewEthermanMock(t)
			proverMock := mocks.NewProverInterfaceMock(t)
			a := Aggregator{
				cfg:                     cfg,
				state:                   stateMock,
				etherman:                etherman,
				ethTxManager:            ethTxManager,
				logger:                  log.GetDefaultLogger(),
				stateDBMutex:            &sync.Mutex{},
				timeSendFinalProofMutex: &sync.RWMutex{},
				timeCleanupLockedProofs: cfg.CleanupLockedProofsInterval,
				finalProof:              make(chan finalProofMsg),
			}
			aggregatorCtx := context.WithValue(context.Background(), "owner", "aggregator") //nolint:staticcheck
			a.ctx, a.exit = context.WithCancel(aggregatorCtx)
			m := mox{
				stateMock:    stateMock,
				ethTxManager: ethTxManager,
				etherman:     etherman,
				proverMock:   proverMock,
			}
			if tc.setup != nil {
				tc.setup(m, &a)
			}
			a.resetVerifyProofTime()

			result, err := a.tryAggregateProofs(proverCtx, proverMock)

			if tc.asserts != nil {
				tc.asserts(result, &a, err)
			}
		})
	}
}
