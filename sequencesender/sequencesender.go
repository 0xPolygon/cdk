package sequencesender

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/rpc"
	"github.com/0xPolygon/cdk/rpc/types"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/state"
	"github.com/0xPolygon/zkevm-ethtx-manager/ethtxmanager"
	ethtxlog "github.com/0xPolygon/zkevm-ethtx-manager/log"
	ethtxtypes "github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

const ten = 10

// EthTxManager represents the eth tx manager interface
type EthTxManager interface {
	Start()
	AddWithGas(
		ctx context.Context,
		to *common.Address,
		value *big.Int,
		data []byte,
		gasOffset uint64,
		sidecar *ethtypes.BlobTxSidecar,
		gas uint64,
	) (common.Hash, error)
	Remove(ctx context.Context, hash common.Hash) error
	ResultsByStatus(ctx context.Context, status []ethtxtypes.MonitoredTxStatus) ([]ethtxtypes.MonitoredTxResult, error)
	Result(ctx context.Context, hash common.Hash) (ethtxtypes.MonitoredTxResult, error)
}

// Etherman represents the etherman behaviour
type Etherman interface {
	CurrentNonce(ctx context.Context, address common.Address) (uint64, error)
	GetLatestBlockHeader(ctx context.Context) (*ethtypes.Header, error)
	EstimateGas(ctx context.Context, from common.Address, to *common.Address, value *big.Int, data []byte) (uint64, error)
	GetLatestBatchNumber() (uint64, error)
}

// RPCInterface represents the RPC interface
type RPCInterface interface {
	GetBatch(batchNumber uint64) (*types.RPCBatch, error)
	GetWitness(batchNumber uint64, fullWitness bool) ([]byte, error)
}

// SequenceSender represents a sequence sender
type SequenceSender struct {
	cfg                      Config
	logger                   *log.Logger
	ethTxManager             EthTxManager
	etherman                 Etherman
	latestVirtualBatchNumber uint64                     // Latest virtualized batch obtained from L1
	latestVirtualTime        time.Time                  // Latest virtual batch timestamp
	latestSentToL1Batch      uint64                     // Latest batch sent to L1
	sequenceList             []uint64                   // Sequence of batch number to be send to L1
	sequenceData             map[uint64]*sequenceData   // All the batch data indexed by batch number
	mutexSequence            sync.Mutex                 // Mutex to access sequenceData and sequenceList
	ethTransactions          map[common.Hash]*ethTxData // All the eth tx sent to L1 indexed by hash
	ethTxData                map[common.Hash][]byte     // Tx data send to or received from L1
	mutexEthTx               sync.Mutex                 // Mutex to access ethTransactions
	sequencesTxFile          *os.File                   // Persistence of sent transactions
	validStream              bool                       // Not valid while receiving data before the desired batch
	seqSendingStopped        uint32                     // If there is a critical error
	TxBuilder                txbuilder.TxBuilder
	rpcClient                RPCInterface
}

type sequenceData struct {
	batchClosed bool
	batch       seqsendertypes.Batch
	batchRaw    *state.BatchRawV2
}

// New inits sequence sender
func New(cfg Config, logger *log.Logger,
	etherman *etherman.Client, txBuilder txbuilder.TxBuilder) (*SequenceSender, error) {
	// Create sequencesender
	s := SequenceSender{
		cfg:             cfg,
		logger:          logger,
		etherman:        etherman,
		ethTransactions: make(map[common.Hash]*ethTxData),
		ethTxData:       make(map[common.Hash][]byte),
		sequenceData:    make(map[uint64]*sequenceData),
		validStream:     false,
		TxBuilder:       txBuilder,
		rpcClient:       rpc.NewBatchEndpoints(cfg.RPCURL),
	}

	logger.Infof("TxBuilder configuration: %s", txBuilder.String())

	// Restore pending sent sequences
	err := s.loadSentSequencesTransactions()
	if err != nil {
		s.logger.Fatalf("error restoring sent sequences from file", err)
		return nil, err
	}

	// Create ethtxmanager client
	cfg.EthTxManager.Log = ethtxlog.Config{
		Environment: ethtxlog.LogEnvironment(cfg.Log.Environment),
		Level:       cfg.Log.Level,
		Outputs:     cfg.Log.Outputs,
	}

	s.ethTxManager, err = ethtxmanager.New(cfg.EthTxManager)
	if err != nil {
		s.logger.Fatalf("error creating ethtxmanager client: %v", err)
		return nil, err
	}

	return &s, nil
}

// Start starts the sequence sender
func (s *SequenceSender) Start(ctx context.Context) {
	// Start ethtxmanager client
	go s.ethTxManager.Start()

	// Get latest virtual state batch from L1
	err := s.updateLatestVirtualBatch()
	if err != nil {
		s.logger.Fatalf("error getting latest sequenced batch, error: %v", err)
	}

	// Sync all monitored sent L1 tx
	err = s.syncAllEthTxResults(ctx)
	if err != nil {
		s.logger.Fatalf("failed to sync monitored tx results, error: %v", err)
	}

	// Current batch to sequence
	atomic.StoreUint64(&s.latestSentToL1Batch, atomic.LoadUint64(&s.latestVirtualBatchNumber))

	// Start retrieving batches from RPC
	go func() {
		err := s.batchRetrieval(ctx)
		if err != nil {
			s.logFatalf("error retrieving batches from RPC: %v", err)
		}
	}()

	// Start sequence sending
	go s.sequenceSending(ctx)
}

// batchRetrieval keeps reading batches from the RPC
func (s *SequenceSender) batchRetrieval(ctx context.Context) error {
	ticker := time.NewTicker(s.cfg.GetBatchWaitInterval.Duration)
	defer ticker.Stop()

	currentBatchNumber := atomic.LoadUint64(&s.latestVirtualBatchNumber) + 1
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("context cancelled, stopping batch retrieval")
			return ctx.Err()
		default:
			// Try to retrieve batch from RPC
			rpcBatch, err := s.rpcClient.GetBatch(currentBatchNumber)
			if err != nil {
				if errors.Is(err, ethtxmanager.ErrNotFound) {
					s.logger.Infof("batch %d not found in RPC", currentBatchNumber)
				} else {
					s.logger.Errorf("error getting batch %d from RPC: %v", currentBatchNumber, err)
				}
				<-ticker.C
				continue
			}

			// Check if the batch is closed
			if !rpcBatch.IsClosed() {
				s.logger.Infof("batch %d is not closed yet", currentBatchNumber)
				<-ticker.C
				continue
			}

			// Process and decode the batch
			if err := s.populateSequenceData(rpcBatch, currentBatchNumber); err != nil {
				return err
			}

			// Increment the batch number for the next iteration
			currentBatchNumber++
		}
	}
}

func (s *SequenceSender) populateSequenceData(rpcBatch *types.RPCBatch, batchNumber uint64) error {
	s.mutexSequence.Lock()
	defer s.mutexSequence.Unlock()

	s.sequenceList = append(s.sequenceList, batchNumber)

	// Decode batch to retrieve the l1 info tree index
	batchRaw, err := state.DecodeBatchV2(rpcBatch.L2Data())
	if err != nil {
		s.logger.Errorf("Failed to decode batch data for batch %d, err: %v", batchNumber, err)
		return err
	}

	if len(batchRaw.Blocks) > 0 {
		rpcBatch.SetL1InfoTreeIndex(batchRaw.Blocks[len(batchRaw.Blocks)-1].IndexL1InfoTree)
	}

	s.sequenceData[batchNumber] = &sequenceData{
		batchClosed: rpcBatch.IsClosed(),
		batch:       rpcBatch,
		batchRaw:    batchRaw,
	}

	return nil
}

// sequenceSending starts loop to check if there are sequences to send and sends them if it's convenient
func (s *SequenceSender) sequenceSending(ctx context.Context) {
	// Create a ticker that fires every WaitPeriodSendSequence
	ticker := time.NewTicker(s.cfg.WaitPeriodSendSequence.Duration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("context canceled, stopping sequence sending")
			return

		case <-ticker.C:
			// Trigger the sequence sending when the ticker fires
			s.tryToSendSequence(ctx)
		}
	}
}

// purgeSequences purges batches from memory structures
func (s *SequenceSender) purgeSequences() {
	// If sequence sending is stopped, do not purge
	if s.IsStopped() {
		return
	}

	// Purge the information of batches that are already virtualized
	s.mutexSequence.Lock()
	defer s.mutexSequence.Unlock()
	truncateUntil := 0
	toPurge := make([]uint64, 0)
	for i, batchNumber := range s.sequenceList {
		if batchNumber <= atomic.LoadUint64(&s.latestVirtualBatchNumber) {
			truncateUntil = i + 1
			toPurge = append(toPurge, batchNumber)
		}
	}

	if len(toPurge) > 0 {
		s.sequenceList = s.sequenceList[truncateUntil:]

		firstPurged := toPurge[0]
		lastPurged := toPurge[len(toPurge)-1]
		for _, batchNum := range toPurge {
			delete(s.sequenceData, batchNum)
		}
		s.logger.Infof("batches purged count: %d, fromBatch: %d, toBatch: %d", len(toPurge), firstPurged, lastPurged)
	}
}

// tryToSendSequence checks if there is a sequence and it's worth it to send to L1
func (s *SequenceSender) tryToSendSequence(ctx context.Context) {
	// Update latest virtual batch
	s.logger.Infof("updating virtual batch")
	err := s.updateLatestVirtualBatch()
	if err != nil {
		return
	}

	// Check if the sequence sending is stopped
	if s.IsStopped() {
		s.logger.Warnf("sending is stopped!")
		return
	}

	// Update state of transactions
	s.logger.Infof("updating tx results")
	pendingTxsCount, err := s.syncEthTxResults(ctx)
	if err != nil {
		return
	}

	// Check if reached the maximum number of pending transactions
	if pendingTxsCount >= s.cfg.MaxPendingTx {
		s.logger.Infof("max number of pending txs (%d) reached. Waiting for some to be completed", pendingTxsCount)
		return
	}

	// Check if should send sequence to L1
	s.logger.Infof("getting sequences to send")
	sequence, err := s.getSequencesToSend(ctx)
	if err != nil || sequence == nil || sequence.Len() == 0 {
		if err != nil {
			s.logger.Errorf("error getting sequences: %v", err)
		}
		return
	}

	// Send sequences to L1
	firstBatch := sequence.FirstBatch()
	lastBatch := sequence.LastBatch()

	s.logger.Debugf(sequence.String())
	s.logger.Infof("sending sequences to L1. From batch %d to batch %d", firstBatch.BatchNumber(), lastBatch.BatchNumber())

	// Wait until last L1 block timestamp is L1BlockTimestampMargin seconds above the timestamp
	// of the last L2 block in the sequence
	timeMargin := int64(s.cfg.L1BlockTimestampMargin.Seconds())

	err = s.waitForMargin(ctx, lastBatch, timeMargin, "L1 block block timestamp",
		func() (uint64, error) {
			lastL1BlockHeader, err := s.etherman.GetLatestBlockHeader(ctx)
			if err != nil {
				return 0, err
			}

			return lastL1BlockHeader.Time, nil
		})
	if err != nil {
		s.logger.Errorf("error waiting for L1 block time margin: %v", err)
		return
	}

	// Sanity check: Wait until the current time is also L1BlockTimestampMargin seconds above the last L2 block timestamp
	err = s.waitForMargin(ctx, lastBatch, timeMargin, "current time",
		func() (uint64, error) { return uint64(time.Now().Unix()), nil })
	if err != nil {
		s.logger.Errorf("error waiting for current time margin: %v", err)
		return
	}

	// Send sequences to L1
	s.logger.Debugf(sequence.String())
	s.logger.Infof("sending sequences to L1. From batch %d to batch %d", firstBatch.BatchNumber(), lastBatch.BatchNumber())

	tx, err := s.TxBuilder.BuildSequenceBatchesTx(ctx, sequence)
	if err != nil {
		s.logger.Errorf("error building sequenceBatches tx: %v", err)
		return
	}

	// Get latest virtual state batch from L1
	err = s.updateLatestVirtualBatch()
	if err != nil {
		s.logger.Fatalf("error getting latest sequenced batch, error: %v", err)
	}

	sequence.SetLastVirtualBatchNumber(atomic.LoadUint64(&s.latestVirtualBatchNumber))

	gas, err := s.etherman.EstimateGas(ctx, s.cfg.SenderAddress, tx.To(), nil, tx.Data())
	if err != nil {
		s.logger.Errorf("error estimating gas: ", err)
		return
	}

	// Add sequence tx
	err = s.sendTx(ctx, false, nil, tx.To(), firstBatch.BatchNumber(), lastBatch.BatchNumber(), tx.Data(), gas)
	if err != nil {
		return
	}

	// Purge sequences data from memory
	s.purgeSequences()
}

// waitForMargin ensures that the time difference between the last L2 block and the current
// timestamp exceeds the time margin before proceeding. It checks immediately, and if not
// satisfied, it waits using a ticker and rechecks periodically.
//
// Params:
// - ctx: Context to handle cancellation.
// - lastBatch: The last batch in the sequence.
// - timeMargin: Required time difference in seconds.
// - description: A description for logging purposes.
// - getTimeFn: Function to get the current time (e.g., L1 block time or current time).
func (s *SequenceSender) waitForMargin(ctx context.Context, lastBatch seqsendertypes.Batch,
	timeMargin int64, description string, getTimeFn func() (uint64, error)) error {
	referentTime, err := getTimeFn()
	if err != nil {
		return err
	}

	lastL2BlockTimestamp := lastBatch.LastL2BLockTimestamp()
	elapsed, waitTime := marginTimeElapsed(lastL2BlockTimestamp, referentTime, timeMargin)
	if elapsed {
		s.logger.Infof("time difference for %s exceeds %d seconds, proceeding (batch number: %d, last l2 block ts: %d)",
			description, timeMargin, lastBatch.BatchNumber(), lastL2BlockTimestamp)
		return nil
	}

	s.logger.Infof("waiting %d seconds for %s, margin less than %d seconds (batch number: %d, last l2 block ts: %d)",
		waitTime, description, timeMargin, lastBatch.BatchNumber(), lastL2BlockTimestamp)
	ticker := time.NewTicker(time.Duration(waitTime) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Infof("context canceled during %s wait (batch number: %d, last l2 block ts: %d)",
				description, lastBatch.BatchNumber(), lastL2BlockTimestamp)
			return ctx.Err()

		case <-ticker.C:
			referentTime, err = getTimeFn()
			if err != nil {
				return err
			}

			elapsed, waitTime = marginTimeElapsed(lastL2BlockTimestamp, referentTime, timeMargin)
			if elapsed {
				s.logger.Infof("time margin for %s now exceeds %d seconds, proceeding (batch number: %d, last l2 block ts: %d)",
					description, timeMargin, lastBatch.BatchNumber(), lastL2BlockTimestamp)
				return nil
			}

			s.logger.Infof(
				"waiting another %d seconds for %s, margin still less than %d seconds (batch number: %d, last l2 block ts: %d)",
				waitTime, description, timeMargin, lastBatch.BatchNumber(), lastL2BlockTimestamp)
			ticker.Reset(time.Duration(waitTime) * time.Second)
		}
	}
}

func (s *SequenceSender) getSequencesToSend(ctx context.Context) (seqsendertypes.Sequence, error) {
	// Add sequences until too big for a single L1 tx or last batch is reached
	s.mutexSequence.Lock()
	defer s.mutexSequence.Unlock()
	var prevCoinbase common.Address
	sequenceBatches := make([]seqsendertypes.Batch, 0)
	for _, batchNumber := range s.sequenceList {
		if batchNumber <= atomic.LoadUint64(&s.latestVirtualBatchNumber) ||
			batchNumber <= atomic.LoadUint64(&s.latestSentToL1Batch) {
			continue
		}

		// Check if the next batch belongs to a new forkid, in this case we need to stop sequencing as we need to
		// wait the upgrade of forkid is completed and s.cfg.NumBatchForkIdUpgrade is disabled (=0) again
		if s.cfg.ForkUpgradeBatchNumber != 0 && batchNumber == (s.cfg.ForkUpgradeBatchNumber+1) {
			return nil, fmt.Errorf(
				"aborting sequencing process as we reached the batch %d where a new forkid is applied (upgrade)",
				s.cfg.ForkUpgradeBatchNumber+1,
			)
		}

		// New potential batch to add to the sequence
		batch := s.sequenceData[batchNumber].batch.DeepCopy()

		// If the coinbase changes, the sequence ends here
		if len(sequenceBatches) > 0 && batch.LastCoinbase() != prevCoinbase {
			s.logger.Infof(
				"batch with different coinbase (batch %v, sequence %v), sequence will be sent to this point",
				prevCoinbase, batch.LastCoinbase,
			)
			return s.TxBuilder.NewSequence(ctx, sequenceBatches, s.cfg.L2Coinbase)
		}
		prevCoinbase = batch.LastCoinbase()

		// Add new sequence batch
		sequenceBatches = append(sequenceBatches, batch)

		newSeq, err := s.TxBuilder.NewSequenceIfWorthToSend(ctx, sequenceBatches, s.cfg.L2Coinbase, batchNumber)
		if err != nil {
			return nil, err
		}
		if newSeq != nil {
			return newSeq, nil
		}

		// Check if the current batch is the last before a change to a new forkid
		// In this case we need to close and send the sequence to L1
		if s.cfg.ForkUpgradeBatchNumber != 0 && batchNumber == s.cfg.ForkUpgradeBatchNumber {
			s.logger.Infof("sequence should be sent to L1, as we have reached the batch %d "+
				"from which a new forkid is applied (upgrade)",
				s.cfg.ForkUpgradeBatchNumber,
			)
			return s.TxBuilder.NewSequence(ctx, sequenceBatches, s.cfg.L2Coinbase)
		}
	}

	// Reached the latest batch. Decide if it's worth to send the sequence, or wait for new batches
	if len(sequenceBatches) == 0 {
		s.logger.Infof("no batches to be sequenced")
		return nil, nil
	}

	if s.latestVirtualTime.Before(time.Now().Add(-s.cfg.LastBatchVirtualizationTimeMaxWaitPeriod.Duration)) {
		s.logger.Infof("sequence should be sent, too much time without sending anything to L1")
		return s.TxBuilder.NewSequence(ctx, sequenceBatches, s.cfg.L2Coinbase)
	}

	s.logger.Infof("not enough time has passed since last batch was virtualized and the sequence could be bigger")
	return nil, nil
}

// updateLatestVirtualBatch queries the value in L1 and updates the latest virtual batch field
func (s *SequenceSender) updateLatestVirtualBatch() error {
	// Get latest virtual state batch from L1
	latestVirtualBatchNumber, err := s.etherman.GetLatestBatchNumber()
	if err != nil {
		s.logger.Errorf("error getting latest virtual batch, error: %v", err)
		return errors.New("fail to get latest virtual batch")
	}

	atomic.StoreUint64(&s.latestVirtualBatchNumber, latestVirtualBatchNumber)
	s.logger.Infof("latest virtual batch is %d", latestVirtualBatchNumber)

	return nil
}

// logFatalf logs error, activates flag to stop sequencing, and remains in an infinite loop
func (s *SequenceSender) logFatalf(template string, args ...interface{}) {
	atomic.StoreUint32(&s.seqSendingStopped, 1)
	for {
		s.logger.Errorf(template, args...)
		s.logger.Errorf("sequence sending stopped.")
		time.Sleep(ten * time.Second)
	}
}

// marginTimeElapsed checks if the time between currentTime and l2BlockTimestamp is greater than timeMargin.
// If it's greater returns true, otherwise it returns false and the waitTime needed to achieve this timeMargin
func marginTimeElapsed(l2BlockTimestamp uint64, currentTime uint64, timeMargin int64) (bool, int64) {
	if int64(l2BlockTimestamp)-timeMargin > int64(currentTime) {
		return true, 0
	}

	timeDiff := int64(currentTime) - int64(l2BlockTimestamp)

	// If the difference is less than the required margin, return false and calculate the remaining wait time
	if timeDiff < timeMargin {
		// Calculate the wait time needed to reach the timeMargin
		waitTime := timeMargin - timeDiff
		return false, waitTime
	}

	// Time difference is greater than or equal to timeMargin, no need to wait
	return true, 0
}
