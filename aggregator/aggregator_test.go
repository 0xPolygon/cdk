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

	"github.com/0xPolygon/cdk/agglayer"
	mocks "github.com/0xPolygon/cdk/aggregator/mocks"
	"github.com/0xPolygon/cdk/aggregator/prover"
	"github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/log"
	rpctypes "github.com/0xPolygon/cdk/rpc/types"
	"github.com/0xPolygon/cdk/state"
	"github.com/0xPolygonHermez/zkevm-synchronizer-l1/synchronizer"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	proofID    = "proofId"
	proof      = "proof"
	proverName = "proverName"
	proverID   = "proverID"
)

const (
	ownerProver     = "prover"
	ownerAggregator = "aggregator"

	// changeL2Block + deltaTimeStamp + indexL1InfoTree
	codedL2BlockHeader = "0b73e6af6f00000001"
	// 2 x [ tx coded in RLP + r,s,v,efficiencyPercentage]
	codedRLP2Txs1 = "ee02843b9aca00830186a0944d5cf5032b2a844602278b01199ed191a86c93ff88016345785d8a0000808203e88080bff0e780ba7db409339fd3f71969fa2cbf1b8535f6c725a1499d3318d3ef9c2b6340ddfab84add2c188f9efddb99771db1fe621c981846394ea4f035c85bcdd51bffee03843b9aca00830186a0944d5cf5032b2a844602278b01199ed191a86c93ff88016345785d8a0000808203e880805b346aa02230b22e62f73608de9ff39a162a6c24be9822209c770e3685b92d0756d5316ef954eefc58b068231ccea001fb7ac763ebe03afd009ad71cab36861e1bff"
	codedL2Block1 = codedL2BlockHeader + codedRLP2Txs1
)

type mox struct {
	stateMock          *mocks.StateInterfaceMock
	ethTxManager       *mocks.EthTxManagerClientMock
	etherman           *mocks.EthermanMock
	proverMock         *mocks.ProverInterfaceMock
	aggLayerClientMock *agglayer.AgglayerClientMock
	synchronizerMock   *mocks.SynchronizerInterfaceMock
	rpcMock            *mocks.RPCInterfaceMock
}

func WaitUntil(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatalf("WaitGroup not done, test time expired after %s", timeout)
	}
}

func Test_Start(t *testing.T) {
	mockState := new(mocks.StateInterfaceMock)
	mockL1Syncr := new(mocks.SynchronizerInterfaceMock)
	mockEtherman := new(mocks.EthermanMock)
	mockEthTxManager := new(mocks.EthTxManagerClientMock)

	mockL1Syncr.On("Sync", mock.Anything).Return(nil)
	mockEtherman.On("GetLatestVerifiedBatchNum").Return(uint64(90), nil).Once()
	mockEtherman.On("GetBatchAccInputHash", mock.Anything, uint64(90)).Return(common.Hash{}, nil).Once()
	mockState.On("DeleteGeneratedProofs", mock.Anything, uint64(90), mock.Anything, nil).Return(nil).Once()
	mockState.On("CleanupLockedProofs", mock.Anything, "", nil).Return(int64(0), nil)

	mockEthTxManager.On("Start").Return(nil)

	ctx := context.Background()
	a := &Aggregator{
		state:                   mockState,
		logger:                  log.GetDefaultLogger(),
		halted:                  atomic.Bool{},
		l1Syncr:                 mockL1Syncr,
		etherman:                mockEtherman,
		ethTxManager:            mockEthTxManager,
		ctx:                     ctx,
		stateDBMutex:            &sync.Mutex{},
		timeSendFinalProofMutex: &sync.RWMutex{},
		timeCleanupLockedProofs: types.Duration{Duration: 5 * time.Second},
		accInputHashes:          make(map[uint64]common.Hash),
		accInputHashesMutex:     &sync.Mutex{},
	}
	go func() {
		err := a.Start()
		require.NoError(t, err)
	}()
	time.Sleep(time.Second)
	a.ctx.Done()
	time.Sleep(time.Second)
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
	mockState.On("DeleteGeneratedProofs", mock.Anything, mock.Anything, mock.Anything, nil).Return(nil).Once()
	mockState.On("DeleteUngeneratedProofs", mock.Anything, nil).Return(nil).Once()

	go a.handleReorg(reorgData)
	time.Sleep(3 * time.Second)

	assert.True(t, a.halted.Load())
	mockState.AssertExpectations(t)
	mockL1Syncr.AssertExpectations(t)
}

func Test_handleRollbackBatches(t *testing.T) {
	t.Parallel()

	mockEtherman := new(mocks.EthermanMock)
	mockState := new(mocks.StateInterfaceMock)

	// Test data
	rollbackData := synchronizer.RollbackBatchesData{
		LastBatchNumber: 100,
	}

	mockEtherman.On("GetLatestVerifiedBatchNum").Return(uint64(90), nil).Once()
	mockEtherman.On("GetBatchAccInputHash", mock.Anything, uint64(90)).Return(common.Hash{}, nil).Once()
	mockState.On("DeleteUngeneratedProofs", mock.Anything, mock.Anything).Return(nil).Once()
	mockState.On("DeleteGeneratedProofs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	a := Aggregator{
		ctx:                 context.Background(),
		etherman:            mockEtherman,
		state:               mockState,
		logger:              log.GetDefaultLogger(),
		halted:              atomic.Bool{},
		accInputHashes:      make(map[uint64]common.Hash),
		accInputHashesMutex: &sync.Mutex{},
	}

	a.halted.Store(false)
	a.handleRollbackBatches(rollbackData)

	assert.False(t, a.halted.Load())
	mockEtherman.AssertExpectations(t)
	mockState.AssertExpectations(t)
}

func Test_handleRollbackBatchesHalt(t *testing.T) {
	t.Parallel()

	mockEtherman := new(mocks.EthermanMock)
	mockState := new(mocks.StateInterfaceMock)

	mockEtherman.On("GetLatestVerifiedBatchNum").Return(uint64(110), nil).Once()
	mockState.On("DeleteUngeneratedProofs", mock.Anything, mock.Anything).Return(nil).Once()
	mockState.On("DeleteGeneratedProofs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	// Test data
	rollbackData := synchronizer.RollbackBatchesData{
		LastBatchNumber: 100,
	}

	a := Aggregator{
		ctx:                 context.Background(),
		etherman:            mockEtherman,
		state:               mockState,
		logger:              log.GetDefaultLogger(),
		halted:              atomic.Bool{},
		accInputHashes:      make(map[uint64]common.Hash),
		accInputHashesMutex: &sync.Mutex{},
	}

	a.halted.Store(false)
	go a.handleRollbackBatches(rollbackData)
	time.Sleep(3 * time.Second)

	assert.True(t, a.halted.Load())
	mockEtherman.AssertExpectations(t)
}

func Test_handleRollbackBatchesError(t *testing.T) {
	t.Parallel()

	mockEtherman := new(mocks.EthermanMock)
	mockState := new(mocks.StateInterfaceMock)

	mockEtherman.On("GetLatestVerifiedBatchNum").Return(uint64(110), fmt.Errorf("error")).Once()

	// Test data
	rollbackData := synchronizer.RollbackBatchesData{
		LastBatchNumber: 100,
	}

	a := Aggregator{
		ctx:                 context.Background(),
		etherman:            mockEtherman,
		state:               mockState,
		logger:              log.GetDefaultLogger(),
		halted:              atomic.Bool{},
		accInputHashes:      make(map[uint64]common.Hash),
		accInputHashesMutex: &sync.Mutex{},
	}

	a.halted.Store(false)
	go a.handleRollbackBatches(rollbackData)
	time.Sleep(3 * time.Second)

	assert.True(t, a.halted.Load())
	mockEtherman.AssertExpectations(t)
}

func Test_sendFinalProofSuccess(t *testing.T) {
	require := require.New(t)
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

				batch := rpctypes.NewRPCBatch(batchNumFinal, common.Hash{}, []string{}, []byte{}, common.Hash{}, common.Hash{}, common.Hash{}, common.Address{}, false)
				m.rpcMock.On("GetBatch", batchNumFinal).Return(batch, nil)

				m.etherman.On("GetRollupId").Return(uint32(1)).Once()
				testHash := common.BytesToHash([]byte("test hash"))
				m.aggLayerClientMock.On("SendTx", mock.Anything).Return(testHash, nil)
				m.aggLayerClientMock.On("WaitTxToBeMined", testHash, mock.Anything).Return(nil)
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

				batch := rpctypes.NewRPCBatch(batchNumFinal, common.Hash{}, []string{}, []byte{}, common.Hash{}, common.Hash{}, common.Hash{}, common.Address{}, false)
				m.rpcMock.On("GetBatch", batchNumFinal).Return(batch, nil)

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
			stateMock := mocks.NewStateInterfaceMock(t)
			ethTxManager := mocks.NewEthTxManagerClientMock(t)
			etherman := mocks.NewEthermanMock(t)
			aggLayerClient := agglayer.NewAgglayerClientMock(t)
			rpcMock := mocks.NewRPCInterfaceMock(t)

			curve := elliptic.P256()
			privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
			require.NoError(err, "error generating key")

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
				rpcClient:               rpcMock,
				accInputHashes:          make(map[uint64]common.Hash),
				accInputHashesMutex:     &sync.Mutex{},
			}
			a.ctx, a.exit = context.WithCancel(context.Background())

			m := mox{
				stateMock:          stateMock,
				ethTxManager:       ethTxManager,
				etherman:           etherman,
				aggLayerClientMock: aggLayerClient,
				rpcMock:            rpcMock,
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
	require := require.New(t)
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
				m.rpcMock.On("GetBatch", batchNumFinal).Run(func(args mock.Arguments) {
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

				batch := rpctypes.NewRPCBatch(batchNumFinal, common.Hash{}, []string{}, []byte{}, common.Hash{}, common.Hash{}, common.Hash{}, common.Address{}, false)
				m.rpcMock.On("GetBatch", batchNumFinal).Return(batch, nil)

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

				batch := rpctypes.NewRPCBatch(batchNumFinal, common.Hash{}, []string{}, []byte{}, common.Hash{}, common.Hash{}, common.Hash{}, common.Address{}, false)
				m.rpcMock.On("GetBatch", batchNumFinal).Return(batch, nil)

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

				batch := rpctypes.NewRPCBatch(batchNumFinal, common.Hash{}, []string{}, []byte{}, common.Hash{}, common.Hash{}, common.Hash{}, common.Address{}, false)
				m.rpcMock.On("GetBatch", batchNumFinal).Return(batch, nil)

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

				batch := rpctypes.NewRPCBatch(batchNumFinal, common.Hash{}, []string{}, []byte{}, common.Hash{}, common.Hash{}, common.Hash{}, common.Address{}, false)
				m.rpcMock.On("GetBatch", batchNumFinal).Return(batch, nil)

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
			stateMock := mocks.NewStateInterfaceMock(t)
			ethTxManager := mocks.NewEthTxManagerClientMock(t)
			etherman := mocks.NewEthermanMock(t)
			aggLayerClient := agglayer.NewAgglayerClientMock(t)
			rpcMock := mocks.NewRPCInterfaceMock(t)

			curve := elliptic.P256()
			privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
			require.NoError(err, "error generating key")

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
				rpcClient:               rpcMock,
				accInputHashes:          make(map[uint64]common.Hash),
				accInputHashesMutex:     &sync.Mutex{},
			}
			a.ctx, a.exit = context.WithCancel(context.Background())

			m := mox{
				stateMock:          stateMock,
				ethTxManager:       ethTxManager,
				etherman:           etherman,
				aggLayerClientMock: aggLayerClient,
				rpcMock:            rpcMock,
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
	finalProofID := "finalProofID"

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

				m.proverMock.On("Name").Return("name").Once()
				m.proverMock.On("ID").Return("id").Once()
				m.proverMock.On("Addr").Return("addr").Once()
				m.proverMock.On("FinalProof", recursiveProof.Proof, a.cfg.SenderAddress).Return(&finalProofID, nil).Once()
				m.proverMock.On("WaitFinalProof", mock.Anything, finalProofID).Return(&finalProof, nil).Once()
				finalBatch := rpctypes.NewRPCBatch(batchNumFinal, common.Hash{}, []string{}, []byte{}, common.Hash{}, common.BytesToHash([]byte("mock LocalExitRoot")), common.BytesToHash([]byte("mock StateRoot")), common.Address{}, false)
				m.rpcMock.On("GetBatch", batchNumFinal).Return(finalBatch, nil).Once()
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
			proverMock := mocks.NewProverInterfaceMock(t)
			stateMock := mocks.NewStateInterfaceMock(t)
			rpcMock := mocks.NewRPCInterfaceMock(t)
			m := mox{
				proverMock: proverMock,
				stateMock:  stateMock,
				rpcMock:    rpcMock,
			}
			a := Aggregator{
				state:  stateMock,
				logger: log.GetDefaultLogger(),
				cfg: Config{
					SenderAddress: common.BytesToAddress([]byte("from")).Hex(),
				},
				rpcClient:           rpcMock,
				accInputHashes:      make(map[uint64]common.Hash),
				accInputHashesMutex: &sync.Mutex{},
			}

			tc.setup(m, &a)
			fProof, err := a.buildFinalProof(context.Background(), proverMock, recursiveProof)
			tc.asserts(err, fProof)
		})
	}
}

func Test_tryBuildFinalProof(t *testing.T) {
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
	finalProofID := "finalProofID"
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

	proverCtx := context.WithValue(context.Background(), "owner", ownerProver) //nolint:staticcheck
	matchProverCtxFn := func(ctx context.Context) bool { return ctx.Value("owner") == ownerProver }
	matchAggregatorCtxFn := func(ctx context.Context) bool { return ctx.Value("owner") == ownerAggregator }
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
			name:  "valid proof",
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
				accInputHashes:          make(map[uint64]common.Hash),
				accInputHashesMutex:     &sync.Mutex{},
			}

			aggregatorCtx := context.WithValue(context.Background(), "owner", ownerAggregator) //nolint:staticcheck
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

			var wg sync.WaitGroup
			if tc.assertFinalMsg != nil {
				// wait for the final proof over the channel
				wg := sync.WaitGroup{}
				wg.Add(1)
				go func() {
					defer wg.Done()
					msg := <-a.finalProof
					tc.assertFinalMsg(&msg)
				}()
			}

			result, err := a.tryBuildFinalProof(proverCtx, proverMock, tc.proof)

			if tc.asserts != nil {
				tc.asserts(result, &a, err)
			}

			if tc.assertFinalMsg != nil {
				WaitUntil(t, &wg, time.Second)
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
	proverCtx := context.WithValue(context.Background(), "owner", ownerProver) //nolint:staticcheck
	matchProverCtxFn := func(ctx context.Context) bool { return ctx.Value("owner") == ownerProver }
	matchAggregatorCtxFn := func(ctx context.Context) bool { return ctx.Value("owner") == ownerAggregator }
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
				m.proverMock.On("WaitRecursiveProof", mock.MatchedBy(matchProverCtxFn), proofID).Return("", common.Hash{}, common.Hash{}, errTest).Once()
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
				m.proverMock.On("WaitRecursiveProof", mock.MatchedBy(matchProverCtxFn), proofID).Return("", common.Hash{}, common.Hash{}, errTest).Once()
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
				m.proverMock.On("WaitRecursiveProof", mock.MatchedBy(matchProverCtxFn), proofID).Return(recursiveProof, common.Hash{}, common.Hash{}, nil).Once()
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
				m.proverMock.On("WaitRecursiveProof", mock.MatchedBy(matchProverCtxFn), proofID).Return(recursiveProof, common.Hash{}, common.Hash{}, nil).Once()
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
				m.proverMock.On("WaitRecursiveProof", mock.MatchedBy(matchProverCtxFn), proofID).Return(recursiveProof, common.Hash{}, common.Hash{}, nil).Once()
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
				accInputHashes:          make(map[uint64]common.Hash),
				accInputHashesMutex:     &sync.Mutex{},
			}
			aggregatorCtx := context.WithValue(context.Background(), "owner", ownerAggregator) //nolint:staticcheck
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

func Test_accInputHashFunctions(t *testing.T) {
	aggregator := Aggregator{
		accInputHashes:      make(map[uint64]common.Hash),
		accInputHashesMutex: &sync.Mutex{},
	}

	hash1 := common.BytesToHash([]byte("hash1"))
	hash2 := common.BytesToHash([]byte("hash2"))

	aggregator.setAccInputHash(1, hash1)
	aggregator.setAccInputHash(2, hash2)

	assert.Equal(t, 2, len(aggregator.accInputHashes))

	hash3 := aggregator.getAccInputHash(1)
	assert.Equal(t, hash1, hash3)

	aggregator.removeAccInputHashes(1, 2)
	assert.Equal(t, 0, len(aggregator.accInputHashes))
}
