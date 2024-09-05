package sequencesender

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/state"
	"github.com/0xPolygonHermez/zkevm-ethtx-manager/ethtxmanager"
	ethtxlog "github.com/0xPolygonHermez/zkevm-ethtx-manager/log"
	"github.com/ethereum/go-ethereum/common"
)

// SequenceSender represents a sequence sender
type SequenceSender struct {
	cfg                      Config
	ethTxManager             *ethtxmanager.Client
	etherman                 *etherman.Client
	currentNonce             uint64
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
	seqSendingStopped        bool                       // If there is a critical error
	TxBuilder                txbuilder.TxBuilder
	latestVirtualBatchLock   sync.Mutex
}

type sequenceData struct {
	batchClosed bool
	batch       seqsendertypes.Batch
	batchRaw    *state.BatchRawV2
}

// New inits sequence sender
func New(cfg Config, etherman *etherman.Client, txBuilder txbuilder.TxBuilder) (*SequenceSender, error) {
	// Create sequencesender
	s := SequenceSender{
		cfg:               cfg,
		etherman:          etherman,
		ethTransactions:   make(map[common.Hash]*ethTxData),
		ethTxData:         make(map[common.Hash][]byte),
		sequenceData:      make(map[uint64]*sequenceData),
		validStream:       false,
		seqSendingStopped: false,
		TxBuilder:         txBuilder,
	}
	log.Infof("Seq_sender: %s", txBuilder.String())

	// Restore pending sent sequences
	err := s.loadSentSequencesTransactions()
	if err != nil {
		log.Fatalf("error restoring sent sequences from file", err)
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
		log.Fatalf("error creating ethtxmanager client: %v", err)
		return nil, err
	}

	return &s, nil
}

// Start starts the sequence sender
func (s *SequenceSender) Start(ctx context.Context) {
	// Start ethtxmanager client
	go s.ethTxManager.Start()

	// Get current nonce
	var err error
	s.currentNonce, err = s.etherman.CurrentNonce(ctx, s.cfg.L2Coinbase)
	if err != nil {
		log.Fatalf("failed to get current nonce from %v, error: %v", s.cfg.L2Coinbase, err)
	} else {
		log.Infof("current nonce for %v is %d", s.cfg.L2Coinbase, s.currentNonce)
	}

	// Get latest virtual state batch from L1
	err = s.getLatestVirtualBatch()
	if err != nil {
		log.Fatalf("error getting latest sequenced batch, error: %v", err)
	}

	// Sync all monitored sent L1 tx
	err = s.syncAllEthTxResults(ctx)
	if err != nil {
		log.Fatalf("failed to sync monitored tx results, error: %v", err)
	}

	// Current batch to sequence
	s.latestSentToL1Batch = s.latestVirtualBatchNumber

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
	currentBatchNumber := s.latestVirtualBatchNumber + 1
	for {
		rpcBatch, err := s.getBatchFromRPC(currentBatchNumber)
		if err != nil {
			if err == state.ErrNotFound {
				log.Infof("batch %d not found in RPC", currentBatchNumber)
			} else {
				log.Errorf("error getting batch %d from RPC: %v", currentBatchNumber, err)
			}
			time.Sleep(s.cfg.GetBatchWaitInterval.Duration)
			continue
		}

		// Check if the batch is closed
		if !rpcBatch.IsClosed() {
			log.Infof("batch %d is not closed yet", currentBatchNumber)
			time.Sleep(s.cfg.GetBatchWaitInterval.Duration)
			continue
		}

		// Create new batch
		s.mutexSequence.Lock()
		s.sequenceList = append(s.sequenceList, currentBatchNumber)

		// Decode batch to retrieve the l1 info tree index
		batchRaw, err := state.DecodeBatchV2(rpcBatch.L2Data())
		if err != nil {
			log.Errorf("Failed to decode batch data, err: %v", err)
			return err
		}

		if len(batchRaw.Blocks) > 0 {
			rpcBatch.SetL1InfoTreeIndex(batchRaw.Blocks[len(batchRaw.Blocks)-1].IndexL1InfoTree)
		}

		data := &sequenceData{
			batchClosed: rpcBatch.IsClosed(),
			batch:       rpcBatch,
			batchRaw:    batchRaw,
		}

		s.sequenceData[currentBatchNumber] = data
		s.mutexSequence.Unlock()

		currentBatchNumber++
	}
}

// sequenceSending starts loop to check if there are sequences to send and sends them if it's convenient
func (s *SequenceSender) sequenceSending(ctx context.Context) {
	for {
		s.tryToSendSequence(ctx)
		time.Sleep(s.cfg.WaitPeriodSendSequence.Duration)
	}
}

// purgeSequences purges batches from memory structures
func (s *SequenceSender) purgeSequences() {
	// If sequence sending is stopped, do not purge
	if s.seqSendingStopped {
		return
	}

	// Purge the information of batches that are already virtualized
	s.mutexSequence.Lock()
	truncateUntil := 0
	toPurge := make([]uint64, 0)
	for i := 0; i < len(s.sequenceList); i++ {
		batchNumber := s.sequenceList[i]
		if batchNumber <= s.latestVirtualBatchNumber {
			truncateUntil = i + 1
			toPurge = append(toPurge, batchNumber)
		}
	}

	if len(toPurge) > 0 {
		s.sequenceList = s.sequenceList[truncateUntil:]

		var firstPurged uint64
		var lastPurged uint64
		for i := 0; i < len(toPurge); i++ {
			if i == 0 {
				firstPurged = toPurge[i]
			}
			if i == len(toPurge)-1 {
				lastPurged = toPurge[i]
			}
			delete(s.sequenceData, toPurge[i])
		}
		log.Infof("batches purged count: %d, fromBatch: %d, toBatch: %d", len(toPurge), firstPurged, lastPurged)
	}
	s.mutexSequence.Unlock()
}

// tryToSendSequence checks if there is a sequence and it's worth it to send to L1
func (s *SequenceSender) tryToSendSequence(ctx context.Context) {
	// Update latest virtual batch
	log.Infof("updating virtual batch")
	err := s.getLatestVirtualBatch()
	if err != nil {
		return
	}

	// Update state of transactions
	log.Infof("updating tx results")
	countPending, err := s.syncEthTxResults(ctx)
	if err != nil {
		return
	}

	// Check if the sequence sending is stopped
	if s.seqSendingStopped {
		log.Warnf("sending is stopped!")
		return
	}

	// Check if reached the maximum number of pending transactions
	if countPending >= s.cfg.MaxPendingTx {
		log.Infof("max number of pending txs (%d) reached. Waiting for some to be completed", countPending)
		return
	}

	// Check if should send sequence to L1
	log.Infof("getting sequences to send")
	sequence, err := s.getSequencesToSend(ctx)
	if err != nil || sequence == nil || sequence.Len() == 0 {
		if err != nil {
			log.Errorf("error getting sequences: %v", err)
		}
		return
	}

	// Send sequences to L1
	firstBatch := sequence.FirstBatch()
	lastBatch := sequence.LastBatch()
	lastL2BlockTimestamp := lastBatch.LastL2BLockTimestamp()

	log.Debugf(sequence.String())
	log.Infof("sending sequences to L1. From batch %d to batch %d", firstBatch.BatchNumber(), lastBatch.BatchNumber())

	// Wait until last L1 block timestamp is L1BlockTimestampMargin seconds above the timestamp of the last L2 block in the sequence
	timeMargin := int64(s.cfg.L1BlockTimestampMargin.Seconds())
	for {
		// Get header of the last L1 block
		lastL1BlockHeader, err := s.etherman.GetLatestBlockHeader(ctx)
		if err != nil {
			log.Errorf("failed to get last L1 block timestamp, err: %v", err)
			return
		}

		elapsed, waitTime := s.marginTimeElapsed(lastL2BlockTimestamp, lastL1BlockHeader.Time, timeMargin)

		if !elapsed {
			log.Infof("waiting at least %d seconds to send sequences, time difference between last L1 block %d (ts: %d) and last L2 block %d (ts: %d) in the sequence is lower than %d seconds",
				waitTime, lastL1BlockHeader.Number, lastL1BlockHeader.Time, lastBatch.BatchNumber(), lastL2BlockTimestamp, timeMargin)
			time.Sleep(time.Duration(waitTime) * time.Second)
		} else {
			log.Infof("continuing, time difference between last L1 block %d (ts: %d) and last L2 block %d (ts: %d) in the sequence is greater than %d seconds",
				lastL1BlockHeader.Number, lastL1BlockHeader.Time, lastBatch.BatchNumber, lastL2BlockTimestamp, timeMargin)
			break
		}
	}

	// Sanity check: Wait also until current time is L1BlockTimestampMargin seconds above the timestamp of the last L2 block in the sequence
	for {
		currentTime := uint64(time.Now().Unix())

		elapsed, waitTime := s.marginTimeElapsed(lastL2BlockTimestamp, currentTime, timeMargin)

		// Wait if the time difference is less than L1BlockTimestampMargin
		if !elapsed {
			log.Infof("waiting at least %d seconds to send sequences, time difference between now (ts: %d) and last L2 block %d (ts: %d) in the sequence is lower than %d seconds",
				waitTime, currentTime, lastBatch.BatchNumber, lastL2BlockTimestamp, timeMargin)
			time.Sleep(time.Duration(waitTime) * time.Second)
		} else {
			log.Infof("[SeqSender]sending sequences now, time difference between now (ts: %d) and last L2 block %d (ts: %d) in the sequence is also greater than %d seconds",
				currentTime, lastBatch.BatchNumber, lastL2BlockTimestamp, timeMargin)
			break
		}
	}

	// Send sequences to L1
	log.Debugf(sequence.String())
	log.Infof("sending sequences to L1. From batch %d to batch %d", firstBatch.BatchNumber(), lastBatch.BatchNumber())

	tx, err := s.TxBuilder.BuildSequenceBatchesTx(ctx, sequence)
	if err != nil {
		log.Errorf("error building sequenceBatches tx: %v", err)
		return
	}

	// Get latest virtual state batch from L1
	err = s.getLatestVirtualBatch()
	if err != nil {
		log.Fatalf("error getting latest sequenced batch, error: %v", err)
	}

	sequence.SetLastVirtualBatchNumber(s.latestVirtualBatchNumber)

	txToEstimateGas, err := s.TxBuilder.BuildSequenceBatchesTx(ctx, sequence)
	if err != nil {
		log.Errorf("error building sequenceBatches tx to estimate gas: %v", err)
		return
	}

	gas, err := s.etherman.EstimateGas(ctx, s.cfg.SenderAddress, tx.To(), nil, txToEstimateGas.Data())
	if err != nil {
		log.Errorf("error estimating gas: ", err)
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

func (s *SequenceSender) getSequencesToSend(ctx context.Context) (seqsendertypes.Sequence, error) {
	// Add sequences until too big for a single L1 tx or last batch is reached
	s.mutexSequence.Lock()
	defer s.mutexSequence.Unlock()
	var prevCoinbase common.Address
	sequenceBatches := make([]seqsendertypes.Batch, 0)
	for i := 0; i < len(s.sequenceList); i++ {
		batchNumber := s.sequenceList[i]
		if batchNumber <= s.latestVirtualBatchNumber || batchNumber <= s.latestSentToL1Batch {
			continue
		}

		// Check if the next batch belongs to a new forkid, in this case we need to stop sequencing as we need to
		// wait the upgrade of forkid is completed and s.cfg.NumBatchForkIdUpgrade is disabled (=0) again
		if (s.cfg.ForkUpgradeBatchNumber != 0) && (batchNumber == (s.cfg.ForkUpgradeBatchNumber + 1)) {
			return nil, fmt.Errorf("aborting sequencing process as we reached the batch %d where a new forkid is applied (upgrade)", s.cfg.ForkUpgradeBatchNumber+1)
		}

		// New potential batch to add to the sequence
		batch := s.sequenceData[batchNumber].batch.DeepCopy()

		// If the coinbase changes, the sequence ends here
		if len(sequenceBatches) > 0 && batch.LastCoinbase() != prevCoinbase {
			log.Infof("batch with different coinbase (batch %v, sequence %v), sequence will be sent to this point", prevCoinbase, batch.LastCoinbase)
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

		// Check if the current batch is the last before a change to a new forkid, in this case we need to close and send the sequence to L1
		if (s.cfg.ForkUpgradeBatchNumber != 0) && (batchNumber == (s.cfg.ForkUpgradeBatchNumber)) {
			log.Infof("sequence should be sent to L1, as we have reached the batch %d from which a new forkid is applied (upgrade)", s.cfg.ForkUpgradeBatchNumber)
			return s.TxBuilder.NewSequence(ctx, sequenceBatches, s.cfg.L2Coinbase)
		}
	}

	// Reached the latest batch. Decide if it's worth to send the sequence, or wait for new batches
	if len(sequenceBatches) == 0 {
		log.Infof("no batches to be sequenced")
		return nil, nil
	}

	if s.latestVirtualTime.Before(time.Now().Add(-s.cfg.LastBatchVirtualizationTimeMaxWaitPeriod.Duration)) {
		log.Infof("sequence should be sent, too much time without sending anything to L1")
		return s.TxBuilder.NewSequence(ctx, sequenceBatches, s.cfg.L2Coinbase)
	}

	log.Infof("not enough time has passed since last batch was virtualized and the sequence could be bigger")
	return nil, nil
}

// getLatestVirtualBatch queries the value in L1 and updates the latest virtual batch field
func (s *SequenceSender) getLatestVirtualBatch() error {
	s.latestVirtualBatchLock.Lock()
	defer s.latestVirtualBatchLock.Unlock()

	// Get latest virtual state batch from L1
	var err error

	s.latestVirtualBatchNumber, err = s.etherman.GetLatestBatchNumber()
	if err != nil {
		log.Errorf("error getting latest virtual batch, error: %v", err)
		return errors.New("fail to get latest virtual batch")
	} else {
		log.Infof("latest virtual batch is %d", s.latestVirtualBatchNumber)
	}
	return nil
}

// marginTimeElapsed checks if the time between currentTime and l2BlockTimestamp is greater than timeMargin.
// If it's greater returns true, otherwise it returns false and the waitTime needed to achieve this timeMargin
func (s *SequenceSender) marginTimeElapsed(l2BlockTimestamp uint64, currentTime uint64, timeMargin int64) (bool, int64) {
	// Check the time difference between L2 block and currentTime
	var timeDiff int64
	if l2BlockTimestamp >= currentTime {
		//L2 block timestamp is above currentTime, negative timeDiff. We do in this way to avoid uint64 overflow
		timeDiff = int64(-(l2BlockTimestamp - currentTime))
	} else {
		timeDiff = int64(currentTime - l2BlockTimestamp)
	}

	// Check if the time difference is less than timeMargin (L1BlockTimestampMargin)
	if timeDiff < timeMargin {
		var waitTime int64
		if timeDiff < 0 { //L2 block timestamp is above currentTime
			waitTime = timeMargin + (-timeDiff)
		} else {
			waitTime = timeMargin - timeDiff
		}
		return false, waitTime
	} else { // timeDiff is greater than timeMargin
		return true, 0
	}
}

// logFatalf logs error, activates flag to stop sequencing, and remains in an infinite loop
func (s *SequenceSender) logFatalf(template string, args ...interface{}) {
	s.seqSendingStopped = true
	for {
		log.Errorf(template, args...)
		log.Errorf("sequence sending stopped.")
		time.Sleep(10 * time.Second)
	}
}
