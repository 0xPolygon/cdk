package aggregator

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	cdkTypes "github.com/0xPolygon/cdk-rpc/types"
	"github.com/0xPolygon/cdk/agglayer"
	ethmanTypes "github.com/0xPolygon/cdk/aggregator/ethmantypes"
	"github.com/0xPolygon/cdk/aggregator/prover"
	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/l1infotree"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/rpc"
	"github.com/0xPolygon/cdk/state"
	"github.com/0xPolygon/zkevm-ethtx-manager/ethtxmanager"
	ethtxlog "github.com/0xPolygon/zkevm-ethtx-manager/log"
	ethtxtypes "github.com/0xPolygon/zkevm-ethtx-manager/types"
	synclog "github.com/0xPolygonHermez/zkevm-synchronizer-l1/log"
	"github.com/0xPolygonHermez/zkevm-synchronizer-l1/state/entities"
	"github.com/0xPolygonHermez/zkevm-synchronizer-l1/synchronizer"
	"github.com/0xPolygonHermez/zkevm-synchronizer-l1/synchronizer/l1_check_block"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/peer"
)

const (
	mockedStateRoot     = "0x090bcaf734c4f06c93954a827b45a6e8c67b8e0fd1e0a35a1c5982d6961828f9"
	mockedLocalExitRoot = "0x17c04c3760510b48c6012742c540a81aba4bca2f78b9d14bfd2f123e2e53ea3e"
	maxDBBigIntValue    = 9223372036854775807
)

type finalProofMsg struct {
	proverName     string
	proverID       string
	recursiveProof *state.Proof
	finalProof     *prover.FinalProof
}

// Aggregator represents an aggregator
type Aggregator struct {
	prover.UnimplementedAggregatorServiceServer

	cfg    Config
	logger *log.Logger

	state               StateInterface
	etherman            Etherman
	ethTxManager        EthTxManagerClient
	l1Syncr             synchronizer.Synchronizer
	halted              atomic.Bool
	accInputHashes      map[uint64]common.Hash
	accInputHashesMutex *sync.Mutex

	profitabilityChecker    aggregatorTxProfitabilityChecker
	timeSendFinalProof      time.Time
	timeCleanupLockedProofs types.Duration
	stateDBMutex            *sync.Mutex
	timeSendFinalProofMutex *sync.RWMutex

	finalProof     chan finalProofMsg
	verifyingProof bool

	witnessRetrievalChan chan state.DBBatch

	srv  *grpc.Server
	ctx  context.Context
	exit context.CancelFunc

	sequencerPrivateKey *ecdsa.PrivateKey
	aggLayerClient      agglayer.AgglayerClientInterface

	rpcClient RPCInterface
}

// New creates a new aggregator.
func New(
	ctx context.Context,
	cfg Config,
	logger *log.Logger,
	stateInterface StateInterface,
	etherman Etherman) (*Aggregator, error) {
	var profitabilityChecker aggregatorTxProfitabilityChecker

	switch cfg.TxProfitabilityCheckerType {
	case ProfitabilityBase:
		profitabilityChecker = NewTxProfitabilityCheckerBase(
			stateInterface, cfg.IntervalAfterWhichBatchConsolidateAnyway.Duration, cfg.TxProfitabilityMinReward.Int,
		)
	case ProfitabilityAcceptAll:
		profitabilityChecker = NewTxProfitabilityCheckerAcceptAll(
			stateInterface, cfg.IntervalAfterWhichBatchConsolidateAnyway.Duration,
		)
	}

	// Create ethtxmanager client
	cfg.EthTxManager.Log = ethtxlog.Config{
		Environment: ethtxlog.LogEnvironment(cfg.Log.Environment),
		Level:       cfg.Log.Level,
		Outputs:     cfg.Log.Outputs,
	}
	ethTxManager, err := ethtxmanager.New(cfg.EthTxManager)
	if err != nil {
		logger.Fatalf("error creating ethtxmanager client: %v", err)
	}

	// Synchonizer logs
	syncLogConfig := synclog.Config{
		Environment: synclog.LogEnvironment(cfg.Log.Environment),
		Level:       cfg.Log.Level,
		Outputs:     cfg.Log.Outputs,
	}

	cfg.Synchronizer.Log = syncLogConfig

	// Create L1 synchronizer client
	cfg.Synchronizer.Etherman.L1URL = cfg.EthTxManager.Etherman.URL
	logger.Debugf("Creating synchronizer client with config: %+v", cfg.Synchronizer)
	l1Syncr, err := synchronizer.NewSynchronizer(ctx, cfg.Synchronizer)
	if err != nil {
		logger.Fatalf("failed to create synchronizer client, error: %v", err)
	}

	var (
		aggLayerClient      agglayer.AgglayerClientInterface
		sequencerPrivateKey *ecdsa.PrivateKey
	)

	if !cfg.SyncModeOnlyEnabled && cfg.SettlementBackend == AggLayer {
		aggLayerClient = agglayer.NewAggLayerClient(cfg.AggLayerURL)

		sequencerPrivateKey, err = cdkcommon.NewKeyFromKeystore(cfg.SequencerPrivateKey)
		if err != nil {
			return nil, err
		}
	}

	a := &Aggregator{
		ctx:                     ctx,
		cfg:                     cfg,
		logger:                  logger,
		state:                   stateInterface,
		etherman:                etherman,
		ethTxManager:            ethTxManager,
		l1Syncr:                 l1Syncr,
		accInputHashes:          make(map[uint64]common.Hash),
		accInputHashesMutex:     &sync.Mutex{},
		profitabilityChecker:    profitabilityChecker,
		stateDBMutex:            &sync.Mutex{},
		timeSendFinalProofMutex: &sync.RWMutex{},
		timeCleanupLockedProofs: cfg.CleanupLockedProofsInterval,
		finalProof:              make(chan finalProofMsg),
		aggLayerClient:          aggLayerClient,
		sequencerPrivateKey:     sequencerPrivateKey,
		witnessRetrievalChan:    make(chan state.DBBatch),
		rpcClient:               rpc.NewBatchEndpoints(cfg.RPCURL),
	}

	if a.ctx == nil {
		a.ctx, a.exit = context.WithCancel(a.ctx)
	}

	// Set function to handle events on L1
	if !cfg.SyncModeOnlyEnabled {
		a.l1Syncr.SetCallbackOnReorgDone(a.handleReorg)
		a.l1Syncr.SetCallbackOnRollbackBatches(a.handleRollbackBatches)
	}

	return a, nil
}

func (a *Aggregator) getAccInputHash(batchNumber uint64) common.Hash {
	a.accInputHashesMutex.Lock()
	defer a.accInputHashesMutex.Unlock()
	return a.accInputHashes[batchNumber]
}

func (a *Aggregator) setAccInputHash(batchNumber uint64, accInputHash common.Hash) {
	a.accInputHashesMutex.Lock()
	defer a.accInputHashesMutex.Unlock()
	a.accInputHashes[batchNumber] = accInputHash
}

func (a *Aggregator) removeAccInputHashes(firstBatch, lastBatch uint64) {
	a.accInputHashesMutex.Lock()
	defer a.accInputHashesMutex.Unlock()
	for i := firstBatch; i <= lastBatch; i++ {
		delete(a.accInputHashes, i)
	}
}

func (a *Aggregator) handleReorg(reorgData synchronizer.ReorgExecutionResult) {
	a.logger.Warnf("Reorg detected, reorgData: %+v", reorgData)

	// Get new latest verified batch number
	lastVBatchNumber, err := a.l1Syncr.GetLastestVirtualBatchNumber(a.ctx)
	if err != nil {
		a.logger.Errorf("Error getting last virtual batch number: %v", err)
	} else {
		// Delete wip proofs
		err = a.state.DeleteUngeneratedProofs(a.ctx, nil)
		if err != nil {
			a.logger.Errorf("Error deleting ungenerated proofs: %v", err)
		} else {
			a.logger.Info("Deleted ungenerated proofs")
		}

		// Delete any proof for the batches that have been rolled back
		err = a.state.DeleteGeneratedProofs(a.ctx, lastVBatchNumber+1, maxDBBigIntValue, nil)
		if err != nil {
			a.logger.Errorf("Error deleting generated proofs: %v", err)
		} else {
			a.logger.Infof("Deleted generated proofs for batches newer than %d", lastVBatchNumber)
		}
	}

	// Halt the aggregator
	a.halted.Store(true)
	for {
		a.logger.Warnf(
			"Halting the aggregator due to a L1 reorg. " +
				"Reorged data has been deleted, so it is safe to manually restart the aggregator.",
		)
		time.Sleep(10 * time.Second) //nolint:mnd
	}
}

func (a *Aggregator) handleRollbackBatches(rollbackData synchronizer.RollbackBatchesData) {
	a.logger.Warnf("Rollback batches event, rollbackBatchesData: %+v", rollbackData)

	var err error
	var accInputHash *common.Hash

	// Get new last verified batch number from L1
	lastVerifiedBatchNumber, err := a.etherman.GetLatestVerifiedBatchNum()
	if err != nil {
		a.logger.Errorf("Error getting latest verified batch number: %v", err)
	}

	a.logger.Infof("Last Verified Batch Number:%v", lastVerifiedBatchNumber)

	// Check lastVerifiedBatchNumber makes sense
	if err == nil && lastVerifiedBatchNumber > rollbackData.LastBatchNumber {
		err = fmt.Errorf(
			"last verified batch number %d is greater than the last batch number %d in the rollback data",
			lastVerifiedBatchNumber, rollbackData.LastBatchNumber,
		)
	}

	if err == nil {
		accInputHash, err = a.getVerifiedBatchAccInputHash(a.ctx, lastVerifiedBatchNumber)
		if err == nil {
			a.accInputHashesMutex.Lock()
			a.accInputHashes = make(map[uint64]common.Hash)
			a.accInputHashesMutex.Unlock()
			a.logger.Infof("Starting AccInputHash:%v", accInputHash.String())
			a.setAccInputHash(lastVerifiedBatchNumber, *accInputHash)
		}
	}

	// Delete wip proofs
	if err == nil {
		err = a.state.DeleteUngeneratedProofs(a.ctx, nil)
		if err != nil {
			a.logger.Errorf("Error deleting ungenerated proofs: %v", err)
		} else {
			a.logger.Info("Deleted ungenerated proofs")
		}
	}

	// Delete any proof for the batches that have been rolled back
	if err == nil {
		err = a.state.DeleteGeneratedProofs(a.ctx, rollbackData.LastBatchNumber+1, maxDBBigIntValue, nil)
		if err != nil {
			a.logger.Errorf("Error deleting generated proofs: %v", err)
		} else {
			a.logger.Infof("Deleted generated proofs for batches newer than %d", rollbackData.LastBatchNumber)
		}
	}

	if err == nil {
		a.logger.Info("Handling rollback batches event finished successfully")
	} else {
		// Halt the aggregator
		a.halted.Store(true)
		for {
			a.logger.Errorf("Halting the aggregator due to an error handling rollback batches event: %v", err)
			time.Sleep(10 * time.Second) //nolint:mnd
		}
	}
}

// Start starts the aggregator
func (a *Aggregator) Start() error {
	// Initial L1 Sync blocking
	err := a.l1Syncr.Sync(true)
	if err != nil {
		a.logger.Fatalf("Failed to synchronize from L1: %v", err)
		return err
	}

	// Keep syncing L1
	go func() {
		err := a.l1Syncr.Sync(false)
		if err != nil {
			a.logger.Fatalf("Failed to synchronize from L1: %v", err)
		}
	}()

	if !a.cfg.SyncModeOnlyEnabled {
		address := fmt.Sprintf("%s:%d", a.cfg.Host, a.cfg.Port)
		lis, err := net.Listen("tcp", address)
		if err != nil {
			a.logger.Fatalf("Failed to listen: %v", err)
		}

		a.srv = grpc.NewServer()
		prover.RegisterAggregatorServiceServer(a.srv, a)

		healthService := newHealthChecker()
		grpchealth.RegisterHealthServer(a.srv, healthService)

		// Get last verified batch number to set the starting point for verifications
		lastVerifiedBatchNumber, err := a.etherman.GetLatestVerifiedBatchNum()
		if err != nil {
			return err
		}

		a.logger.Infof("Last Verified Batch Number:%v", lastVerifiedBatchNumber)

		accInputHash, err := a.getVerifiedBatchAccInputHash(a.ctx, lastVerifiedBatchNumber)
		if err != nil {
			return err
		}

		a.logger.Infof("Starting AccInputHash:%v", accInputHash.String())
		a.setAccInputHash(lastVerifiedBatchNumber, *accInputHash)

		// Delete existing proofs
		err = a.state.DeleteGeneratedProofs(a.ctx, lastVerifiedBatchNumber, maxDBBigIntValue, nil)
		if err != nil {
			return fmt.Errorf("failed to delete proofs table %w", err)
		}

		a.resetVerifyProofTime()

		go a.cleanupLockedProofs()
		go a.sendFinalProof()
		go a.ethTxManager.Start()

		// A this point everything is ready, so start serving
		go func() {
			a.logger.Infof("Server listening on port %d", a.cfg.Port)
			if err := a.srv.Serve(lis); err != nil {
				a.exit()
				a.logger.Fatalf("Failed to serve: %v", err)
			}
		}()
	}

	<-a.ctx.Done()

	return a.ctx.Err()
}

// Stop stops the Aggregator server.
func (a *Aggregator) Stop() {
	a.exit()
	a.srv.Stop()
}

// Channel implements the bi-directional communication channel between the
// Prover client and the Aggregator server.
func (a *Aggregator) Channel(stream prover.AggregatorService_ChannelServer) error {
	ctx := stream.Context()
	var proverAddr net.Addr
	p, ok := peer.FromContext(ctx)
	if ok {
		proverAddr = p.Addr
	}
	proverLogger := log.WithFields("module", cdkcommon.PROVER)
	prover, err := prover.New(proverLogger, stream, proverAddr, a.cfg.ProofStatePollingInterval)
	if err != nil {
		return err
	}

	tmpLogger := proverLogger.WithFields(
		"prover", prover.Name(),
		"proverId", prover.ID(),
		"proverAddr", prover.Addr(),
	)
	tmpLogger.Info("Establishing stream connection with prover")

	// Check if prover supports the required Fork ID
	if !prover.SupportsForkID(a.cfg.ForkId) {
		err := errors.New("prover does not support required fork ID")
		tmpLogger.Warn(FirstToUpper(err.Error()))

		return err
	}

	for {
		select {
		case <-a.ctx.Done():
			// server disconnected
			return a.ctx.Err()
		case <-ctx.Done():
			// client disconnected
			return ctx.Err()

		default:
			if !a.halted.Load() {
				isIdle, err := prover.IsIdle()
				if err != nil {
					tmpLogger.Errorf("Failed to check if prover is idle: %v", err)
					time.Sleep(a.cfg.RetryTime.Duration)

					continue
				}
				if !isIdle {
					tmpLogger.Debug("Prover is not idle")
					time.Sleep(a.cfg.RetryTime.Duration)

					continue
				}

				_, err = a.tryBuildFinalProof(ctx, prover, nil)
				if err != nil {
					tmpLogger.Errorf("Error checking proofs to verify: %v", err)
				}

				proofGenerated, err := a.tryAggregateProofs(ctx, prover)
				if err != nil {
					tmpLogger.Errorf("Error trying to aggregate proofs: %v", err)
				}

				if !proofGenerated {
					proofGenerated, err = a.tryGenerateBatchProof(ctx, prover)
					if err != nil {
						tmpLogger.Errorf("Error trying to generate proof: %v", err)
					}
				}
				if !proofGenerated {
					// if no proof was generated (aggregated or batch) wait some time before retry
					time.Sleep(a.cfg.RetryTime.Duration)
				} // if proof was generated we retry immediately as probably we have more proofs to process
			}
		}
	}
}

// This function waits to receive a final proof from a prover. Once it receives
// the proof, it performs these steps in order:
// - send the final proof to L1
// - wait for the synchronizer to catch up
// - clean up the cache of recursive proofs
func (a *Aggregator) sendFinalProof() {
	for {
		select {
		case <-a.ctx.Done():
			return
		case msg := <-a.finalProof:
			ctx := a.ctx
			proof := msg.recursiveProof

			tmpLogger := a.logger.WithFields(
				"proofId", proof.ProofID,
				"batches", fmt.Sprintf("%d-%d", proof.BatchNumber, proof.BatchNumberFinal))
			tmpLogger.Info("Verifying final proof with ethereum smart contract")

			a.startProofVerification()

			// Get Batch from RPC
			rpcFinalBatch, err := a.rpcClient.GetBatch(proof.BatchNumberFinal)
			if err != nil {
				a.logger.Errorf("error getting batch %d from RPC: %v.", proof.BatchNumberFinal, err)
				a.endProofVerification()
				continue
			}

			inputs := ethmanTypes.FinalProofInputs{
				FinalProof:       msg.finalProof,
				NewLocalExitRoot: rpcFinalBatch.LocalExitRoot().Bytes(),
				NewStateRoot:     rpcFinalBatch.StateRoot().Bytes(),
			}

			switch a.cfg.SettlementBackend {
			case AggLayer:
				if success := a.settleWithAggLayer(ctx, proof, inputs); !success {
					continue
				}
			default:
				if success := a.settleDirect(ctx, proof, inputs); !success {
					continue
				}
			}

			a.resetVerifyProofTime()
			a.endProofVerification()
		}
	}
}

func (a *Aggregator) settleWithAggLayer(
	ctx context.Context,
	proof *state.Proof,
	inputs ethmanTypes.FinalProofInputs) bool {
	proofStrNo0x := strings.TrimPrefix(inputs.FinalProof.Proof, "0x")
	proofBytes := common.Hex2Bytes(proofStrNo0x)
	tx := agglayer.Tx{
		LastVerifiedBatch: cdkTypes.ArgUint64(proof.BatchNumber - 1),
		NewVerifiedBatch:  cdkTypes.ArgUint64(proof.BatchNumberFinal),
		ZKP: agglayer.ZKP{
			NewStateRoot:     common.BytesToHash(inputs.NewStateRoot),
			NewLocalExitRoot: common.BytesToHash(inputs.NewLocalExitRoot),
			Proof:            cdkTypes.ArgBytes(proofBytes),
		},
		RollupID: a.etherman.GetRollupId(),
	}
	signedTx, err := tx.Sign(a.sequencerPrivateKey)
	if err != nil {
		a.logger.Errorf("failed to sign tx: %v", err)
		a.handleFailureToAddVerifyBatchToBeMonitored(ctx, proof)

		return false
	}

	a.logger.Debug("final proof: %+v", tx)
	a.logger.Debug("final proof signedTx: ", signedTx.Tx.ZKP.Proof.Hex())
	txHash, err := a.aggLayerClient.SendTx(*signedTx)
	if err != nil {
		if errors.Is(err, agglayer.ErrAgglayerRateLimitExceeded) {
			a.logger.Errorf("%s. Config param VerifyProofInterval should match the agglayer configured rate limit.", err)
		} else {
			a.logger.Errorf("failed to send tx to the agglayer: %v", err)
		}
		a.handleFailureToAddVerifyBatchToBeMonitored(ctx, proof)
		return false
	}

	a.logger.Infof("tx %s sent to agglayer, waiting to be mined", txHash.Hex())
	a.logger.Debugf("Timeout set to %f seconds", a.cfg.AggLayerTxTimeout.Duration.Seconds())
	waitCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(a.cfg.AggLayerTxTimeout.Duration))
	defer cancelFunc()
	if err := a.aggLayerClient.WaitTxToBeMined(txHash, waitCtx); err != nil {
		a.logger.Errorf("agglayer didn't mine the tx: %v", err)
		a.handleFailureToAddVerifyBatchToBeMonitored(ctx, proof)

		return false
	}

	return true
}

// settleDirect sends the final proof to the L1 smart contract directly.
func (a *Aggregator) settleDirect(
	ctx context.Context,
	proof *state.Proof,
	inputs ethmanTypes.FinalProofInputs) bool {
	// add batch verification to be monitored
	sender := common.HexToAddress(a.cfg.SenderAddress)
	to, data, err := a.etherman.BuildTrustedVerifyBatchesTxData(
		proof.BatchNumber-1, proof.BatchNumberFinal, &inputs, sender,
	)
	if err != nil {
		a.logger.Errorf("Error estimating batch verification to add to eth tx manager: %v", err)
		a.handleFailureToAddVerifyBatchToBeMonitored(ctx, proof)

		return false
	}

	monitoredTxID, err := a.ethTxManager.Add(ctx, to, big.NewInt(0), data, a.cfg.GasOffset, nil)
	if err != nil {
		a.logger.Errorf("Error Adding TX to ethTxManager: %v", err)
		mTxLogger := ethtxmanager.CreateLogger(monitoredTxID, sender, to)
		mTxLogger.Errorf("Error to add batch verification tx to eth tx manager: %v", err)
		a.handleFailureToAddVerifyBatchToBeMonitored(ctx, proof)

		return false
	}

	// process monitored batch verifications before starting a next cycle
	a.ethTxManager.ProcessPendingMonitoredTxs(ctx, func(result ethtxtypes.MonitoredTxResult) {
		a.handleMonitoredTxResult(result, proof.BatchNumber, proof.BatchNumberFinal)
	})

	return true
}

func (a *Aggregator) handleFailureToAddVerifyBatchToBeMonitored(ctx context.Context, proof *state.Proof) {
	tmpLogger := a.logger.WithFields(
		"proofId", proof.ProofID,
		"batches", fmt.Sprintf("%d-%d", proof.BatchNumber, proof.BatchNumberFinal),
	)
	proof.GeneratingSince = nil
	err := a.state.UpdateGeneratedProof(ctx, proof, nil)
	if err != nil {
		tmpLogger.Errorf("Failed updating proof state (false): %v", err)
	}
	a.endProofVerification()
}

// buildFinalProof builds and return the final proof for an aggregated/batch proof.
func (a *Aggregator) buildFinalProof(
	ctx context.Context, prover ProverInterface, proof *state.Proof) (*prover.FinalProof, error) {
	tmpLogger := a.logger.WithFields(
		"prover", prover.Name(),
		"proverId", prover.ID(),
		"proverAddr", prover.Addr(),
		"recursiveProofId", *proof.ProofID,
		"batches", fmt.Sprintf("%d-%d", proof.BatchNumber, proof.BatchNumberFinal),
	)

	finalProofID, err := prover.FinalProof(proof.Proof, a.cfg.SenderAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get final proof id: %w", err)
	}
	proof.ProofID = finalProofID

	tmpLogger.Infof("Final proof ID for batches [%d-%d]: %s", proof.BatchNumber, proof.BatchNumberFinal, *proof.ProofID)
	tmpLogger = tmpLogger.WithFields("finalProofId", finalProofID)

	finalProof, err := prover.WaitFinalProof(ctx, *proof.ProofID)
	if err != nil {
		return nil, fmt.Errorf("failed to get final proof from prover: %w", err)
	}

	// mock prover sanity check
	if string(finalProof.Public.NewStateRoot) == mockedStateRoot &&
		string(finalProof.Public.NewLocalExitRoot) == mockedLocalExitRoot {
		// This local exit root and state root come from the mock
		// prover, use the one captured by the executor instead
		rpcFinalBatch, err := a.rpcClient.GetBatch(proof.BatchNumberFinal)
		if err != nil {
			return nil, fmt.Errorf("error getting batch %d from RPC: %w", proof.BatchNumberFinal, err)
		}

		tmpLogger.Warnf(
			"NewLocalExitRoot and NewStateRoot look like a mock values, using values from executor instead: LER: %v, SR: %v",
			rpcFinalBatch.LocalExitRoot().TerminalString(), rpcFinalBatch.StateRoot().TerminalString())
		finalProof.Public.NewStateRoot = rpcFinalBatch.StateRoot().Bytes()
		finalProof.Public.NewLocalExitRoot = rpcFinalBatch.LocalExitRoot().Bytes()
	}

	return finalProof, nil
}

// tryBuildFinalProof checks if the provided proof is eligible to be used to
// build the final proof.  If no proof is provided it looks for a previously
// generated proof.  If the proof is eligible, then the final proof generation
// is triggered.
func (a *Aggregator) tryBuildFinalProof(ctx context.Context, prover ProverInterface, proof *state.Proof) (bool, error) {
	proverName := prover.Name()
	proverID := prover.ID()

	tmpLogger := a.logger.WithFields(
		"prover", proverName,
		"proverId", proverID,
		"proverAddr", prover.Addr(),
	)
	tmpLogger.Debug("tryBuildFinalProof start")

	if !a.canVerifyProof() {
		tmpLogger.Debug("Time to verify proof not reached or proof verification in progress")
		return false, nil
	}
	tmpLogger.Debug("Send final proof time reached")

	lastVerifiedBatchNumber, err := a.etherman.GetLatestVerifiedBatchNum()
	if err != nil {
		return false, err
	}

	if proof == nil {
		// we don't have a proof generating at the moment, check if we
		// have a proof ready to verify
		proof, err = a.getAndLockProofReadyToVerify(ctx, lastVerifiedBatchNumber)
		if errors.Is(err, state.ErrNotFound) {
			// nothing to verify, swallow the error
			tmpLogger.Debug("No proof ready to verify")
			return false, nil
		}
		if err != nil {
			return false, err
		}

		defer func() {
			if err != nil {
				// Set the generating state to false for the proof ("unlock" it)
				proof.GeneratingSince = nil
				err2 := a.state.UpdateGeneratedProof(a.ctx, proof, nil)
				if err2 != nil {
					tmpLogger.Errorf("Failed to unlock proof: %v", err2)
				}
			}
		}()
	} else {
		// we do have a proof generating at the moment, check if it is
		// eligible to be verified
		eligible, err := a.validateEligibleFinalProof(ctx, proof, lastVerifiedBatchNumber)
		if err != nil {
			return false, fmt.Errorf("failed to validate eligible final proof, %w", err)
		}
		if !eligible {
			return false, nil
		}
	}

	tmpLogger = tmpLogger.WithFields(
		"proofId", *proof.ProofID,
		"batches", fmt.Sprintf("%d-%d", proof.BatchNumber, proof.BatchNumberFinal),
	)

	// at this point we have an eligible proof, build the final one using it
	finalProof, err := a.buildFinalProof(ctx, prover, proof)
	if err != nil {
		err = fmt.Errorf("failed to build final proof, %w", err)
		tmpLogger.Error(FirstToUpper(err.Error()))
		return false, err
	}

	msg := finalProofMsg{
		proverName:     proverName,
		proverID:       proverID,
		recursiveProof: proof,
		finalProof:     finalProof,
	}

	select {
	case <-a.ctx.Done():
		return false, a.ctx.Err()
	case a.finalProof <- msg:
	}

	tmpLogger.Debug("tryBuildFinalProof end")
	return true, nil
}

func (a *Aggregator) validateEligibleFinalProof(
	ctx context.Context, proof *state.Proof, lastVerifiedBatchNum uint64,
) (bool, error) {
	batchNumberToVerify := lastVerifiedBatchNum + 1

	if proof.BatchNumber != batchNumberToVerify {
		if proof.BatchNumber < batchNumberToVerify &&
			proof.BatchNumberFinal >= batchNumberToVerify {
			// We have a proof that contains some batches below the last batch verified, anyway can be eligible as final proof
			a.logger.Warnf("Proof %d-%d contains some batches lower than last batch verified %d. Check anyway if it is eligible",
				proof.BatchNumber, proof.BatchNumberFinal, lastVerifiedBatchNum)
		} else if proof.BatchNumberFinal < batchNumberToVerify {
			// We have a proof that contains batches below that the last batch verified, we need to delete this proof
			a.logger.Warnf("Proof %d-%d lower than next batch to verify %d. Deleting it",
				proof.BatchNumber, proof.BatchNumberFinal, batchNumberToVerify)
			err := a.state.DeleteGeneratedProofs(ctx, proof.BatchNumber, proof.BatchNumberFinal, nil)
			if err != nil {
				return false, fmt.Errorf("failed to delete discarded proof, err: %w", err)
			}

			return false, nil
		} else {
			a.logger.Debugf("Proof batch number %d is not the following to last verfied batch number %d",
				proof.BatchNumber, lastVerifiedBatchNum)
			return false, nil
		}
	}

	bComplete, err := a.state.CheckProofContainsCompleteSequences(ctx, proof, nil)
	if err != nil {
		return false, fmt.Errorf("failed to check if proof contains complete sequences, %w", err)
	}
	if !bComplete {
		a.logger.Infof("Recursive proof %d-%d not eligible to be verified: not containing complete sequences",
			proof.BatchNumber, proof.BatchNumberFinal)
		return false, nil
	}

	return true, nil
}

func (a *Aggregator) getAndLockProofReadyToVerify(
	ctx context.Context, lastVerifiedBatchNum uint64,
) (*state.Proof, error) {
	a.stateDBMutex.Lock()
	defer a.stateDBMutex.Unlock()

	// Get proof ready to be verified
	proofToVerify, err := a.state.GetProofReadyToVerify(ctx, lastVerifiedBatchNum, nil)
	if err != nil {
		return nil, err
	}

	now := time.Now().Round(time.Microsecond)
	proofToVerify.GeneratingSince = &now

	err = a.state.UpdateGeneratedProof(ctx, proofToVerify, nil)
	if err != nil {
		return nil, err
	}

	return proofToVerify, nil
}

func (a *Aggregator) unlockProofsToAggregate(ctx context.Context, proof1 *state.Proof, proof2 *state.Proof) error {
	// Release proofs from generating state in a single transaction
	dbTx, err := a.state.BeginStateTransaction(ctx)
	if err != nil {
		a.logger.Warnf("Failed to begin transaction to release proof aggregation state, err: %v", err)
		return err
	}

	proof1.GeneratingSince = nil
	err = a.state.UpdateGeneratedProof(ctx, proof1, dbTx)
	if err == nil {
		proof2.GeneratingSince = nil
		err = a.state.UpdateGeneratedProof(ctx, proof2, dbTx)
	}

	if err != nil {
		if err := dbTx.Rollback(ctx); err != nil {
			err := fmt.Errorf("failed to rollback proof aggregation state: %w", err)
			a.logger.Error(FirstToUpper(err.Error()))
			return err
		}

		return fmt.Errorf("failed to release proof aggregation state: %w", err)
	}

	err = dbTx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to release proof aggregation state %w", err)
	}

	return nil
}

func (a *Aggregator) getAndLockProofsToAggregate(
	ctx context.Context, prover ProverInterface) (*state.Proof, *state.Proof, error) {
	tmpLogger := a.logger.WithFields(
		"prover", prover.Name(),
		"proverId", prover.ID(),
		"proverAddr", prover.Addr(),
	)

	a.stateDBMutex.Lock()
	defer a.stateDBMutex.Unlock()

	proof1, proof2, err := a.state.GetProofsToAggregate(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	// Set proofs in generating state in a single transaction
	dbTx, err := a.state.BeginStateTransaction(ctx)
	if err != nil {
		tmpLogger.Errorf("Failed to begin transaction to set proof aggregation state, err: %v", err)
		return nil, nil, err
	}

	now := time.Now().Round(time.Microsecond)
	proof1.GeneratingSince = &now
	err = a.state.UpdateGeneratedProof(ctx, proof1, dbTx)
	if err == nil {
		proof2.GeneratingSince = &now
		err = a.state.UpdateGeneratedProof(ctx, proof2, dbTx)
	}

	if err != nil {
		if err := dbTx.Rollback(ctx); err != nil {
			err := fmt.Errorf("failed to rollback proof aggregation state %w", err)
			tmpLogger.Error(FirstToUpper(err.Error()))
			return nil, nil, err
		}

		return nil, nil, fmt.Errorf("failed to set proof aggregation state %w", err)
	}

	err = dbTx.Commit(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to set proof aggregation state %w", err)
	}

	return proof1, proof2, nil
}

func (a *Aggregator) tryAggregateProofs(ctx context.Context, prover ProverInterface) (bool, error) {
	proverName := prover.Name()
	proverID := prover.ID()

	tmpLogger := a.logger.WithFields(
		"prover", proverName,
		"proverId", proverID,
		"proverAddr", prover.Addr(),
	)
	tmpLogger.Debug("tryAggregateProofs start")

	proof1, proof2, err0 := a.getAndLockProofsToAggregate(ctx, prover)
	if errors.Is(err0, state.ErrNotFound) {
		// nothing to aggregate, swallow the error
		tmpLogger.Debug("Nothing to aggregate")
		return false, nil
	}
	if err0 != nil {
		return false, err0
	}

	var (
		aggrProofID *string
		err         error
	)

	defer func() {
		if err != nil {
			err2 := a.unlockProofsToAggregate(a.ctx, proof1, proof2)
			if err2 != nil {
				tmpLogger.Errorf("Failed to release aggregated proofs, err: %v", err2)
			}
		}
		tmpLogger.Debug("tryAggregateProofs end")
	}()

	tmpLogger.Infof("Aggregating proofs: %d-%d and %d-%d",
		proof1.BatchNumber, proof1.BatchNumberFinal, proof2.BatchNumber, proof2.BatchNumberFinal)

	batches := fmt.Sprintf("%d-%d", proof1.BatchNumber, proof2.BatchNumberFinal)
	tmpLogger = tmpLogger.WithFields("batches", batches)

	inputProver := map[string]interface{}{
		"recursive_proof_1": proof1.Proof,
		"recursive_proof_2": proof2.Proof,
	}
	b, err := json.Marshal(inputProver)
	if err != nil {
		err = fmt.Errorf("failed to serialize input prover, %w", err)
		tmpLogger.Error(FirstToUpper(err.Error()))
		return false, err
	}

	proof := &state.Proof{
		BatchNumber:      proof1.BatchNumber,
		BatchNumberFinal: proof2.BatchNumberFinal,
		Prover:           &proverName,
		ProverID:         &proverID,
		InputProver:      string(b),
	}

	aggrProofID, err = prover.AggregatedProof(proof1.Proof, proof2.Proof)
	if err != nil {
		err = fmt.Errorf("failed to get aggregated proof id, %w", err)
		tmpLogger.Error(FirstToUpper(err.Error()))
		return false, err
	}

	proof.ProofID = aggrProofID

	tmpLogger.Infof("Proof ID for aggregated proof: %v", *proof.ProofID)
	tmpLogger = tmpLogger.WithFields("proofId", *proof.ProofID)

	recursiveProof, _, _, err := prover.WaitRecursiveProof(ctx, *proof.ProofID)
	if err != nil {
		err = fmt.Errorf("failed to get aggregated proof from prover, %w", err)
		tmpLogger.Error(FirstToUpper(err.Error()))
		return false, err
	}

	tmpLogger.Info("Aggregated proof generated")

	proof.Proof = recursiveProof

	// update the state by removing the 2 aggregated proofs and storing the
	// newly generated recursive proof
	dbTx, err := a.state.BeginStateTransaction(ctx)
	if err != nil {
		err = fmt.Errorf("failed to begin transaction to update proof aggregation state, %w", err)
		tmpLogger.Error(FirstToUpper(err.Error()))
		return false, err
	}

	err = a.state.DeleteGeneratedProofs(ctx, proof1.BatchNumber, proof2.BatchNumberFinal, dbTx)
	if err != nil {
		if err := dbTx.Rollback(ctx); err != nil {
			err := fmt.Errorf("failed to rollback proof aggregation state, %w", err)
			tmpLogger.Error(FirstToUpper(err.Error()))
			return false, err
		}
		err = fmt.Errorf("failed to delete previously aggregated proofs, %w", err)
		tmpLogger.Error(FirstToUpper(err.Error()))
		return false, err
	}

	now := time.Now().Round(time.Microsecond)
	proof.GeneratingSince = &now

	err = a.state.AddGeneratedProof(ctx, proof, dbTx)
	if err != nil {
		if err := dbTx.Rollback(ctx); err != nil {
			err := fmt.Errorf("failed to rollback proof aggregation state, %w", err)
			tmpLogger.Error(FirstToUpper(err.Error()))
			return false, err
		}
		err = fmt.Errorf("failed to store the recursive proof, %w", err)
		tmpLogger.Error(FirstToUpper(err.Error()))
		return false, err
	}

	err = dbTx.Commit(ctx)
	if err != nil {
		err = fmt.Errorf("failed to store the recursive proof, %w", err)
		tmpLogger.Error(FirstToUpper(err.Error()))
		return false, err
	}

	// NOTE(pg): the defer func is useless from now on, use a different variable
	// name for errors (or shadow err in inner scopes) to not trigger it.

	// state is up to date, check if we can send the final proof using the
	// one just crafted.
	finalProofBuilt, finalProofErr := a.tryBuildFinalProof(ctx, prover, proof)
	if finalProofErr != nil {
		// just log the error and continue to handle the aggregated proof
		tmpLogger.Errorf("Failed trying to check if recursive proof can be verified: %v", finalProofErr)
	}

	// NOTE(pg): prover is done, use a.ctx from now on

	if !finalProofBuilt {
		proof.GeneratingSince = nil

		// final proof has not been generated, update the recursive proof
		err := a.state.UpdateGeneratedProof(a.ctx, proof, nil)
		if err != nil {
			err = fmt.Errorf("failed to store batch proof result, %w", err)
			tmpLogger.Error(FirstToUpper(err.Error()))
			return false, err
		}
	}

	return true, nil
}

func (a *Aggregator) getVerifiedBatchAccInputHash(ctx context.Context, batchNumber uint64) (*common.Hash, error) {
	accInputHash, err := a.etherman.GetBatchAccInputHash(ctx, batchNumber)
	if err != nil {
		return nil, err
	}

	return &accInputHash, nil
}

func (a *Aggregator) getAndLockBatchToProve(
	ctx context.Context, prover ProverInterface,
) (*state.Batch, []byte, *state.Proof, error) {
	proverID := prover.ID()
	proverName := prover.Name()

	tmpLogger := a.logger.WithFields(
		"prover", proverName,
		"proverId", proverID,
		"proverAddr", prover.Addr(),
	)

	a.stateDBMutex.Lock()
	defer a.stateDBMutex.Unlock()

	// Get last virtual batch number from L1
	lastVerifiedBatchNumber, err := a.etherman.GetLatestVerifiedBatchNum()
	if err != nil {
		return nil, nil, nil, err
	}

	proofExists := true
	batchNumberToVerify := lastVerifiedBatchNumber

	// Look for the batch number to verify
	for proofExists {
		batchNumberToVerify++
		proofExists, err = a.state.CheckProofExistsForBatch(ctx, batchNumberToVerify, nil)
		if err != nil {
			tmpLogger.Infof("Error checking proof exists for batch %d", batchNumberToVerify)

			return nil, nil, nil, err
		}

		if proofExists {
			accInputHash := a.getAccInputHash(batchNumberToVerify - 1)
			if accInputHash == (common.Hash{}) && batchNumberToVerify > 1 {
				tmpLogger.Warnf("AccInputHash for batch %d is not in memory, "+
					"deleting proofs to regenerate acc input hash chain in memory", batchNumberToVerify)

				err := a.state.CleanupGeneratedProofs(ctx, math.MaxInt, nil)
				if err != nil {
					tmpLogger.Infof("Error cleaning up generated proofs for batch %d", batchNumberToVerify)
					return nil, nil, nil, err
				}
				batchNumberToVerify--
				break
			}
		}
	}

	// Check if the batch has been sequenced
	sequence, err := a.l1Syncr.GetSequenceByBatchNumber(ctx, batchNumberToVerify)
	if err != nil && !errors.Is(err, entities.ErrNotFound) {
		return nil, nil, nil, err
	}

	// Not found, so it it not possible to verify the batch yet
	if sequence == nil || errors.Is(err, entities.ErrNotFound) {
		tmpLogger.Infof("Sequencing event for batch %d has not been synced yet, "+
			"so it is not possible to verify it yet. Waiting ...", batchNumberToVerify)

		return nil, nil, nil, state.ErrNotFound
	}

	stateSequence := state.Sequence{
		FromBatchNumber: sequence.FromBatchNumber,
		ToBatchNumber:   sequence.ToBatchNumber,
	}

	// Get Batch from L1 Syncer
	virtualBatch, err := a.l1Syncr.GetVirtualBatchByBatchNumber(a.ctx, batchNumberToVerify)
	if err != nil && !errors.Is(err, entities.ErrNotFound) {
		a.logger.Errorf("Error getting virtual batch: %v", err)
		return nil, nil, nil, err
	} else if errors.Is(err, entities.ErrNotFound) {
		a.logger.Infof("Virtual batch %d has not been synced yet, "+
			"so it is not possible to verify it yet. Waiting ...", batchNumberToVerify)
		return nil, nil, nil, state.ErrNotFound
	}

	// Get Batch from RPC
	rpcBatch, err := a.rpcClient.GetBatch(batchNumberToVerify)
	if err != nil {
		a.logger.Errorf("error getting batch %d from RPC: %v.", batchNumberToVerify, err)
		return nil, nil, nil, err
	}

	// Compare BatchL2Data from virtual batch and rpcBatch (skipping injected batch (1))
	if batchNumberToVerify != 1 && (common.Bytes2Hex(virtualBatch.BatchL2Data) != common.Bytes2Hex(rpcBatch.L2Data())) {
		a.logger.Warnf("BatchL2Data from virtual batch %d does not match the one from RPC", batchNumberToVerify)
		a.logger.Warnf("VirtualBatch BatchL2Data:%v", common.Bytes2Hex(virtualBatch.BatchL2Data))
		a.logger.Warnf("RPC BatchL2Data:%v", common.Bytes2Hex(rpcBatch.L2Data()))
	}

	l1InfoRoot := common.Hash{}

	if virtualBatch.L1InfoRoot == nil {
		log.Debugf("L1InfoRoot is nil for batch %d", batchNumberToVerify)
		virtualBatch.L1InfoRoot = &l1InfoRoot
	}

	// Ensure the old acc input hash is in memory
	oldAccInputHash := a.getAccInputHash(batchNumberToVerify - 1)
	if oldAccInputHash == (common.Hash{}) && batchNumberToVerify > 1 {
		tmpLogger.Warnf("AccInputHash for previous batch (%d) is not in memory. Waiting ...", batchNumberToVerify-1)
		return nil, nil, nil, state.ErrNotFound
	}

	forcedBlockHashL1 := rpcBatch.ForcedBlockHashL1()
	l1InfoRoot = *virtualBatch.L1InfoRoot

	if batchNumberToVerify == 1 {
		l1Block, err := a.l1Syncr.GetL1BlockByNumber(ctx, virtualBatch.BlockNumber)
		if err != nil {
			a.logger.Errorf("Error getting l1 block: %v", err)
			return nil, nil, nil, err
		}

		forcedBlockHashL1 = l1Block.ParentHash
		l1InfoRoot = rpcBatch.GlobalExitRoot()
	}

	// Calculate acc input hash as the RPC is not returning the correct one at the moment
	accInputHash := cdkcommon.CalculateAccInputHash(
		a.logger,
		oldAccInputHash,
		virtualBatch.BatchL2Data,
		l1InfoRoot,
		uint64(sequence.Timestamp.Unix()),
		rpcBatch.LastCoinbase(),
		forcedBlockHashL1,
	)
	// Store the acc input hash
	a.setAccInputHash(batchNumberToVerify, accInputHash)

	// Log params to calculate acc input hash
	a.logger.Debugf("Calculated acc input hash for batch %d: %v", batchNumberToVerify, accInputHash)
	a.logger.Debugf("OldAccInputHash: %v", oldAccInputHash)
	a.logger.Debugf("L1InfoRoot: %v", virtualBatch.L1InfoRoot)
	// a.logger.Debugf("LastL2BLockTimestamp: %v", rpcBatch.LastL2BLockTimestamp())
	a.logger.Debugf("TimestampLimit: %v", uint64(sequence.Timestamp.Unix()))
	a.logger.Debugf("LastCoinbase: %v", rpcBatch.LastCoinbase())
	a.logger.Debugf("ForcedBlockHashL1: %v", rpcBatch.ForcedBlockHashL1())

	// Create state batch
	stateBatch := &state.Batch{
		BatchNumber: rpcBatch.BatchNumber(),
		Coinbase:    rpcBatch.LastCoinbase(),
		// Use L1 batch data
		BatchL2Data:   virtualBatch.BatchL2Data,
		StateRoot:     rpcBatch.StateRoot(),
		LocalExitRoot: rpcBatch.LocalExitRoot(),
		// Use calculated acc input
		AccInputHash:    accInputHash,
		L1InfoTreeIndex: rpcBatch.L1InfoTreeIndex(),
		L1InfoRoot:      *virtualBatch.L1InfoRoot,
		Timestamp:       sequence.Timestamp,
		GlobalExitRoot:  rpcBatch.GlobalExitRoot(),
		ChainID:         a.cfg.ChainID,
		ForkID:          a.cfg.ForkId,
	}

	// Request the witness from the server, if it is busy just keep looping until it is available
	start := time.Now()
	witness, err := a.rpcClient.GetWitness(batchNumberToVerify, a.cfg.UseFullWitness)
	for err != nil {
		if errors.Is(err, rpc.ErrBusy) {
			a.logger.Debugf(
				"Witness server is busy, retrying get witness for batch %d in %v",
				batchNumberToVerify, a.cfg.RetryTime.Duration,
			)
		} else {
			a.logger.Errorf("Failed to get witness for batch %d, err: %v", batchNumberToVerify, err)
		}
		time.Sleep(a.cfg.RetryTime.Duration)
	}
	end := time.Now()
	a.logger.Debugf("Time to get witness for batch %d: %v", batchNumberToVerify, end.Sub(start))

	// Store the sequence in aggregator DB
	err = a.state.AddSequence(ctx, stateSequence, nil)
	if err != nil {
		tmpLogger.Infof("Error storing sequence for batch %d", batchNumberToVerify)

		return nil, nil, nil, err
	}

	// All the data required to generate a proof is ready
	tmpLogger.Infof("Found virtual batch %d pending to generate proof", virtualBatch.BatchNumber)
	tmpLogger = tmpLogger.WithFields("batch", virtualBatch.BatchNumber)

	tmpLogger.Info("Checking profitability to aggregate batch")

	// pass pol collateral as zero here, bcs in smart contract fee for aggregator is not defined yet
	isProfitable, err := a.profitabilityChecker.IsProfitable(ctx, big.NewInt(0))
	if err != nil {
		tmpLogger.Errorf("Failed to check aggregator profitability, err: %v", err)

		return nil, nil, nil, err
	}

	if !isProfitable {
		tmpLogger.Infof("Batch is not profitable, pol collateral %d", big.NewInt(0))

		return nil, nil, nil, err
	}

	now := time.Now().Round(time.Microsecond)
	proof := &state.Proof{
		BatchNumber:      virtualBatch.BatchNumber,
		BatchNumberFinal: virtualBatch.BatchNumber,
		Prover:           &proverName,
		ProverID:         &proverID,
		GeneratingSince:  &now,
	}

	// Avoid other prover to process the same batch
	err = a.state.AddGeneratedProof(ctx, proof, nil)
	if err != nil {
		tmpLogger.Errorf("Failed to add batch proof, err: %v", err)

		return nil, nil, nil, err
	}

	return stateBatch, witness, proof, nil
}

func (a *Aggregator) tryGenerateBatchProof(ctx context.Context, prover ProverInterface) (bool, error) {
	tmpLogger := a.logger.WithFields(
		"prover", prover.Name(),
		"proverId", prover.ID(),
		"proverAddr", prover.Addr(),
	)
	tmpLogger.Debug("tryGenerateBatchProof start")

	batchToProve, witness, proof, err0 := a.getAndLockBatchToProve(ctx, prover)
	if errors.Is(err0, state.ErrNotFound) || errors.Is(err0, entities.ErrNotFound) {
		// nothing to proof, swallow the error
		tmpLogger.Debug("Nothing to generate proof")
		return false, nil
	}
	if err0 != nil {
		return false, err0
	}

	tmpLogger = tmpLogger.WithFields("batch", batchToProve.BatchNumber)

	var (
		genProofID *string
		err        error
	)

	defer func() {
		if err != nil {
			tmpLogger.Debug("Deleting proof in progress")
			err2 := a.state.DeleteGeneratedProofs(a.ctx, proof.BatchNumber, proof.BatchNumberFinal, nil)
			if err2 != nil {
				tmpLogger.Errorf("Failed to delete proof in progress, err: %v", err2)
			}
		}
		tmpLogger.Debug("tryGenerateBatchProof end")
	}()

	tmpLogger.Infof("Sending zki + batch to the prover, batchNumber [%d]", batchToProve.BatchNumber)
	inputProver, err := a.buildInputProver(ctx, batchToProve, witness)
	if err != nil {
		err = fmt.Errorf("failed to build input prover, %w", err)
		tmpLogger.Error(FirstToUpper(err.Error()))
		return false, err
	}

	tmpLogger.Infof("Sending a batch to the prover. OldAccInputHash [%#x], L1InfoRoot [%#x]",
		inputProver.PublicInputs.OldAccInputHash, inputProver.PublicInputs.L1InfoRoot)

	genProofID, err = prover.BatchProof(inputProver)
	if err != nil {
		err = fmt.Errorf("failed to get batch proof id, %w", err)
		tmpLogger.Error(FirstToUpper(err.Error()))
		return false, err
	}

	proof.ProofID = genProofID

	tmpLogger = tmpLogger.WithFields("proofId", *proof.ProofID)

	resGetProof, stateRoot, accInputHash, err := prover.WaitRecursiveProof(ctx, *proof.ProofID)
	if err != nil {
		err = fmt.Errorf("failed to get proof from prover, %w", err)
		tmpLogger.Error(FirstToUpper(err.Error()))
		return false, err
	}

	tmpLogger.Info("Batch proof generated")

	// Sanity Check: state root from the proof must match the one from the batch
	if a.cfg.BatchProofSanityCheckEnabled && (stateRoot != common.Hash{}) && (stateRoot != batchToProve.StateRoot) {
		for {
			tmpLogger.Errorf("HALTING: "+
				"State root from the proof does not match the expected for batch %d: Proof = [%s] Expected = [%s]",
				batchToProve.BatchNumber, stateRoot.String(), batchToProve.StateRoot.String(),
			)
			time.Sleep(a.cfg.RetryTime.Duration)
		}
	} else {
		tmpLogger.Infof("State root sanity check for batch %d passed", batchToProve.BatchNumber)
	}

	// Sanity Check: acc input hash from the proof must match the one from the batch
	if a.cfg.BatchProofSanityCheckEnabled && (accInputHash != common.Hash{}) &&
		(accInputHash != batchToProve.AccInputHash) {
		for {
			tmpLogger.Errorf("HALTING: Acc input hash from the proof does not match the expected for "+
				"batch %d: Proof = [%s] Expected = [%s]",
				batchToProve.BatchNumber, accInputHash.String(), batchToProve.AccInputHash.String(),
			)
			time.Sleep(a.cfg.RetryTime.Duration)
		}
	} else {
		tmpLogger.Infof("Acc input hash sanity check for batch %d passed", batchToProve.BatchNumber)
	}

	proof.Proof = resGetProof

	// NOTE(pg): the defer func is useless from now on, use a different variable
	// name for errors (or shadow err in inner scopes) to not trigger it.

	finalProofBuilt, finalProofErr := a.tryBuildFinalProof(ctx, prover, proof)
	if finalProofErr != nil {
		// just log the error and continue to handle the generated proof
		tmpLogger.Errorf("Error trying to build final proof: %v", finalProofErr)
	}

	// NOTE(pg): prover is done, use a.ctx from now on

	if !finalProofBuilt {
		proof.GeneratingSince = nil

		// final proof has not been generated, update the batch proof
		err := a.state.UpdateGeneratedProof(a.ctx, proof, nil)
		if err != nil {
			err = fmt.Errorf("failed to store batch proof result, %w", err)
			tmpLogger.Error(FirstToUpper(err.Error()))
			return false, err
		}
	}

	return true, nil
}

// canVerifyProof returns true if we have reached the timeout to verify a proof
// and no other prover is verifying a proof (verifyingProof = false).
func (a *Aggregator) canVerifyProof() bool {
	a.timeSendFinalProofMutex.RLock()
	defer a.timeSendFinalProofMutex.RUnlock()

	return a.timeSendFinalProof.Before(time.Now()) && !a.verifyingProof
}

// startProofVerification sets the verifyingProof variable to true
// to indicate that there is a proof verification in progress.
func (a *Aggregator) startProofVerification() {
	a.timeSendFinalProofMutex.Lock()
	defer a.timeSendFinalProofMutex.Unlock()
	a.verifyingProof = true
}

// endProofVerification set verifyingProof to false to indicate that there is not proof verification in progress
func (a *Aggregator) endProofVerification() {
	a.timeSendFinalProofMutex.Lock()
	defer a.timeSendFinalProofMutex.Unlock()
	a.verifyingProof = false
}

// resetVerifyProofTime updates the timeout to verify a proof.
func (a *Aggregator) resetVerifyProofTime() {
	a.timeSendFinalProofMutex.Lock()
	defer a.timeSendFinalProofMutex.Unlock()
	a.timeSendFinalProof = time.Now().Add(a.cfg.VerifyProofInterval.Duration)
}

func (a *Aggregator) buildInputProver(
	ctx context.Context, batchToVerify *state.Batch, witness []byte,
) (*prover.StatelessInputProver, error) {
	isForcedBatch := false
	batchRawData := &state.BatchRawV2{}
	var err error

	if batchToVerify.BatchNumber == 1 || batchToVerify.ForcedBatchNum != nil {
		isForcedBatch = true
	} else {
		batchRawData, err = state.DecodeBatchV2(batchToVerify.BatchL2Data)
		if err != nil {
			a.logger.Errorf("Failed to decode batch data, err: %v", err)
			return nil, err
		}
	}

	l1InfoTreeData := map[uint32]*prover.L1Data{}
	forcedBlockhashL1 := common.Hash{}
	l1InfoRoot := batchToVerify.L1InfoRoot.Bytes()
	//nolint:gocritic
	if !isForcedBatch {
		tree, err := l1infotree.NewL1InfoTree(a.logger, 32, [][32]byte{}) //nolint:mnd
		if err != nil {
			return nil, err
		}

		leaves, err := a.l1Syncr.GetLeafsByL1InfoRoot(ctx, batchToVerify.L1InfoRoot)
		if err != nil && !errors.Is(err, entities.ErrNotFound) {
			return nil, err
		}

		aLeaves := make([][32]byte, len(leaves))
		for i, leaf := range leaves {
			aLeaves[i] = l1infotree.HashLeafData(
				leaf.GlobalExitRoot,
				leaf.PreviousBlockHash,
				uint64(leaf.Timestamp.Unix()))
		}

		for _, l2blockRaw := range batchRawData.Blocks {
			_, contained := l1InfoTreeData[l2blockRaw.IndexL1InfoTree]
			if !contained && l2blockRaw.IndexL1InfoTree != 0 {
				leaves, err := a.l1Syncr.GetL1InfoTreeLeaves(ctx, []uint32{l2blockRaw.IndexL1InfoTree})
				if err != nil {
					a.logger.Errorf("Error getting l1InfoTreeLeaf: %v", err)
					return nil, err
				}

				l1InfoTreeLeaf := leaves[l2blockRaw.IndexL1InfoTree]

				// Calculate smt proof
				a.logger.Infof("Calling tree.ComputeMerkleProof")
				smtProof, calculatedL1InfoRoot, err := tree.ComputeMerkleProof(l2blockRaw.IndexL1InfoTree, aLeaves)
				if err != nil {
					a.logger.Errorf("Error computing merkle proof: %v", err)
					return nil, err
				}

				if batchToVerify.L1InfoRoot != calculatedL1InfoRoot {
					return nil, fmt.Errorf(
						"error: l1InfoRoot mismatch. L1InfoRoot: %s, calculatedL1InfoRoot: %s. l1InfoTreeIndex: %d",
						batchToVerify.L1InfoRoot.String(), calculatedL1InfoRoot.String(), l2blockRaw.IndexL1InfoTree,
					)
				}

				protoProof := make([][]byte, len(smtProof))

				for i, proof := range smtProof {
					tmpProof := proof
					protoProof[i] = tmpProof[:]
				}

				l1InfoTreeData[l2blockRaw.IndexL1InfoTree] = &prover.L1Data{
					GlobalExitRoot: l1InfoTreeLeaf.GlobalExitRoot.Bytes(),
					BlockhashL1:    l1InfoTreeLeaf.PreviousBlockHash.Bytes(),
					MinTimestamp:   uint32(l1InfoTreeLeaf.Timestamp.Unix()),
					SmtProof:       protoProof,
				}
			}
		}
	} else {
		// Initial batch must be handled differently
		if batchToVerify.BatchNumber == 1 {
			virtualBatch, err := a.l1Syncr.GetVirtualBatchByBatchNumber(ctx, batchToVerify.BatchNumber)
			if err != nil {
				a.logger.Errorf("Error getting virtual batch: %v", err)
				return nil, err
			}
			l1Block, err := a.l1Syncr.GetL1BlockByNumber(ctx, virtualBatch.BlockNumber)
			if err != nil {
				a.logger.Errorf("Error getting l1 block: %v", err)
				return nil, err
			}

			forcedBlockhashL1 = l1Block.ParentHash
			l1InfoRoot = batchToVerify.GlobalExitRoot.Bytes()
		}
	}

	// Ensure the old acc input hash is in memory
	oldAccInputHash := a.getAccInputHash(batchToVerify.BatchNumber - 1)
	if oldAccInputHash == (common.Hash{}) && batchToVerify.BatchNumber > 1 {
		a.logger.Warnf("AccInputHash for previous batch (%d) is not in memory. Waiting ...", batchToVerify.BatchNumber-1)
		return nil, fmt.Errorf("acc input hash for previous batch (%d) is not in memory", batchToVerify.BatchNumber-1)
	}

	inputProver := &prover.StatelessInputProver{
		PublicInputs: &prover.StatelessPublicInputs{
			Witness:           witness,
			OldAccInputHash:   oldAccInputHash.Bytes(),
			OldBatchNum:       batchToVerify.BatchNumber - 1,
			ChainId:           batchToVerify.ChainID,
			ForkId:            batchToVerify.ForkID,
			BatchL2Data:       batchToVerify.BatchL2Data,
			L1InfoRoot:        l1InfoRoot,
			TimestampLimit:    uint64(batchToVerify.Timestamp.Unix()),
			SequencerAddr:     batchToVerify.Coinbase.String(),
			AggregatorAddr:    a.cfg.SenderAddress,
			L1InfoTreeData:    l1InfoTreeData,
			ForcedBlockhashL1: forcedBlockhashL1.Bytes(),
		},
	}

	printInputProver(a.logger, inputProver)
	return inputProver, nil
}

func printInputProver(logger *log.Logger, inputProver *prover.StatelessInputProver) {
	if !logger.IsEnabledLogLevel(zapcore.DebugLevel) {
		return
	}

	logger.Debugf("Witness length: %v", len(inputProver.PublicInputs.Witness))
	logger.Debugf("BatchL2Data length: %v", len(inputProver.PublicInputs.BatchL2Data))
	// logger.Debugf("Full DataStream: %v", common.Bytes2Hex(inputProver.PublicInputs.DataStream))
	logger.Debugf("OldAccInputHash: %v", common.BytesToHash(inputProver.PublicInputs.OldAccInputHash))
	logger.Debugf("L1InfoRoot: %v", common.BytesToHash(inputProver.PublicInputs.L1InfoRoot))
	logger.Debugf("TimestampLimit: %v", inputProver.PublicInputs.TimestampLimit)
	logger.Debugf("SequencerAddr: %v", inputProver.PublicInputs.SequencerAddr)
	logger.Debugf("AggregatorAddr: %v", inputProver.PublicInputs.AggregatorAddr)
	logger.Debugf("L1InfoTreeData: %+v", inputProver.PublicInputs.L1InfoTreeData)
	logger.Debugf("ForcedBlockhashL1: %v", common.BytesToHash(inputProver.PublicInputs.ForcedBlockhashL1))
}

// healthChecker will provide an implementation of the HealthCheck interface.
type healthChecker struct{}

// newHealthChecker returns a health checker according to standard package
// grpc.health.v1.
func newHealthChecker() *healthChecker {
	return &healthChecker{}
}

// HealthCheck interface implementation.

// Check returns the current status of the server for unary gRPC health requests,
// for now if the server is up and able to respond we will always return SERVING.
func (hc *healthChecker) Check(
	ctx context.Context, req *grpchealth.HealthCheckRequest,
) (*grpchealth.HealthCheckResponse, error) {
	log.Info("Serving the Check request for health check")

	return &grpchealth.HealthCheckResponse{
		Status: grpchealth.HealthCheckResponse_SERVING,
	}, nil
}

// Watch returns the current status of the server for stream gRPC health requests,
// for now if the server is up and able to respond we will always return SERVING.
func (hc *healthChecker) Watch(req *grpchealth.HealthCheckRequest, server grpchealth.Health_WatchServer) error {
	log.Info("Serving the Watch request for health check")

	return server.Send(&grpchealth.HealthCheckResponse{
		Status: grpchealth.HealthCheckResponse_SERVING,
	})
}

func (a *Aggregator) handleMonitoredTxResult(result ethtxtypes.MonitoredTxResult, firstBatch, lastBatch uint64) {
	mTxResultLogger := ethtxmanager.CreateMonitoredTxResultLogger(result)
	if result.Status == ethtxtypes.MonitoredTxStatusFailed {
		mTxResultLogger.Fatal("failed to send batch verification, TODO: review this fatal and define what to do in this case")
	}

	// Wait for the transaction to be finalized, then we can safely delete all recursive
	// proofs up to the last batch in this proof

	finaLizedBlockNumber, err := l1_check_block.L1FinalizedFetch.BlockNumber(a.ctx, a.etherman)
	if err != nil {
		mTxResultLogger.Errorf("failed to get finalized block number: %v", err)
	}

	for result.MinedAtBlockNumber.Uint64() > finaLizedBlockNumber {
		select {
		case <-a.ctx.Done():
			return
		case <-time.After(a.cfg.RetryTime.Duration):
			finaLizedBlockNumber, err = l1_check_block.L1FinalizedFetch.BlockNumber(a.ctx, a.etherman)
			if err != nil {
				mTxResultLogger.Errorf("failed to get finalized block number: %v", err)
			}
		}
	}

	err = a.state.DeleteGeneratedProofs(a.ctx, firstBatch, lastBatch, nil)
	if err != nil {
		mTxResultLogger.Errorf("failed to delete generated proofs from %d to %d: %v", firstBatch, lastBatch, err)
	}

	mTxResultLogger.Debugf("deleted generated proofs from %d to %d", firstBatch, lastBatch)

	// Remove the acc input hashes from the map
	// leaving the last batch acc input hash as it will be used as old acc input hash
	a.removeAccInputHashes(firstBatch, lastBatch-1)
}

func (a *Aggregator) cleanupLockedProofs() {
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-time.After(a.timeCleanupLockedProofs.Duration):
			n, err := a.state.CleanupLockedProofs(a.ctx, a.cfg.GeneratingProofCleanupThreshold, nil)
			if err != nil {
				a.logger.Errorf("Failed to cleanup locked proofs: %v", err)
			}
			if n == 1 {
				a.logger.Warn("Found a stale proof and removed from cache")
			} else if n > 1 {
				a.logger.Warnf("Found %d stale proofs and removed from cache", n)
			}
		}
	}
}

// FirstToUpper returns the string passed as argument with the first letter in
// uppercase.
func FirstToUpper(s string) string {
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}
