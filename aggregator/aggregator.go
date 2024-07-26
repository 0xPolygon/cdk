package aggregator

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/0xPolygon/cdk-rpc/rpc"
	cdkTypes "github.com/0xPolygon/cdk-rpc/types"
	ethmanTypes "github.com/0xPolygon/cdk/aggregator/ethmantypes"
	"github.com/0xPolygon/cdk/aggregator/prover"
	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/l1infotree"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/state"
	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/0xPolygonHermez/zkevm-data-streamer/datastreamer"
	streamlog "github.com/0xPolygonHermez/zkevm-data-streamer/log"
	"github.com/0xPolygonHermez/zkevm-ethtx-manager/ethtxmanager"
	ethtxlog "github.com/0xPolygonHermez/zkevm-ethtx-manager/log"
	synclog "github.com/0xPolygonHermez/zkevm-synchronizer-l1/log"
	"github.com/0xPolygonHermez/zkevm-synchronizer-l1/state/entities"
	"github.com/0xPolygonHermez/zkevm-synchronizer-l1/synchronizer"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/proto"
)

const (
	dataStreamType      = 1
	mockedStateRoot     = "0x090bcaf734c4f06c93954a827b45a6e8c67b8e0fd1e0a35a1c5982d6961828f9"
	mockedLocalExitRoot = "0x17c04c3760510b48c6012742c540a81aba4bca2f78b9d14bfd2f123e2e53ea3e"
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

	cfg Config

	state        stateInterface
	etherman     etherman
	ethTxManager *ethtxmanager.Client
	streamClient *datastreamer.StreamClient
	l1Syncr      synchronizer.Synchronizer
	halted       atomic.Bool

	profitabilityChecker    aggregatorTxProfitabilityChecker
	timeSendFinalProof      time.Time
	timeCleanupLockedProofs types.Duration
	stateDBMutex            *sync.Mutex
	timeSendFinalProofMutex *sync.RWMutex

	// Data stream handling variables
	currentBatchStreamData []byte
	currentStreamBatch     state.Batch
	currentStreamBatchRaw  state.BatchRawV2
	currentStreamL2Block   state.L2BlockRaw

	finalProof     chan finalProofMsg
	verifyingProof bool

	srv  *grpc.Server
	ctx  context.Context
	exit context.CancelFunc

	sequencerPrivateKey *ecdsa.PrivateKey
	aggLayerClient      AgglayerClientInterface
}

// New creates a new aggregator.
func New(
	ctx context.Context,
	cfg Config,
	stateInterface stateInterface,
	etherman etherman) (*Aggregator, error) {
	var profitabilityChecker aggregatorTxProfitabilityChecker

	switch cfg.TxProfitabilityCheckerType {
	case ProfitabilityBase:
		profitabilityChecker = NewTxProfitabilityCheckerBase(stateInterface, cfg.IntervalAfterWhichBatchConsolidateAnyway.Duration, cfg.TxProfitabilityMinReward.Int)
	case ProfitabilityAcceptAll:
		profitabilityChecker = NewTxProfitabilityCheckerAcceptAll(stateInterface, cfg.IntervalAfterWhichBatchConsolidateAnyway.Duration)
	}

	// Create ethtxmanager client
	cfg.EthTxManager.Log = ethtxlog.Config{
		Environment: ethtxlog.LogEnvironment(cfg.Log.Environment),
		Level:       cfg.Log.Level,
		Outputs:     cfg.Log.Outputs,
	}
	ethTxManager, err := ethtxmanager.New(cfg.EthTxManager)
	if err != nil {
		log.Fatalf("error creating ethtxmanager client: %v", err)
	}

	// Data stream client logs
	streamLogConfig := streamlog.Config{
		Environment: streamlog.LogEnvironment(cfg.Log.Environment),
		Level:       cfg.Log.Level,
		Outputs:     cfg.Log.Outputs,
	}

	log.Init(cfg.Log)

	log.Info("Creating data stream client....")
	streamClient, err := datastreamer.NewClientWithLogsConfig(cfg.StreamClient.Server, dataStreamType, streamLogConfig)
	if err != nil {
		log.Fatalf("failed to create stream client, error: %v", err)
	}
	log.Info("Data stream client created.")

	// Synchonizer logs
	syncLogConfig := synclog.Config{
		Environment: synclog.LogEnvironment(cfg.Log.Environment),
		Level:       cfg.Log.Level,
		Outputs:     cfg.Log.Outputs,
	}

	cfg.Synchronizer.Log = syncLogConfig

	// Create L1 synchronizer client
	cfg.Synchronizer.Etherman.L1URL = cfg.EthTxManager.Etherman.URL
	log.Debugf("Creating synchronizer client with config: %+v", cfg.Synchronizer)
	l1Syncr, err := synchronizer.NewSynchronizer(ctx, cfg.Synchronizer)
	if err != nil {
		log.Fatalf("failed to create synchronizer client, error: %v", err)
	}

	var (
		aggLayerClient      AgglayerClientInterface
		sequencerPrivateKey *ecdsa.PrivateKey
	)

	if cfg.SettlementBackend == AggLayer {
		aggLayerClient = NewAggLayerClient(cfg.AggLayerURL)

		sequencerPrivateKey, err = newKeyFromKeystore(cfg.SequencerPrivateKey)
		if err != nil {
			return nil, err
		}
	}

	a := &Aggregator{
		cfg:                     cfg,
		state:                   stateInterface,
		etherman:                etherman,
		ethTxManager:            ethTxManager,
		streamClient:            streamClient,
		l1Syncr:                 l1Syncr,
		profitabilityChecker:    profitabilityChecker,
		stateDBMutex:            &sync.Mutex{},
		timeSendFinalProofMutex: &sync.RWMutex{},
		timeCleanupLockedProofs: cfg.CleanupLockedProofsInterval,
		finalProof:              make(chan finalProofMsg),
		currentBatchStreamData:  []byte{},
		aggLayerClient:          aggLayerClient,
		sequencerPrivateKey:     sequencerPrivateKey,
	}

	// Set function to handle the batches from the data stream
	a.streamClient.SetProcessEntryFunc(a.handleReceivedDataStream)
	a.l1Syncr.SetCallbackOnReorgDone(a.handleReorg)

	return a, nil
}

func (a *Aggregator) handleReorg(reorgData synchronizer.ReorgExecutionResult) {
	log.Warnf("Reorg detected, reorgData: %+v", reorgData)

	ctx := context.Background()

	// Get new latest verified batch number
	lastVBatchNumber, err := a.l1Syncr.GetLastestVirtualBatchNumber(ctx)
	if err != nil {
		log.Errorf("Error getting last virtual batch number: %v", err)
	} else {
		err = a.state.DeleteBatchesNewerThanBatchNumber(ctx, lastVBatchNumber, nil)
		if err != nil {
			log.Errorf("Error deleting batches newer than batch number %d: %v", lastVBatchNumber, err)
		}
	}

	// Halt the aggregator
	a.halted.Store(true)
	for {
		log.Warnf("Halting the aggregator due to a L1 reorg. Reorged data has been delete so it is safe to manually restart the aggregator.")
		time.Sleep(10 * time.Second) // nolint:gomnd
	}
}

func (a *Aggregator) handleReceivedDataStream(entry *datastreamer.FileEntry, client *datastreamer.StreamClient, server *datastreamer.StreamServer) error {
	ctx := context.Background()
	forcedBlockhashL1 := common.Hash{}

	if !a.halted.Load() {
		if entry.Type != datastreamer.EntryType(datastreamer.EtBookmark) {
			a.currentBatchStreamData = append(a.currentBatchStreamData, entry.Encode()...)

			switch entry.Type {
			case datastreamer.EntryType(datastream.EntryType_ENTRY_TYPE_BATCH_START):
				batch := &datastream.BatchStart{}
				err := proto.Unmarshal(entry.Data, batch)
				if err != nil {
					log.Errorf("Error unmarshalling batch: %v", err)
					return err
				}

				a.currentStreamBatch.BatchNumber = batch.Number
				a.currentStreamBatch.ChainID = batch.ChainId
				a.currentStreamBatch.ForkID = batch.ForkId
				a.currentStreamBatch.Type = batch.Type
			case datastreamer.EntryType(datastream.EntryType_ENTRY_TYPE_BATCH_END):
				batch := &datastream.BatchEnd{}
				err := proto.Unmarshal(entry.Data, batch)
				if err != nil {
					log.Errorf("Error unmarshalling batch: %v", err)
					return err
				}

				a.currentStreamBatch.LocalExitRoot = common.BytesToHash(batch.LocalExitRoot)
				a.currentStreamBatch.StateRoot = common.BytesToHash(batch.StateRoot)

				// Add last block (if any) to the current batch
				if a.currentStreamL2Block.BlockNumber != 0 {
					a.currentStreamBatchRaw.Blocks = append(a.currentStreamBatchRaw.Blocks, a.currentStreamL2Block)
				}

				// Save Current Batch
				if a.currentStreamBatch.BatchNumber != 0 {
					var batchl2Data []byte

					// Get batchl2Data from L1
					virtualBatch, err := a.l1Syncr.GetVirtualBatchByBatchNumber(ctx, a.currentStreamBatch.BatchNumber)
					if err != nil && !errors.Is(err, entities.ErrNotFound) {
						log.Errorf("Error getting virtual batch: %v", err)
						return err
					}

					for errors.Is(err, entities.ErrNotFound) {
						log.Debug("Waiting for virtual batch to be available")
						time.Sleep(a.cfg.RetryTime.Duration)
						virtualBatch, err = a.l1Syncr.GetVirtualBatchByBatchNumber(ctx, a.currentStreamBatch.BatchNumber)

						if err != nil && !errors.Is(err, entities.ErrNotFound) {
							log.Errorf("Error getting virtual batch: %v", err)
							return err
						}
					}

					// Encode batch
					if a.currentStreamBatch.Type != datastream.BatchType_BATCH_TYPE_INVALID {
						batchl2Data, err = state.EncodeBatchV2(&a.currentStreamBatchRaw)
						if err != nil {
							log.Errorf("Error encoding batch: %v", err)
							return err
						}
					}

					// If the batch is marked as Invalid in the DS we enforce retrieve the data from L1
					if a.cfg.UseL1BatchData || a.currentStreamBatch.Type == datastream.BatchType_BATCH_TYPE_INVALID {
						a.currentStreamBatch.BatchL2Data = virtualBatch.BatchL2Data
					} else {
						a.currentStreamBatch.BatchL2Data = batchl2Data
					}

					// Compare BatchL2Data from L1 and DataStream
					if common.Bytes2Hex(batchl2Data) != common.Bytes2Hex(virtualBatch.BatchL2Data) {
						log.Warnf("BatchL2Data from L1 and data stream are different for batch %d", a.currentStreamBatch.BatchNumber)

						if a.currentStreamBatch.Type == datastream.BatchType_BATCH_TYPE_INVALID {
							log.Warnf("Batch is marked as invalid in data stream")
						} else {
							log.Warnf("DataStream BatchL2Data:%v", common.Bytes2Hex(batchl2Data))
						}
						log.Warnf("L1 BatchL2Data:%v", common.Bytes2Hex(virtualBatch.BatchL2Data))
					}

					// Get L1InfoRoot
					sequence, err := a.l1Syncr.GetSequenceByBatchNumber(ctx, a.currentStreamBatch.BatchNumber)
					if err != nil {
						log.Errorf("Error getting sequence: %v", err)
						return err
					}

					for sequence == nil {
						log.Debug("Waiting for sequence to be available")
						time.Sleep(a.cfg.RetryTime.Duration)
						sequence, err = a.l1Syncr.GetSequenceByBatchNumber(ctx, a.currentStreamBatch.BatchNumber)
						if err != nil {
							log.Errorf("Error getting sequence: %v", err)
							return err
						}
					}

					a.currentStreamBatch.L1InfoRoot = sequence.L1InfoRoot
					a.currentStreamBatch.Timestamp = sequence.Timestamp

					// Calculate Acc Input Hash
					oldBatch, _, err := a.state.GetBatch(ctx, a.currentStreamBatch.BatchNumber-1, nil)
					if err != nil {
						log.Errorf("Error getting batch %d: %v", a.currentStreamBatch.BatchNumber-1, err)
						return err
					}

					// Injected Batch
					if a.currentStreamBatch.BatchNumber == 1 {
						l1Block, err := a.l1Syncr.GetL1BlockByNumber(ctx, virtualBatch.BlockNumber)
						if err != nil {
							log.Errorf("Error getting L1 block: %v", err)
							return err
						}

						forcedBlockhashL1 = l1Block.ParentHash
						a.currentStreamBatch.L1InfoRoot = a.currentStreamBatch.GlobalExitRoot
					}

					accInputHash, err := cdkcommon.CalculateAccInputHash(oldBatch.AccInputHash, a.currentStreamBatch.BatchL2Data, a.currentStreamBatch.L1InfoRoot, uint64(a.currentStreamBatch.Timestamp.Unix()), a.currentStreamBatch.Coinbase, forcedBlockhashL1)
					if err != nil {
						log.Errorf("Error calculating acc input hash: %v", err)
						return err
					}

					a.currentStreamBatch.AccInputHash = accInputHash

					err = a.state.AddBatch(ctx, &a.currentStreamBatch, a.currentBatchStreamData, nil)
					if err != nil {
						log.Errorf("Error adding batch: %v", err)
						return err
					}
				}

				// Reset current batch data
				a.currentBatchStreamData = []byte{}
				a.currentStreamBatchRaw = state.BatchRawV2{
					Blocks: make([]state.L2BlockRaw, 0),
				}
				a.currentStreamL2Block = state.L2BlockRaw{}

			case datastreamer.EntryType(datastream.EntryType_ENTRY_TYPE_L2_BLOCK):
				// Add previous block (if any) to the current batch
				if a.currentStreamL2Block.BlockNumber != 0 {
					a.currentStreamBatchRaw.Blocks = append(a.currentStreamBatchRaw.Blocks, a.currentStreamL2Block)
				}
				// "Open" the new block
				l2Block := &datastream.L2Block{}
				err := proto.Unmarshal(entry.Data, l2Block)
				if err != nil {
					log.Errorf("Error unmarshalling L2Block: %v", err)
					return err
				}

				header := state.ChangeL2BlockHeader{
					DeltaTimestamp:  l2Block.DeltaTimestamp,
					IndexL1InfoTree: l2Block.L1InfotreeIndex,
				}

				a.currentStreamL2Block.ChangeL2BlockHeader = header
				a.currentStreamL2Block.Transactions = make([]state.L2TxRaw, 0)
				a.currentStreamL2Block.BlockNumber = l2Block.Number
				a.currentStreamBatch.L1InfoTreeIndex = l2Block.L1InfotreeIndex
				a.currentStreamBatch.Coinbase = common.BytesToAddress(l2Block.Coinbase)
				a.currentStreamBatch.GlobalExitRoot = common.BytesToHash(l2Block.GlobalExitRoot)

			case datastreamer.EntryType(datastream.EntryType_ENTRY_TYPE_TRANSACTION):
				l2Tx := &datastream.Transaction{}
				err := proto.Unmarshal(entry.Data, l2Tx)
				if err != nil {
					log.Errorf("Error unmarshalling L2Tx: %v", err)
					return err
				}
				// New Tx raw
				tx, err := state.DecodeTx(common.Bytes2Hex(l2Tx.Encoded))
				if err != nil {
					log.Errorf("Error decoding tx: %v", err)
					return err
				}

				l2TxRaw := state.L2TxRaw{
					EfficiencyPercentage: uint8(l2Tx.EffectiveGasPricePercentage),
					TxAlreadyEncoded:     false,
					Tx:                   tx,
				}
				a.currentStreamL2Block.Transactions = append(a.currentStreamL2Block.Transactions, l2TxRaw)
			}
		}
	}
	return nil
}

// Start starts the aggregator
func (a *Aggregator) Start(ctx context.Context) error {
	var cancel context.CancelFunc
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel = context.WithCancel(ctx)
	a.ctx = ctx
	a.exit = cancel

	address := fmt.Sprintf("%s:%d", a.cfg.Host, a.cfg.Port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	a.srv = grpc.NewServer()
	prover.RegisterAggregatorServiceServer(a.srv, a)

	healthService := newHealthChecker()
	grpchealth.RegisterHealthServer(a.srv, healthService)

	// Initial L1 Sync blocking
	err = a.l1Syncr.Sync(true)
	if err != nil {
		log.Fatalf("Failed to synchronize from L1: %v", err)
		return err
	}

	// Get last verified batch number to set the starting point for verifications
	lastVerifiedBatchNumber, err := a.etherman.GetLatestVerifiedBatchNum()
	if err != nil {
		return err
	}

	// Cleanup data base
	err = a.state.DeleteBatchesOlderThanBatchNumber(ctx, lastVerifiedBatchNumber, nil)
	if err != nil {
		return err
	}

	// Delete ungenerated recursive proofs
	err = a.state.DeleteUngeneratedProofs(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize proofs cache %w", err)
	}

	accInputHash, err := a.getVerifiedBatchAccInputHash(ctx, lastVerifiedBatchNumber)
	if err != nil {
		return err
	}

	log.Infof("Last Verified Batch Number:%v", lastVerifiedBatchNumber)
	log.Infof("Starting AccInputHash:%v", accInputHash.String())

	// Store Acc Input Hash of the latest verified batch
	dummyBatch := state.Batch{BatchNumber: lastVerifiedBatchNumber, AccInputHash: *accInputHash}
	err = a.state.AddBatch(ctx, &dummyBatch, []byte{0}, nil)
	if err != nil {
		return err
	}

	a.resetVerifyProofTime()

	go a.cleanupLockedProofs()
	go a.sendFinalProof()
	go a.ethTxManager.Start()

	// Keep syncing L1
	go func() {
		err := a.l1Syncr.Sync(false)
		if err != nil {
			log.Fatalf("Failed to synchronize from L1: %v", err)
		}
	}()

	// Start stream client
	err = a.streamClient.Start()
	if err != nil {
		log.Fatalf("failed to start stream client, error: %v", err)
	}

	bookMark := &datastream.BookMark{
		Type:  datastream.BookmarkType_BOOKMARK_TYPE_BATCH,
		Value: lastVerifiedBatchNumber + 1,
	}

	marshalledBookMark, err := proto.Marshal(bookMark)
	if err != nil {
		log.Fatalf("failed to marshal bookmark: %v", err)
	}

	err = a.streamClient.ExecCommandStartBookmark(marshalledBookMark)
	if err != nil {
		log.Fatalf("failed to connect to data stream: %v", err)
	}

	// A this point everything is ready, so start serving
	go func() {
		log.Infof("Server listening on port %d", a.cfg.Port)
		if err := a.srv.Serve(lis); err != nil {
			a.exit()
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	<-ctx.Done()
	return ctx.Err()
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
	prover, err := prover.New(stream, proverAddr, a.cfg.ProofStatePollingInterval)
	if err != nil {
		return err
	}

	log := log.WithFields(
		"prover", prover.Name(),
		"proverId", prover.ID(),
		"proverAddr", prover.Addr(),
	)
	log.Info("Establishing stream connection with prover")

	// Check if prover supports the required Fork ID
	if !prover.SupportsForkID(a.cfg.ForkId) {
		err := errors.New("prover does not support required fork ID")
		log.Warn(FirstToUpper(err.Error()))
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
					log.Errorf("Failed to check if prover is idle: %v", err)
					time.Sleep(a.cfg.RetryTime.Duration)
					continue
				}
				if !isIdle {
					log.Debug("Prover is not idle")
					time.Sleep(a.cfg.RetryTime.Duration)
					continue
				}

				_, err = a.tryBuildFinalProof(ctx, prover, nil)
				if err != nil {
					log.Errorf("Error checking proofs to verify: %v", err)
				}

				proofGenerated, err := a.tryAggregateProofs(ctx, prover)
				if err != nil {
					log.Errorf("Error trying to aggregate proofs: %v", err)
				}

				if !proofGenerated {
					proofGenerated, err = a.tryGenerateBatchProof(ctx, prover)
					if err != nil {
						log.Errorf("Error trying to generate proof: %v", err)
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

			log.WithFields("proofId", proof.ProofID, "batches", fmt.Sprintf("%d-%d", proof.BatchNumber, proof.BatchNumberFinal))
			log.Info("Verifying final proof with ethereum smart contract")

			a.startProofVerification()

			finalBatch, _, err := a.state.GetBatch(ctx, proof.BatchNumberFinal, nil)
			if err != nil {
				log.Errorf("Failed to retrieve batch with number [%d]: %v", proof.BatchNumberFinal, err)
				a.endProofVerification()
				continue
			}

			inputs := ethmanTypes.FinalProofInputs{
				FinalProof:       msg.finalProof,
				NewLocalExitRoot: finalBatch.LocalExitRoot.Bytes(),
				NewStateRoot:     finalBatch.StateRoot.Bytes(),
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
	tx := Tx{
		LastVerifiedBatch: cdkTypes.ArgUint64(proof.BatchNumber - 1),
		NewVerifiedBatch:  cdkTypes.ArgUint64(proof.BatchNumberFinal),
		ZKP: ZKP{
			NewStateRoot:     common.BytesToHash(inputs.NewStateRoot),
			NewLocalExitRoot: common.BytesToHash(inputs.NewLocalExitRoot),
			Proof:            cdkTypes.ArgBytes(proofBytes),
		},
		RollupID: a.etherman.GetRollupId(),
	}
	signedTx, err := tx.Sign(a.sequencerPrivateKey)
	if err != nil {
		log.Errorf("failed to sign tx: %v", err)
		a.handleFailureToAddVerifyBatchToBeMonitored(ctx, proof)

		return false
	}

	log.Debug("final proof signedTx: ", signedTx.Tx.ZKP.Proof.Hex())
	txHash, err := a.aggLayerClient.SendTx(*signedTx)
	if err != nil {
		log.Errorf("failed to send tx to the agglayer: %v", err)
		a.handleFailureToAddVerifyBatchToBeMonitored(ctx, proof)

		return false
	}

	log.Infof("tx %s sent to agglayer, waiting to be mined", txHash.Hex())
	log.Debugf("Timeout set to %f seconds", a.cfg.AggLayerTxTimeout.Duration.Seconds())
	waitCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(a.cfg.AggLayerTxTimeout.Duration))
	defer cancelFunc()
	if err := a.aggLayerClient.WaitTxToBeMined(txHash, waitCtx); err != nil {
		log.Errorf("agglayer didn't mine the tx: %v", err)
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
	to, data, err := a.etherman.BuildTrustedVerifyBatchesTxData(proof.BatchNumber-1, proof.BatchNumberFinal, &inputs, sender)
	if err != nil {
		log.Errorf("Error estimating batch verification to add to eth tx manager: %v", err)
		a.handleFailureToAddVerifyBatchToBeMonitored(ctx, proof)
		return false
	}

	monitoredTxID, err := a.ethTxManager.Add(ctx, to, nil, big.NewInt(0), data, a.cfg.GasOffset, nil)
	if err != nil {
		log.Errorf("Error Adding TX to ethTxManager: %v", err)
		mTxLogger := ethtxmanager.CreateLogger(monitoredTxID, sender, to)
		mTxLogger.Errorf("Error to add batch verification tx to eth tx manager: %v", err)
		a.handleFailureToAddVerifyBatchToBeMonitored(ctx, proof)
		return false
	}

	// process monitored batch verifications before starting a next cycle
	a.ethTxManager.ProcessPendingMonitoredTxs(ctx, func(result ethtxmanager.MonitoredTxResult) {
		a.handleMonitoredTxResult(result)
	})

	return true
}

func (a *Aggregator) handleFailureToAddVerifyBatchToBeMonitored(ctx context.Context, proof *state.Proof) {
	log := log.WithFields("proofId", proof.ProofID, "batches", fmt.Sprintf("%d-%d", proof.BatchNumber, proof.BatchNumberFinal))
	proof.GeneratingSince = nil
	err := a.state.UpdateGeneratedProof(ctx, proof, nil)
	if err != nil {
		log.Errorf("Failed updating proof state (false): %v", err)
	}
	a.endProofVerification()
}

// buildFinalProof builds and return the final proof for an aggregated/batch proof.
func (a *Aggregator) buildFinalProof(ctx context.Context, prover proverInterface, proof *state.Proof) (*prover.FinalProof, error) {
	log := log.WithFields(
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

	log.Infof("Final proof ID for batches [%d-%d]: %s", proof.BatchNumber, proof.BatchNumberFinal, *proof.ProofID)
	log = log.WithFields("finalProofId", finalProofID)

	finalProof, err := prover.WaitFinalProof(ctx, *proof.ProofID)
	if err != nil {
		return nil, fmt.Errorf("failed to get final proof from prover: %w", err)
	}

	// mock prover sanity check
	if string(finalProof.Public.NewStateRoot) == mockedStateRoot && string(finalProof.Public.NewLocalExitRoot) == mockedLocalExitRoot {
		// This local exit root and state root come from the mock
		// prover, use the one captured by the executor instead
		finalBatch, _, err := a.state.GetBatch(ctx, proof.BatchNumberFinal, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve batch with number [%d]", proof.BatchNumberFinal)
		}
		log.Warnf("NewLocalExitRoot and NewStateRoot look like a mock values, using values from executor instead: LER: %v, SR: %v",
			finalBatch.LocalExitRoot.TerminalString(), finalBatch.StateRoot.TerminalString())
		finalProof.Public.NewStateRoot = finalBatch.StateRoot.Bytes()
		finalProof.Public.NewLocalExitRoot = finalBatch.LocalExitRoot.Bytes()
	}

	// Sanity Check: state root from the proof must match the one from the final batch
	finalBatch, _, err := a.state.GetBatch(ctx, proof.BatchNumberFinal, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve batch with number [%d]", proof.BatchNumberFinal)
	}

	if !bytes.Equal(finalProof.Public.NewStateRoot, finalBatch.StateRoot.Bytes()) {
		for {
			log.Errorf("State root from the proof [%#x] does not match the one from the batch [%#x]. HALTED", finalProof.Public.NewStateRoot, finalBatch.StateRoot.Bytes())
			time.Sleep(a.cfg.RetryTime.Duration)
		}
	}

	return finalProof, nil
}

// tryBuildFinalProof checks if the provided proof is eligible to be used to
// build the final proof.  If no proof is provided it looks for a previously
// generated proof.  If the proof is eligible, then the final proof generation
// is triggered.
func (a *Aggregator) tryBuildFinalProof(ctx context.Context, prover proverInterface, proof *state.Proof) (bool, error) {
	proverName := prover.Name()
	proverID := prover.ID()

	log := log.WithFields(
		"prover", proverName,
		"proverId", proverID,
		"proverAddr", prover.Addr(),
	)
	log.Debug("tryBuildFinalProof start")

	var err error
	if !a.canVerifyProof() {
		log.Debug("Time to verify proof not reached or proof verification in progress")
		return false, nil
	}
	log.Debug("Send final proof time reached")

	lastVerifiedBatchNumber, err := a.etherman.GetLatestVerifiedBatchNum()
	if err != nil {
		return false, err
	}

	if proof == nil {
		// we don't have a proof generating at the moment, check if we
		// have a proof ready to verify

		proof, err = a.getAndLockProofReadyToVerify(ctx, prover, lastVerifiedBatchNumber)
		if errors.Is(err, state.ErrNotFound) {
			// nothing to verify, swallow the error
			log.Debug("No proof ready to verify")
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
					log.Errorf("Failed to unlock proof: %v", err2)
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

	log = log.WithFields(
		"proofId", *proof.ProofID,
		"batches", fmt.Sprintf("%d-%d", proof.BatchNumber, proof.BatchNumberFinal),
	)

	// at this point we have an eligible proof, build the final one using it
	finalProof, err := a.buildFinalProof(ctx, prover, proof)
	if err != nil {
		err = fmt.Errorf("failed to build final proof, %w", err)
		log.Error(FirstToUpper(err.Error()))
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

	log.Debug("tryBuildFinalProof end")
	return true, nil
}

func (a *Aggregator) validateEligibleFinalProof(ctx context.Context, proof *state.Proof, lastVerifiedBatchNum uint64) (bool, error) {
	batchNumberToVerify := lastVerifiedBatchNum + 1

	if proof.BatchNumber != batchNumberToVerify {
		if proof.BatchNumber < batchNumberToVerify && proof.BatchNumberFinal >= batchNumberToVerify {
			// We have a proof that contains some batches below the last batch verified, anyway can be eligible as final proof
			log.Warnf("Proof %d-%d contains some batches lower than last batch verified %d. Check anyway if it is eligible", proof.BatchNumber, proof.BatchNumberFinal, lastVerifiedBatchNum)
		} else if proof.BatchNumberFinal < batchNumberToVerify {
			// We have a proof that contains batches below that the last batch verified, we need to delete this proof
			log.Warnf("Proof %d-%d lower than next batch to verify %d. Deleting it", proof.BatchNumber, proof.BatchNumberFinal, batchNumberToVerify)
			err := a.state.DeleteGeneratedProofs(ctx, proof.BatchNumber, proof.BatchNumberFinal, nil)
			if err != nil {
				return false, fmt.Errorf("failed to delete discarded proof, err: %w", err)
			}
			return false, nil
		} else {
			log.Debugf("Proof batch number %d is not the following to last verfied batch number %d", proof.BatchNumber, lastVerifiedBatchNum)
			return false, nil
		}
	}

	bComplete, err := a.state.CheckProofContainsCompleteSequences(ctx, proof, nil)
	if err != nil {
		return false, fmt.Errorf("failed to check if proof contains complete sequences, %w", err)
	}
	if !bComplete {
		log.Infof("Recursive proof %d-%d not eligible to be verified: not containing complete sequences", proof.BatchNumber, proof.BatchNumberFinal)
		return false, nil
	}
	return true, nil
}

func (a *Aggregator) getAndLockProofReadyToVerify(ctx context.Context, prover proverInterface, lastVerifiedBatchNum uint64) (*state.Proof, error) {
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
		log.Warnf("Failed to begin transaction to release proof aggregation state, err: %v", err)
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
			log.Error(FirstToUpper(err.Error()))
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

func (a *Aggregator) getAndLockProofsToAggregate(ctx context.Context, prover proverInterface) (*state.Proof, *state.Proof, error) {
	log := log.WithFields(
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
		log.Errorf("Failed to begin transaction to set proof aggregation state, err: %v", err)
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
			log.Error(FirstToUpper(err.Error()))
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

func (a *Aggregator) tryAggregateProofs(ctx context.Context, prover proverInterface) (bool, error) {
	proverName := prover.Name()
	proverID := prover.ID()

	log := log.WithFields(
		"prover", proverName,
		"proverId", proverID,
		"proverAddr", prover.Addr(),
	)
	log.Debug("tryAggregateProofs start")

	proof1, proof2, err0 := a.getAndLockProofsToAggregate(ctx, prover)
	if errors.Is(err0, state.ErrNotFound) {
		// nothing to aggregate, swallow the error
		log.Debug("Nothing to aggregate")
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
				log.Errorf("Failed to release aggregated proofs, err: %v", err2)
			}
		}
		log.Debug("tryAggregateProofs end")
	}()

	log.Infof("Aggregating proofs: %d-%d and %d-%d", proof1.BatchNumber, proof1.BatchNumberFinal, proof2.BatchNumber, proof2.BatchNumberFinal)

	batches := fmt.Sprintf("%d-%d", proof1.BatchNumber, proof2.BatchNumberFinal)
	log = log.WithFields("batches", batches)

	inputProver := map[string]interface{}{
		"recursive_proof_1": proof1.Proof,
		"recursive_proof_2": proof2.Proof,
	}
	b, err := json.Marshal(inputProver)
	if err != nil {
		err = fmt.Errorf("failed to serialize input prover, %w", err)
		log.Error(FirstToUpper(err.Error()))
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
		log.Error(FirstToUpper(err.Error()))
		return false, err
	}

	proof.ProofID = aggrProofID

	log.Infof("Proof ID for aggregated proof: %v", *proof.ProofID)
	log = log.WithFields("proofId", *proof.ProofID)

	recursiveProof, err := prover.WaitRecursiveProof(ctx, *proof.ProofID)
	if err != nil {
		err = fmt.Errorf("failed to get aggregated proof from prover, %w", err)
		log.Error(FirstToUpper(err.Error()))
		return false, err
	}

	log.Info("Aggregated proof generated")

	proof.Proof = recursiveProof

	// update the state by removing the 2 aggregated proofs and storing the
	// newly generated recursive proof
	dbTx, err := a.state.BeginStateTransaction(ctx)
	if err != nil {
		err = fmt.Errorf("failed to begin transaction to update proof aggregation state, %w", err)
		log.Error(FirstToUpper(err.Error()))
		return false, err
	}

	err = a.state.DeleteGeneratedProofs(ctx, proof1.BatchNumber, proof2.BatchNumberFinal, dbTx)
	if err != nil {
		if err := dbTx.Rollback(ctx); err != nil {
			err := fmt.Errorf("failed to rollback proof aggregation state, %w", err)
			log.Error(FirstToUpper(err.Error()))
			return false, err
		}
		err = fmt.Errorf("failed to delete previously aggregated proofs, %w", err)
		log.Error(FirstToUpper(err.Error()))
		return false, err
	}

	now := time.Now().Round(time.Microsecond)
	proof.GeneratingSince = &now

	err = a.state.AddGeneratedProof(ctx, proof, dbTx)
	if err != nil {
		if err := dbTx.Rollback(ctx); err != nil {
			err := fmt.Errorf("failed to rollback proof aggregation state, %w", err)
			log.Error(FirstToUpper(err.Error()))
			return false, err
		}
		err = fmt.Errorf("failed to store the recursive proof, %w", err)
		log.Error(FirstToUpper(err.Error()))
		return false, err
	}

	err = dbTx.Commit(ctx)
	if err != nil {
		err = fmt.Errorf("failed to store the recursive proof, %w", err)
		log.Error(FirstToUpper(err.Error()))
		return false, err
	}

	// NOTE(pg): the defer func is useless from now on, use a different variable
	// name for errors (or shadow err in inner scopes) to not trigger it.

	// state is up to date, check if we can send the final proof using the
	// one just crafted.
	finalProofBuilt, finalProofErr := a.tryBuildFinalProof(ctx, prover, proof)
	if finalProofErr != nil {
		// just log the error and continue to handle the aggregated proof
		log.Errorf("Failed trying to check if recursive proof can be verified: %v", finalProofErr)
	}

	// NOTE(pg): prover is done, use a.ctx from now on

	if !finalProofBuilt {
		proof.GeneratingSince = nil

		// final proof has not been generated, update the recursive proof
		err := a.state.UpdateGeneratedProof(a.ctx, proof, nil)
		if err != nil {
			err = fmt.Errorf("failed to store batch proof result, %w", err)
			log.Error(FirstToUpper(err.Error()))
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

func (a *Aggregator) getAndLockBatchToProve(ctx context.Context, prover proverInterface) (*state.Batch, *state.Proof, error) {
	proverID := prover.ID()
	proverName := prover.Name()

	log := log.WithFields(
		"prover", proverName,
		"proverId", proverID,
		"proverAddr", prover.Addr(),
	)

	a.stateDBMutex.Lock()
	defer a.stateDBMutex.Unlock()

	// Get last virtual batch number from L1
	lastVerifiedBatchNumber, err := a.etherman.GetLatestVerifiedBatchNum()
	if err != nil {
		return nil, nil, err
	}

	proofExists := true
	batchNumberToVerify := lastVerifiedBatchNumber

	// Look for the batch number to verify
	for proofExists {
		batchNumberToVerify++
		proofExists, err = a.state.CheckProofExistsForBatch(ctx, batchNumberToVerify, nil)
		if err != nil {
			log.Infof("Error checking proof exists for batch %d", batchNumberToVerify)
			return nil, nil, err
		}
	}

	// Check if the batch has been sequenced
	sequence, err := a.l1Syncr.GetSequenceByBatchNumber(ctx, batchNumberToVerify)
	if err != nil && !errors.Is(err, entities.ErrNotFound) {
		return nil, nil, err
	}

	// Not found, so it it not possible to verify the batch yet
	if sequence == nil || errors.Is(err, entities.ErrNotFound) {
		log.Infof("No sequence found for batch %d", batchNumberToVerify)
		return nil, nil, state.ErrNotFound
	}

	stateSequence := state.Sequence{
		FromBatchNumber: sequence.FromBatchNumber,
		ToBatchNumber:   sequence.ToBatchNumber,
	}

	err = a.state.AddSequence(ctx, stateSequence, nil)
	if err != nil {
		log.Infof("Error storing sequence for batch %d", batchNumberToVerify)
		return nil, nil, err
	}

	batch, _, err := a.state.GetBatch(ctx, batchNumberToVerify, nil)
	if err != nil {
		return batch, nil, err
	}

	// All the data required to generate a proof is ready
	log.Infof("Found virtual batch %d pending to generate proof", batch.BatchNumber)
	log = log.WithFields("batch", batch.BatchNumber)

	log.Info("Checking profitability to aggregate batch")

	// pass pol collateral as zero here, bcs in smart contract fee for aggregator is not defined yet
	isProfitable, err := a.profitabilityChecker.IsProfitable(ctx, big.NewInt(0))
	if err != nil {
		log.Errorf("Failed to check aggregator profitability, err: %v", err)
		return nil, nil, err
	}

	if !isProfitable {
		log.Infof("Batch is not profitable, pol collateral %d", big.NewInt(0))
		return nil, nil, err
	}

	now := time.Now().Round(time.Microsecond)
	proof := &state.Proof{
		BatchNumber:      batch.BatchNumber,
		BatchNumberFinal: batch.BatchNumber,
		Prover:           &proverName,
		ProverID:         &proverID,
		GeneratingSince:  &now,
	}

	// Avoid other prover to process the same batch
	err = a.state.AddGeneratedProof(ctx, proof, nil)
	if err != nil {
		log.Errorf("Failed to add batch proof, err: %v", err)
		return nil, nil, err
	}

	return batch, proof, nil
}

func (a *Aggregator) tryGenerateBatchProof(ctx context.Context, prover proverInterface) (bool, error) {
	log := log.WithFields(
		"prover", prover.Name(),
		"proverId", prover.ID(),
		"proverAddr", prover.Addr(),
	)
	log.Debug("tryGenerateBatchProof start")

	batchToProve, proof, err0 := a.getAndLockBatchToProve(ctx, prover)
	if errors.Is(err0, state.ErrNotFound) {
		// nothing to proof, swallow the error
		log.Debug("Nothing to generate proof")
		return false, nil
	}
	if err0 != nil {
		return false, err0
	}

	log = log.WithFields("batch", batchToProve.BatchNumber)

	var (
		genProofID *string
		err        error
	)

	defer func() {
		if err != nil {
			log.Debug("Deleting proof in progress")
			err2 := a.state.DeleteGeneratedProofs(a.ctx, proof.BatchNumber, proof.BatchNumberFinal, nil)
			if err2 != nil {
				log.Errorf("Failed to delete proof in progress, err: %v", err2)
			}
		}
		log.Debug("tryGenerateBatchProof end")
	}()

	log.Infof("Sending zki + batch to the prover, batchNumber [%d]", batchToProve.BatchNumber)
	inputProver, err := a.buildInputProver(ctx, batchToProve)
	if err != nil {
		err = fmt.Errorf("failed to build input prover, %w", err)
		log.Error(FirstToUpper(err.Error()))
		return false, err
	}

	log.Infof("Sending a batch to the prover. OldAccInputHash [%#x], L1InfoRoot [%#x]",
		inputProver.PublicInputs.OldAccInputHash, inputProver.PublicInputs.L1InfoRoot)

	genProofID, err = prover.BatchProof(inputProver)
	if err != nil {
		err = fmt.Errorf("failed to get batch proof id, %w", err)
		log.Error(FirstToUpper(err.Error()))
		return false, err
	}

	proof.ProofID = genProofID

	log = log.WithFields("proofId", *proof.ProofID)

	resGetProof, err := prover.WaitRecursiveProof(ctx, *proof.ProofID)
	if err != nil {
		err = fmt.Errorf("failed to get proof from prover, %w", err)
		log.Error(FirstToUpper(err.Error()))
		return false, err
	}

	log.Info("Batch proof generated")

	proof.Proof = resGetProof

	// NOTE(pg): the defer func is useless from now on, use a different variable
	// name for errors (or shadow err in inner scopes) to not trigger it.

	finalProofBuilt, finalProofErr := a.tryBuildFinalProof(ctx, prover, proof)
	if finalProofErr != nil {
		// just log the error and continue to handle the generated proof
		log.Errorf("Error trying to build final proof: %v", finalProofErr)
	}

	// NOTE(pg): prover is done, use a.ctx from now on

	if !finalProofBuilt {
		proof.GeneratingSince = nil

		// final proof has not been generated, update the batch proof
		err := a.state.UpdateGeneratedProof(a.ctx, proof, nil)
		if err != nil {
			err = fmt.Errorf("failed to store batch proof result, %w", err)
			log.Error(FirstToUpper(err.Error()))
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

// startProofVerification sets to true the verifyingProof variable to indicate that there is a proof verification in progress
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

func (a *Aggregator) buildInputProver(ctx context.Context, batchToVerify *state.Batch) (*prover.StatelessInputProver, error) {
	isForcedBatch := false
	batchRawData := &state.BatchRawV2{}
	var err error

	if batchToVerify.BatchNumber == 1 || batchToVerify.ForcedBatchNum != nil {
		isForcedBatch = true
	} else {
		batchRawData, err = state.DecodeBatchV2(batchToVerify.BatchL2Data)
		if err != nil {
			log.Errorf("Failed to decode batch data, err: %v", err)
			return nil, err
		}
	}

	l1InfoTreeData := map[uint32]*prover.L1Data{}
	forcedBlockhashL1 := common.Hash{}
	l1InfoRoot := batchToVerify.L1InfoRoot.Bytes()
	if !isForcedBatch {
		tree, err := l1infotree.NewL1InfoTree(32, [][32]byte{}) // nolint:gomnd
		if err != nil {
			return nil, err
		}

		leaves, err := a.l1Syncr.GetLeafsByL1InfoRoot(ctx, batchToVerify.L1InfoRoot)
		if err != nil && !errors.Is(err, entities.ErrNotFound) {
			return nil, err
		}

		aLeaves := make([][32]byte, len(leaves))
		for i, leaf := range leaves {
			aLeaves[i] = l1infotree.HashLeafData(leaf.GlobalExitRoot, leaf.PreviousBlockHash, uint64(leaf.Timestamp.Unix()))
		}

		for _, l2blockRaw := range batchRawData.Blocks {
			_, contained := l1InfoTreeData[l2blockRaw.IndexL1InfoTree]
			if !contained && l2blockRaw.IndexL1InfoTree != 0 {
				leaves, err := a.l1Syncr.GetL1InfoTreeLeaves(ctx, []uint32{l2blockRaw.IndexL1InfoTree})
				if err != nil {
					log.Errorf("Error getting l1InfoTreeLeaf: %v", err)
					return nil, err
				}

				l1InfoTreeLeaf := leaves[l2blockRaw.IndexL1InfoTree]

				// Calculate smt proof
				log.Infof("Calling tree.ComputeMerkleProof")
				smtProof, calculatedL1InfoRoot, err := tree.ComputeMerkleProof(l2blockRaw.IndexL1InfoTree, aLeaves)
				if err != nil {
					log.Errorf("Error computing merkle proof: %v", err)
					return nil, err
				}

				if batchToVerify.L1InfoRoot != calculatedL1InfoRoot {
					return nil, fmt.Errorf("error: l1InfoRoot mismatch. L1InfoRoot: %s, calculatedL1InfoRoot: %s. l1InfoTreeIndex: %d", batchToVerify.L1InfoRoot.String(), calculatedL1InfoRoot.String(), l2blockRaw.IndexL1InfoTree)
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
				log.Errorf("Error getting virtual batch: %v", err)
				return nil, err
			}
			l1Block, err := a.l1Syncr.GetL1BlockByNumber(ctx, virtualBatch.BlockNumber)
			if err != nil {
				log.Errorf("Error getting l1 block: %v", err)
				return nil, err
			}

			forcedBlockhashL1 = l1Block.ParentHash
			l1InfoRoot = batchToVerify.GlobalExitRoot.Bytes()
		} /*else {
			forcedBlockhashL1, err = a.state.GetForcedBatchParentHash(ctx, *batchToVerify.ForcedBatchNum, nil)
			if err != nil {
				return nil, err
			}
		}*/
	}

	// Get Witness
	witness, err := getWitness(batchToVerify.BatchNumber, a.cfg.WitnessURL, a.cfg.UseFullWitness)
	if err != nil {
		log.Errorf("Failed to get witness, err: %v", err)
		return nil, err
	}

	// Get Old Acc Input Hash
	oldBatch, _, err := a.state.GetBatch(ctx, batchToVerify.BatchNumber-1, nil)
	if err != nil {
		return nil, err
	}

	inputProver := &prover.StatelessInputProver{
		PublicInputs: &prover.StatelessPublicInputs{
			Witness:           witness,
			OldAccInputHash:   oldBatch.AccInputHash.Bytes(),
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

	printInputProver(inputProver)
	return inputProver, nil
}

func getWitness(batchNumber uint64, URL string, fullWitness bool) ([]byte, error) {
	var witness string
	var response rpc.Response
	var err error

	witnessType := "trimmed"
	if fullWitness {
		witnessType = "full"
	}

	response, err = rpc.JSONRPCCall(URL, "zkevm_getBatchWitness", batchNumber, witnessType)
	if err != nil {
		return nil, err
	}

	// Check if the response is an error
	if response.Error != nil {
		return nil, fmt.Errorf("error from witness for batch %d: %v", batchNumber, response.Error)
	}

	err = json.Unmarshal(response.Result, &witness)
	if err != nil {
		return nil, err
	}

	witnessString := strings.TrimLeft(witness, "0x")
	if len(witnessString)%2 != 0 {
		witnessString = "0" + witnessString
	}
	bytes := common.Hex2Bytes(witnessString)

	return bytes, nil
}

func printInputProver(inputProver *prover.StatelessInputProver) {
	log.Debugf("Witness length: %v", len(inputProver.PublicInputs.Witness))
	log.Debugf("BatchL2Data length: %v", len(inputProver.PublicInputs.BatchL2Data))
	// log.Debugf("Full DataStream: %v", common.Bytes2Hex(inputProver.PublicInputs.DataStream))
	log.Debugf("OldAccInputHash: %v", common.BytesToHash(inputProver.PublicInputs.OldAccInputHash))
	log.Debugf("L1InfoRoot: %v", common.BytesToHash(inputProver.PublicInputs.L1InfoRoot))
	log.Debugf("TimestampLimit: %v", inputProver.PublicInputs.TimestampLimit)
	log.Debugf("SequencerAddr: %v", inputProver.PublicInputs.SequencerAddr)
	log.Debugf("AggregatorAddr: %v", inputProver.PublicInputs.AggregatorAddr)
	log.Debugf("L1InfoTreeData: %+v", inputProver.PublicInputs.L1InfoTreeData)
	log.Debugf("ForcedBlockhashL1: %v", common.BytesToHash(inputProver.PublicInputs.ForcedBlockhashL1))
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
func (hc *healthChecker) Check(ctx context.Context, req *grpchealth.HealthCheckRequest) (*grpchealth.HealthCheckResponse, error) {
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

func (a *Aggregator) handleMonitoredTxResult(result ethtxmanager.MonitoredTxResult) {
	mTxResultLogger := ethtxmanager.CreateMonitoredTxResultLogger(result)
	if result.Status == ethtxmanager.MonitoredTxStatusFailed {
		mTxResultLogger.Fatal("failed to send batch verification, TODO: review this fatal and define what to do in this case")
	}

	// TODO: REVIEW THIS

	/*
	   // monitoredIDFormat: "proof-from-%v-to-%v"
	   idSlice := strings.Split(result.ID, "-")
	   proofBatchNumberStr := idSlice[2]
	   proofBatchNumber, err := strconv.ParseUint(proofBatchNumberStr, encoding.Base10, 0)

	   	if err != nil {
	   		mTxResultLogger.Errorf("failed to read final proof batch number from monitored tx: %v", err)
	   	}

	   proofBatchNumberFinalStr := idSlice[4]
	   proofBatchNumberFinal, err := strconv.ParseUint(proofBatchNumberFinalStr, encoding.Base10, 0)

	   	if err != nil {
	   		mTxResultLogger.Errorf("failed to read final proof batch number final from monitored tx: %v", err)
	   	}

	   log := log.WithFields("txId", result.ID, "batches", fmt.Sprintf("%d-%d", proofBatchNumber, proofBatchNumberFinal))
	   log.Info("Final proof verified")

	   // wait for the synchronizer to catch up the verified batches
	   log.Debug("A final proof has been sent, waiting for the network to be synced")

	   	for !a.isSynced(a.ctx, &proofBatchNumberFinal) {
	   		log.Info("Waiting for synchronizer to sync...")
	   		time.Sleep(a.cfg.RetryTime.Duration)
	   	}

	   // network is synced with the final proof, we can safely delete all recursive
	   // proofs up to the last synced batch
	   err = a.State.CleanupGeneratedProofs(a.ctx, proofBatchNumberFinal, nil)

	   	if err != nil {
	   		log.Errorf("Failed to store proof aggregation result: %v", err)
	   	}
	*/
}

/*
func buildMonitoredTxID(batchNumber, batchNumberFinal uint64) string {
	return fmt.Sprintf(monitoredIDFormat, batchNumber, batchNumberFinal)
}
*/

func (a *Aggregator) cleanupLockedProofs() {
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-time.After(a.timeCleanupLockedProofs.Duration):
			n, err := a.state.CleanupLockedProofs(a.ctx, a.cfg.GeneratingProofCleanupThreshold, nil)
			if err != nil {
				log.Errorf("Failed to cleanup locked proofs: %v", err)
			}
			if n == 1 {
				log.Warn("Found a stale proof and removed from cache")
			} else if n > 1 {
				log.Warnf("Found %d stale proofs and removed from cache", n)
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
