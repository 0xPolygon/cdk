package sequencesender

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/state"
	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/0xPolygonHermez/zkevm-data-streamer/datastreamer"
	"github.com/0xPolygonHermez/zkevm-ethtx-manager/ethtxmanager"
	ethtxlog "github.com/0xPolygonHermez/zkevm-ethtx-manager/log"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/proto"
)

// SequenceSender represents a sequence sender
type SequenceSender struct {
	cfg                 Config
	ethTxManager        *ethtxmanager.Client
	etherman            *etherman.Client
	currentNonce        uint64
	latestVirtualBatch  uint64                     // Latest virtualized batch obtained from L1
	latestVirtualTime   time.Time                  // Latest virtual batch timestamp
	latestSentToL1Batch uint64                     // Latest batch sent to L1
	wipBatch            uint64                     // Work in progress batch
	sequenceList        []uint64                   // Sequence of batch number to be send to L1
	sequenceData        map[uint64]*sequenceData   // All the batch data indexed by batch number
	mutexSequence       sync.Mutex                 // Mutex to access sequenceData and sequenceList
	ethTransactions     map[common.Hash]*ethTxData // All the eth tx sent to L1 indexed by hash
	ethTxData           map[common.Hash][]byte     // Tx data send to or received from L1
	mutexEthTx          sync.Mutex                 // Mutex to access ethTransactions
	sequencesTxFile     *os.File                   // Persistence of sent transactions
	validStream         bool                       // Not valid while receiving data before the desired batch
	fromStreamBatch     uint64                     // Initial batch to connect to the streaming
	latestStreamBatch   uint64                     // Latest batch received by the streaming
	seqSendingStopped   bool                       // If there is a critical error
	prevStreamEntry     *datastreamer.FileEntry
	streamClient        *datastreamer.StreamClient
	TxBuilder           txbuilder.TxBuilder
}

type sequenceData struct {
	batchClosed bool
	batch       seqsendertypes.Batch
	batchRaw    *state.BatchRawV2
	batchType   datastream.BatchType
}

type ethTxData struct {
	Nonce           uint64                              `json:"nonce"`
	Status          string                              `json:"status"`
	SentL1Timestamp time.Time                           `json:"sentL1Timestamp"`
	StatusTimestamp time.Time                           `json:"statusTimestamp"`
	FromBatch       uint64                              `json:"fromBatch"`
	ToBatch         uint64                              `json:"toBatch"`
	MinedAtBlock    big.Int                             `json:"minedAtBlock"`
	OnMonitor       bool                                `json:"onMonitor"`
	To              common.Address                      `json:"to"`
	StateHistory    []string                            `json:"stateHistory"`
	Txs             map[common.Hash]ethTxAdditionalData `json:"txs"`
}

type ethTxAdditionalData struct {
	GasPrice      *big.Int `json:"gasPrice,omitempty"`
	RevertMessage string   `json:"revertMessage,omitempty"`
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
		latestStreamBatch: 0,
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

	// Create datastream client
	s.streamClient, err = datastreamer.NewClient(s.cfg.StreamClient.Server, 1)
	if err != nil {
		log.Fatalf("failed to create stream client, error: %v", err)
	} else {
		log.Infof("new stream client")
	}
	// Set func to handle the streaming
	s.streamClient.SetProcessEntryFunc(s.handleReceivedDataStream)

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
	err = s.updateLatestVirtualBatch()
	if err != nil {
		log.Fatalf("error getting latest sequenced batch, error: %v", err)
	}

	// Sync all monitored sent L1 tx
	err = s.syncAllEthTxResults(ctx)
	if err != nil {
		log.Fatalf("failed to sync monitored tx results, error: %v", err)
	}

	// Start datastream client
	err = s.streamClient.Start()
	if err != nil {
		log.Fatalf("failed to start stream client, error: %v", err)
	}

	// Set starting point of the streaming
	s.fromStreamBatch = s.latestVirtualBatch

	bookmark := &datastream.BookMark{
		Type:  datastream.BookmarkType_BOOKMARK_TYPE_BATCH,
		Value: s.fromStreamBatch,
	}

	marshalledBookmark, err := proto.Marshal(bookmark)
	if err != nil {
		log.Fatalf("failed to marshal bookmark, error: %v", err)
	}

	log.Infof("stream client from bookmark %v", bookmark)

	// Current batch to sequence
	s.wipBatch = s.latestVirtualBatch + 1
	s.latestSentToL1Batch = s.latestVirtualBatch

	// Start sequence sending
	go s.sequenceSending(ctx)

	// Start receiving the streaming
	err = s.streamClient.ExecCommandStartBookmark(marshalledBookmark)
	if err != nil {
		log.Fatalf("failed to connect to the streaming: %v", err)
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
		if batchNumber <= s.latestVirtualBatch {
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

// purgeEthTx purges transactions from memory structures
func (s *SequenceSender) purgeEthTx(ctx context.Context) {
	// If sequence sending is stopped, do not purge
	if s.seqSendingStopped {
		return
	}

	// Purge old transactions that are finalized
	s.mutexEthTx.Lock()
	timePurge := time.Now().Add(-s.cfg.WaitPeriodPurgeTxFile.Duration)
	toPurge := make([]common.Hash, 0)
	for hash, data := range s.ethTransactions {
		if !data.StatusTimestamp.Before(timePurge) {
			continue
		}

		if !data.OnMonitor || data.Status == ethtxmanager.MonitoredTxStatusFinalized.String() {
			toPurge = append(toPurge, hash)

			// Remove from tx monitor
			if data.OnMonitor {
				err := s.ethTxManager.Remove(ctx, hash)
				if err != nil {
					log.Warnf("error removing monitor tx %v from ethtxmanager: %v", hash, err)
				} else {
					log.Infof("removed monitor tx %v from ethtxmanager", hash)
				}
			}
		}
	}

	if len(toPurge) > 0 {
		var firstPurged uint64 = math.MaxUint64
		var lastPurged uint64
		for i := 0; i < len(toPurge); i++ {
			if s.ethTransactions[toPurge[i]].Nonce < firstPurged {
				firstPurged = s.ethTransactions[toPurge[i]].Nonce
			}
			if s.ethTransactions[toPurge[i]].Nonce > lastPurged {
				lastPurged = s.ethTransactions[toPurge[i]].Nonce
			}
			delete(s.ethTransactions, toPurge[i])
			delete(s.ethTxData, toPurge[i])
		}
		log.Infof("txs purged count: %d, fromNonce: %d, toNonce: %d", len(toPurge), firstPurged, lastPurged)
	}
	s.mutexEthTx.Unlock()
}

// syncEthTxResults syncs results from L1 for transactions in the memory structure
func (s *SequenceSender) syncEthTxResults(ctx context.Context) (uint64, error) {
	s.mutexEthTx.Lock()
	var txPending uint64
	var txSync uint64
	for hash, data := range s.ethTransactions {
		if data.Status == ethtxmanager.MonitoredTxStatusFinalized.String() {
			continue
		}

		_ = s.getResultAndUpdateEthTx(ctx, hash)
		txSync++
		txStatus := s.ethTransactions[hash].Status
		// Count if it is not in a final state
		if s.ethTransactions[hash].OnMonitor &&
			txStatus != ethtxmanager.MonitoredTxStatusFailed.String() &&
			txStatus != ethtxmanager.MonitoredTxStatusSafe.String() &&
			txStatus != ethtxmanager.MonitoredTxStatusFinalized.String() {
			txPending++
		}
	}
	s.mutexEthTx.Unlock()

	// Save updated sequences transactions
	err := s.saveSentSequencesTransactions(ctx)
	if err != nil {
		log.Errorf("error saving tx sequence, error: %v", err)
	}

	log.Infof("%d tx results synchronized (%d in pending state)", txSync, txPending)
	return txPending, nil
}

// syncAllEthTxResults syncs all tx results from L1
func (s *SequenceSender) syncAllEthTxResults(ctx context.Context) error {
	// Get all results
	results, err := s.ethTxManager.ResultsByStatus(ctx, nil)
	if err != nil {
		log.Warnf("error getting results for all tx: %v", err)
		return err
	}

	// Check and update tx status
	numResults := len(results)
	s.mutexEthTx.Lock()
	for _, result := range results {
		txSequence, exists := s.ethTransactions[result.ID]
		if !exists {
			log.Infof("transaction %v missing in memory structure. Adding it", result.ID)
			// No info: from/to batch and the sent timestamp
			s.ethTransactions[result.ID] = &ethTxData{
				SentL1Timestamp: time.Time{},
				StatusTimestamp: time.Now(),
				OnMonitor:       true,
				Status:          "*missing",
			}
			txSequence = s.ethTransactions[result.ID]
		}

		s.updateEthTxResult(txSequence, result)
	}
	s.mutexEthTx.Unlock()

	// Save updated sequences transactions
	err = s.saveSentSequencesTransactions(ctx)
	if err != nil {
		log.Errorf("error saving tx sequence, error: %v", err)
	}

	log.Infof("%d tx results synchronized", numResults)
	return nil
}

// copyTxData copies tx data in the internal structure
func (s *SequenceSender) copyTxData(txHash common.Hash, txData []byte, txsResults map[common.Hash]ethtxmanager.TxResult) {
	s.ethTxData[txHash] = make([]byte, len(txData))
	copy(s.ethTxData[txHash], txData)

	s.ethTransactions[txHash].Txs = make(map[common.Hash]ethTxAdditionalData, 0)
	for hash, result := range txsResults {
		var gasPrice *big.Int
		if result.Tx != nil {
			gasPrice = result.Tx.GasPrice()
		}

		add := ethTxAdditionalData{
			GasPrice:      gasPrice,
			RevertMessage: result.RevertMessage,
		}
		s.ethTransactions[txHash].Txs[hash] = add
	}
}

// updateEthTxResult handles updating transaction state
func (s *SequenceSender) updateEthTxResult(txData *ethTxData, txResult ethtxmanager.MonitoredTxResult) {
	if txData.Status != txResult.Status.String() {
		log.Infof("update transaction %v to state %s", txResult.ID, txResult.Status.String())
		txData.StatusTimestamp = time.Now()
		stTrans := txData.StatusTimestamp.Format("2006-01-02T15:04:05.000-07:00") + ", " + txData.Status + ", " + txResult.Status.String()
		txData.Status = txResult.Status.String()
		txData.StateHistory = append(txData.StateHistory, stTrans)

		// Manage according to the state
		statusConsolidated := txData.Status == ethtxmanager.MonitoredTxStatusSafe.String() || txData.Status == ethtxmanager.MonitoredTxStatusFinalized.String()
		if txData.Status == ethtxmanager.MonitoredTxStatusFailed.String() {
			s.logFatalf("transaction %v result failed!")
		} else if statusConsolidated && txData.ToBatch >= s.latestVirtualBatch {
			s.latestVirtualTime = txData.StatusTimestamp
		}
	}

	// Update info received from L1
	txData.Nonce = txResult.Nonce
	if txResult.To != nil {
		txData.To = *txResult.To
	}
	if txResult.MinedAtBlockNumber != nil {
		txData.MinedAtBlock = *txResult.MinedAtBlockNumber
	}
	s.copyTxData(txResult.ID, txResult.Data, txResult.Txs)
}

// getResultAndUpdateEthTx updates the tx status from the ethTxManager
func (s *SequenceSender) getResultAndUpdateEthTx(ctx context.Context, txHash common.Hash) error {
	txData, exists := s.ethTransactions[txHash]
	if !exists {
		log.Errorf("transaction %v not found in memory", txHash)
		return errors.New("transaction not found in memory structure")
	}

	txResult, err := s.ethTxManager.Result(ctx, txHash)
	if err == ethtxmanager.ErrNotFound {
		log.Infof("transaction %v does not exist in ethtxmanager. Marking it", txHash)
		txData.OnMonitor = false
		// Resend tx
		errSend := s.sendTx(ctx, true, &txHash, nil, 0, 0, nil)
		if errSend == nil {
			txData.OnMonitor = false
		}
	} else if err != nil {
		log.Errorf("error getting result for tx %v: %v", txHash, err)
		return err
	} else {
		s.updateEthTxResult(txData, txResult)
	}

	return nil
}

// tryToSendSequence checks if there is a sequence and it's worth it to send to L1
func (s *SequenceSender) tryToSendSequence(ctx context.Context) {
	// Update latest virtual batch
	log.Infof("updating virtual batch")
	err := s.updateLatestVirtualBatch()
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
	firstSequence := sequence.FirstBatch()
	lastSequence := sequence.LastBatch()
	lastL2BlockTimestamp := lastSequence.LastL2BLockTimestamp()

	log.Infof("sending sequences to L1. From batch %d to batch %d", firstSequence.BatchNumber(), lastSequence.BatchNumber())
	log.Infof(sequence.String())

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
				waitTime, lastL1BlockHeader.Number, lastL1BlockHeader.Time, lastSequence.BatchNumber(), lastL2BlockTimestamp, timeMargin)
			time.Sleep(time.Duration(waitTime) * time.Second)
		} else {
			log.Infof("continuing, time difference between last L1 block %d (ts: %d) and last L2 block %d (ts: %d) in the sequence is greater than %d seconds",
				lastL1BlockHeader.Number, lastL1BlockHeader.Time, lastSequence.BatchNumber, lastL2BlockTimestamp, timeMargin)
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
				waitTime, currentTime, lastSequence.BatchNumber, lastL2BlockTimestamp, timeMargin)
			time.Sleep(time.Duration(waitTime) * time.Second)
		} else {
			log.Infof("[SeqSender]sending sequences now, time difference between now (ts: %d) and last L2 block %d (ts: %d) in the sequence is also greater than %d seconds",
				currentTime, lastSequence.BatchNumber, lastL2BlockTimestamp, timeMargin)
			break
		}
	}

	// Send sequences to L1
	log.Infof("sending sequences to L1. From batch %d to batch %d", firstSequence.BatchNumber(), lastSequence.BatchNumber())
	log.Infof(sequence.String())

	tx, err := s.TxBuilder.BuildSequenceBatchesTx(ctx, sequence)
	if err != nil {
		log.Errorf("error building sequenceBatches tx: %v", err)
		return
	}

	// Add sequence tx
	err = s.sendTx(ctx, false, nil, tx.To(), firstSequence.BatchNumber(), lastSequence.BatchNumber(), tx.Data())
	if err != nil {
		return
	}

	// Purge sequences data from memory
	s.purgeSequences()
}

// sendTx adds transaction to the ethTxManager to send it to L1
func (s *SequenceSender) sendTx(ctx context.Context, resend bool, txOldHash *common.Hash, to *common.Address, fromBatch uint64, toBatch uint64, data []byte) error {
	// Params if new tx to send or resend a previous tx
	var paramTo *common.Address
	var paramNonce *uint64
	var paramData []byte
	var valueFromBatch uint64
	var valueToBatch uint64
	var valueToAddress common.Address

	if !resend {
		paramTo = to
		paramNonce = &s.currentNonce
		paramData = data
		valueFromBatch = fromBatch
		valueToBatch = toBatch
	} else {
		if txOldHash == nil {
			log.Errorf("trying to resend a tx with nil hash")
			return errors.New("resend tx with nil hash monitor id")
		}
		paramTo = &s.ethTransactions[*txOldHash].To
		paramNonce = &s.ethTransactions[*txOldHash].Nonce
		paramData = s.ethTxData[*txOldHash]
		valueFromBatch = s.ethTransactions[*txOldHash].FromBatch
		valueToBatch = s.ethTransactions[*txOldHash].ToBatch
	}
	if paramTo != nil {
		valueToAddress = *paramTo
	}

	// Add sequence tx
	txHash, err := s.ethTxManager.Add(ctx, paramTo, paramNonce, big.NewInt(0), paramData, s.cfg.GasOffset, nil)
	if err != nil {
		log.Errorf("error adding sequence to ethtxmanager: %v", err)
		return err
	}
	if !resend {
		s.currentNonce++
	}

	// Add new eth tx
	txData := ethTxData{
		SentL1Timestamp: time.Now(),
		StatusTimestamp: time.Now(),
		Status:          "*new",
		FromBatch:       valueFromBatch,
		ToBatch:         valueToBatch,
		OnMonitor:       true,
		To:              valueToAddress,
	}

	// Add tx to internal structure
	s.mutexEthTx.Lock()
	s.ethTransactions[txHash] = &txData
	txResults := make(map[common.Hash]ethtxmanager.TxResult, 0)
	s.copyTxData(txHash, paramData, txResults)
	_ = s.getResultAndUpdateEthTx(ctx, txHash)
	if !resend {
		s.latestSentToL1Batch = valueToBatch
	} else {
		s.ethTransactions[*txOldHash].Status = "*resent"
	}
	s.mutexEthTx.Unlock()

	// Save sent sequences
	err = s.saveSentSequencesTransactions(ctx)
	if err != nil {
		log.Errorf("error saving tx sequence sent, error: %v", err)
	}
	return nil
}

// getSequencesToSend generates sequences to be sent to L1. Empty array means there are no sequences to send or it's not worth sending
func (s *SequenceSender) getSequencesToSend(ctx context.Context) (seqsendertypes.Sequence, error) {
	// Add sequences until too big for a single L1 tx or last batch is reached
	s.mutexSequence.Lock()
	defer s.mutexSequence.Unlock()
	var prevCoinbase common.Address
	sequenceBatches := make([]seqsendertypes.Batch, 0)
	for i := 0; i < len(s.sequenceList); i++ {
		batchNumber := s.sequenceList[i]
		if batchNumber <= s.latestVirtualBatch || batchNumber <= s.latestSentToL1Batch {
			continue
		}

		// Check if the next batch belongs to a new forkid, in this case we need to stop sequencing as we need to
		// wait the upgrade of forkid is completed and s.cfg.NumBatchForkIdUpgrade is disabled (=0) again
		if (s.cfg.ForkUpgradeBatchNumber != 0) && (batchNumber == (s.cfg.ForkUpgradeBatchNumber + 1)) {
			return nil, fmt.Errorf("aborting sequencing process as we reached the batch %d where a new forkid is applied (upgrade)", s.cfg.ForkUpgradeBatchNumber+1)
		}

		// Check if batch is closed
		if !s.sequenceData[batchNumber].batchClosed {
			// Reached current wip batch
			break
		}

		// New potential batch to add to the sequence
		batch := s.sequenceData[batchNumber].batch.DeepCopy()

		// If the coinbase changes, the sequence ends here
		if len(sequenceBatches) > 0 && batch.LastCoinbase() != prevCoinbase {
			log.Infof("batch with different coinbase (batch %v, sequence %v), sequence will be sent to this point", prevCoinbase, batch.LastCoinbase)
			return s.TxBuilder.NewSequence(sequenceBatches, s.cfg.L2Coinbase)
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
			return s.TxBuilder.NewSequence(sequenceBatches, s.cfg.L2Coinbase)
		}
	}

	// Reached the latest batch. Decide if it's worth to send the sequence, or wait for new batches
	if len(sequenceBatches) == 0 {
		log.Infof("no batches to be sequenced")
		return nil, nil
	}

	if s.latestVirtualTime.Before(time.Now().Add(-s.cfg.LastBatchVirtualizationTimeMaxWaitPeriod.Duration)) {
		log.Infof("sequence should be sent, too much time without sending anything to L1")
		return s.TxBuilder.NewSequence(sequenceBatches, s.cfg.L2Coinbase)
	}

	log.Infof("not enough time has passed since last batch was virtualized and the sequence could be bigger")
	return nil, nil
}

// loadSentSequencesTransactions loads the file into the memory structure
func (s *SequenceSender) loadSentSequencesTransactions() error {
	// Check if file exists
	if _, err := os.Stat(s.cfg.SequencesTxFileName); os.IsNotExist(err) {
		log.Infof("file not found %s: %v", s.cfg.SequencesTxFileName, err)
		return nil
	} else if err != nil {
		log.Errorf("error opening file %s: %v", s.cfg.SequencesTxFileName, err)
		return err
	}

	// Read file
	data, err := os.ReadFile(s.cfg.SequencesTxFileName)
	if err != nil {
		log.Errorf("error reading file %s: %v", s.cfg.SequencesTxFileName, err)
		return err
	}

	// Restore memory structure
	s.mutexEthTx.Lock()
	err = json.Unmarshal(data, &s.ethTransactions)
	s.mutexEthTx.Unlock()
	if err != nil {
		log.Errorf("error decoding data from %s: %v", s.cfg.SequencesTxFileName, err)
		return err
	}

	return nil
}

// saveSentSequencesTransactions saves memory structure into persistent file
func (s *SequenceSender) saveSentSequencesTransactions(ctx context.Context) error {
	var err error

	// Purge tx
	s.purgeEthTx(ctx)

	// Ceate file
	fileName := s.cfg.SequencesTxFileName[0:strings.IndexRune(s.cfg.SequencesTxFileName, '.')] + ".tmp"
	s.sequencesTxFile, err = os.Create(fileName)
	if err != nil {
		log.Errorf("error creating file %s: %v", fileName, err)
		return err
	}
	defer s.sequencesTxFile.Close()

	// Write data JSON encoded
	encoder := json.NewEncoder(s.sequencesTxFile)
	encoder.SetIndent("", "  ")
	s.mutexEthTx.Lock()
	err = encoder.Encode(s.ethTransactions)
	s.mutexEthTx.Unlock()
	if err != nil {
		log.Errorf("error writing file %s: %v", fileName, err)
		return err
	}

	// Rename the new file
	err = os.Rename(fileName, s.cfg.SequencesTxFileName)
	if err != nil {
		log.Errorf("error renaming file %s to %s: %v", fileName, s.cfg.SequencesTxFileName, err)
		return err
	}

	return nil
}

func (s *SequenceSender) entryTypeToString(entryType datastream.EntryType) string {
	switch entryType {
	case datastream.EntryType_ENTRY_TYPE_BATCH_START:
		return "BatchStart"
	case datastream.EntryType_ENTRY_TYPE_L2_BLOCK:
		return "L2Block"
	case datastream.EntryType_ENTRY_TYPE_TRANSACTION:
		return "Transaction"
	case datastream.EntryType_ENTRY_TYPE_BATCH_END:
		return "BatchEnd"
	default:
		return fmt.Sprintf("%d", entryType)
	}
}

// handleReceivedDataStream manages the events received by the streaming
func (s *SequenceSender) handleReceivedDataStream(entry *datastreamer.FileEntry, client *datastreamer.StreamClient, server *datastreamer.StreamServer) error {
	dsType := datastream.EntryType(entry.Type)

	var prevEntryType datastream.EntryType
	if s.prevStreamEntry != nil {
		prevEntryType = datastream.EntryType(s.prevStreamEntry.Type)
	}

	switch dsType {
	case datastream.EntryType_ENTRY_TYPE_L2_BLOCK:
		// Handle stream entry: L2Block
		l2Block := &datastream.L2Block{}

		err := proto.Unmarshal(entry.Data, l2Block)
		if err != nil {
			log.Errorf("error unmarshalling L2Block: %v", err)
			return err
		}

		log.Infof("received L2Block entry, l2Block.Number: %d, l2Block.BatchNumber: %d, entry.Number: %d", l2Block.Number, l2Block.BatchNumber, entry.Number)

		// Sanity checks
		if s.prevStreamEntry != nil && !(prevEntryType == datastream.EntryType_ENTRY_TYPE_BATCH_START || prevEntryType == datastream.EntryType_ENTRY_TYPE_L2_BLOCK || prevEntryType == datastream.EntryType_ENTRY_TYPE_TRANSACTION) {
			log.Fatalf("unexpected L2Block entry received, entry.Number: %d, l2Block.Number: %d, prevEntry: %s, prevEntry.Number: %d",
				entry.Number, l2Block.Number, s.entryTypeToString(prevEntryType), s.prevStreamEntry.Number)
		} else if prevEntryType == datastream.EntryType_ENTRY_TYPE_L2_BLOCK {
			prevL2Block := &datastream.L2Block{}

			err := proto.Unmarshal(s.prevStreamEntry.Data, prevL2Block)
			if err != nil {
				log.Errorf("error unmarshalling prevL2Block: %v", err)
				return err
			}
			if l2Block.Number != prevL2Block.Number+1 {
				log.Fatalf("unexpected L2Block number %d received, it should be %d, entry.Number: %d, prevEntry.Number: %d",
					l2Block.Number, prevL2Block.Number+1, entry.Number, s.prevStreamEntry.Number)
			}
		}

		// Already virtualized
		if l2Block.BatchNumber <= s.fromStreamBatch {
			if l2Block.BatchNumber != s.latestStreamBatch {
				log.Infof("skipped! batch already virtualized, number %d", l2Block.BatchNumber)
			}
		} else if !s.validStream && l2Block.BatchNumber == s.fromStreamBatch+1 {
			// Initial case after startup
			s.addNewSequenceBatch(l2Block)
			s.validStream = true
		} else {
			// Handle whether it's only a new block or also a new batch
			if l2Block.BatchNumber > s.wipBatch {
				// Create new sequential batch
				s.addNewSequenceBatch(l2Block)
			}
		}

		// Latest stream batch
		s.latestStreamBatch = l2Block.BatchNumber
		if !s.validStream {
			return nil
		}

		// Add L2 block
		s.addNewBatchL2Block(l2Block)

		s.prevStreamEntry = entry

	case datastream.EntryType_ENTRY_TYPE_TRANSACTION:
		// Handle stream entry: Transaction
		if !s.validStream {
			return nil
		}

		l2Tx := &datastream.Transaction{}
		err := proto.Unmarshal(entry.Data, l2Tx)
		if err != nil {
			log.Errorf("error unmarshalling Transaction: %v", err)
			return err
		}

		log.Debugf("received Transaction entry, tx.L2BlockNumber: %d, tx.Index: %d, entry.Number: %d", l2Tx.L2BlockNumber, l2Tx.Index, entry.Number)

		// Sanity checks
		if !(prevEntryType == datastream.EntryType_ENTRY_TYPE_L2_BLOCK || prevEntryType == datastream.EntryType_ENTRY_TYPE_TRANSACTION) {
			log.Fatalf("unexpected Transaction entry received, entry.Number: %d, transaction.L2BlockNumber: %d, transaction.Index: %d, prevEntry: %s, prevEntry.Number: %d",
				entry.Number, l2Tx.L2BlockNumber, l2Tx.Index, s.entryTypeToString(prevEntryType), s.prevStreamEntry.Number)
		}

		// Add tx data
		s.addNewBlockTx(l2Tx)

		s.prevStreamEntry = entry

	case datastream.EntryType_ENTRY_TYPE_BATCH_START:
		// Handle stream entry: BatchStart
		if !s.validStream {
			return nil
		}

		batch := &datastream.BatchStart{}
		err := proto.Unmarshal(entry.Data, batch)
		if err != nil {
			log.Errorf("error unmarshalling BatchStart: %v", err)
			return err
		}

		log.Infof("received BatchStart entry, batchStart.Number: %d, entry.Number: %d", batch.Number, entry.Number)

		// Sanity checks
		if !(prevEntryType == datastream.EntryType_ENTRY_TYPE_BATCH_END) {
			log.Fatalf("unexpected BatchStart entry received, entry.Number: %d, batchStart.Number: %d, prevEntry.Type: %s, prevEntry.Number: %d",
				entry.Number, batch.Number, s.entryTypeToString(prevEntryType), s.prevStreamEntry.Number)
		} else if batch.Number != s.wipBatch+1 {
			log.Fatalf("unexpected BatchStart.Number %d received, if should be wipBatch %d+1, entry.Number: %d", s.wipBatch, batch.Number, entry.Number)
		}

		// Add batch start data
		s.addInfoSequenceBatchStart(batch)

		s.prevStreamEntry = entry

	case datastream.EntryType_ENTRY_TYPE_BATCH_END:
		// Handle stream entry: BatchEnd
		if !s.validStream {
			return nil
		}

		batch := &datastream.BatchEnd{}
		err := proto.Unmarshal(entry.Data, batch)
		if err != nil {
			log.Errorf("error unmarshalling BatchEnd: %v", err)
			return err
		}

		log.Infof("received BatchEnd entry, batchEnd.Number: %d, entry.Number: %d", batch.Number, entry.Number)

		// Sanity checks
		if !(prevEntryType == datastream.EntryType_ENTRY_TYPE_L2_BLOCK || prevEntryType == datastream.EntryType_ENTRY_TYPE_TRANSACTION) {
			log.Fatalf("unexpected BatchEnd entry received, entry.Number: %d, batchEnd.Number: %d, prevEntry.Type: %s, prevEntry.Number: %d",
				entry.Number, batch.Number, s.entryTypeToString(prevEntryType), s.prevStreamEntry.Number)
		}

		// Add batch end data
		s.addInfoSequenceBatchEnd(batch)

		// Close current wip batch
		err = s.closeSequenceBatch()
		if err != nil {
			log.Fatalf("error closing wip batch")
			return err
		}

		s.prevStreamEntry = entry
	}

	return nil
}

// closeSequenceBatch closes the current batch
func (s *SequenceSender) closeSequenceBatch() error {
	s.mutexSequence.Lock()
	defer s.mutexSequence.Unlock()

	log.Infof("closing batch %d", s.wipBatch)

	data := s.sequenceData[s.wipBatch]
	if data != nil {
		data.batchClosed = true

		batchL2Data, err := state.EncodeBatchV2(data.batchRaw)
		if err != nil {
			log.Errorf("error closing and encoding the batch %d: %v", s.wipBatch, err)
			return err
		}

		data.batch.SetL2Data(batchL2Data)
	} else {
		log.Fatalf("wipBatch %d not found in sequenceData slice", s.wipBatch)
	}

	// Sanity Check
	if s.cfg.SanityCheckRPCURL != "" {
		rpcNumberOfBlocks, batchL2Data, err := s.getBatchFromRPC(s.wipBatch)
		if err != nil {
			log.Fatalf("error getting batch number from RPC while trying to perform sanity check: %v", err)
		} else {
			dsNumberOfBlocks := len(s.sequenceData[s.wipBatch].batchRaw.Blocks)
			if rpcNumberOfBlocks != dsNumberOfBlocks {
				log.Fatalf("number of blocks in batch %d (%d) does not match the number of blocks in the batch from the RPC (%d)", s.wipBatch, dsNumberOfBlocks, rpcNumberOfBlocks)
			}

			if data.batchType == datastream.BatchType_BATCH_TYPE_REGULAR && common.Bytes2Hex(data.batch.L2Data()) != batchL2Data {
				log.Infof("datastream batchL2Data: %s", common.Bytes2Hex(data.batch.L2Data()))
				log.Infof("RPC batchL2Data: %s", batchL2Data)
				log.Fatalf("batchL2Data in batch %d does not match batchL2Data from the RPC (%d)", s.wipBatch)
			}

			log.Infof("sanity check of batch %d against RPC successful", s.wipBatch)
		}
	} else {
		log.Warnf("config param SanityCheckRPCURL not set, sanity check with RPC can't be done")
	}

	return nil
}

func (s *SequenceSender) getBatchFromRPC(batchNumber uint64) (int, string, error) {
	type zkEVMBatch struct {
		Blocks      []string `mapstructure:"blocks"`
		BatchL2Data string   `mapstructure:"batchL2Data"`
	}

	zkEVMBatchData := zkEVMBatch{}

	response, err := rpc.JSONRPCCall(s.cfg.SanityCheckRPCURL, "zkevm_getBatchByNumber", batchNumber)
	if err != nil {
		return 0, "", err
	}

	// Check if the response is an error
	if response.Error != nil {
		return 0, "", fmt.Errorf("error in the response calling zkevm_getBatchByNumber: %v", response.Error)
	}

	// Get the batch number from the response hex string
	err = json.Unmarshal(response.Result, &zkEVMBatchData)
	if err != nil {
		return 0, "", fmt.Errorf("error unmarshalling the batch number from the response calling zkevm_getBatchByNumber: %v", err)
	}

	return len(zkEVMBatchData.Blocks), zkEVMBatchData.BatchL2Data, nil
}

// addNewSequenceBatch adds a new batch to the sequence
func (s *SequenceSender) addNewSequenceBatch(l2Block *datastream.L2Block) {
	s.mutexSequence.Lock()
	log.Infof("...new batch, number %d", l2Block.BatchNumber)

	if l2Block.BatchNumber > s.wipBatch+1 {
		s.logFatalf("new batch number (%d) is not consecutive to the current one (%d)", l2Block.BatchNumber, s.wipBatch)
	} else if l2Block.BatchNumber < s.wipBatch {
		s.logFatalf("new batch number (%d) is lower than the current one (%d)", l2Block.BatchNumber, s.wipBatch)
	}

	batch := s.TxBuilder.NewBatchFromL2Block(l2Block)

	// Add to the list
	s.sequenceList = append(s.sequenceList, l2Block.BatchNumber)

	// Create initial data
	batchRaw := state.BatchRawV2{}
	data := sequenceData{
		batchClosed: false,
		batch:       batch,
		batchRaw:    &batchRaw,
	}
	s.sequenceData[l2Block.BatchNumber] = &data

	// Update wip batch
	s.wipBatch = l2Block.BatchNumber
	s.mutexSequence.Unlock()
}

// addInfoSequenceBatchStart adds info from the batch start
func (s *SequenceSender) addInfoSequenceBatchStart(batch *datastream.BatchStart) {
	s.mutexSequence.Lock()
	log.Infof("batch %d (%s) Start: type %d forkId %d chainId %d", batch.Number, datastream.BatchType_name[int32(batch.Type)], batch.Type, batch.ForkId, batch.ChainId)

	// Current batch
	data := s.sequenceData[s.wipBatch]
	if data != nil {
		wipBatch := data.batch
		if wipBatch.BatchNumber()+1 != batch.Number {
			s.logFatalf("batch start number (%d) does not match the current consecutive one (%d)", batch.Number, wipBatch.BatchNumber)
		}
		data.batchType = batch.Type
	}

	s.mutexSequence.Unlock()
}

// addInfoSequenceBatchEnd adds info from the batch end
func (s *SequenceSender) addInfoSequenceBatchEnd(batch *datastream.BatchEnd) {
	s.mutexSequence.Lock()

	// Current batch
	data := s.sequenceData[s.wipBatch]
	if data != nil {
		wipBatch := data.batch
		if wipBatch.BatchNumber() == batch.Number {
			// wipBatch.StateRoot = common.BytesToHash(batch) TODO: check if this is needed
		} else {
			s.logFatalf("batch end number (%d) does not match the current one (%d)", batch.Number, wipBatch.BatchNumber)
		}
	}

	s.mutexSequence.Unlock()
}

// addNewBatchL2Block adds a new L2 block to the work in progress batch
func (s *SequenceSender) addNewBatchL2Block(l2Block *datastream.L2Block) {
	s.mutexSequence.Lock()
	log.Infof(".....new L2 block, number %d (batch %d)", l2Block.Number, l2Block.BatchNumber)

	// Current batch
	data := s.sequenceData[s.wipBatch]
	if data != nil {
		wipBatchRaw := data.batchRaw
		data.batch.SetLastL2BLockTimestamp(l2Block.Timestamp)
		// Sanity check: should be the same coinbase within the batch
		if common.BytesToAddress(l2Block.Coinbase) != data.batch.LastCoinbase() {
			s.logFatalf("coinbase changed within the batch! (Previous %v, Current %v)", data.batch.LastCoinbase, common.BytesToAddress(l2Block.Coinbase))
		}
		data.batch.SetLastCoinbase(common.BytesToAddress(l2Block.Coinbase))
		data.batch.SetL1InfoTreeIndex(l2Block.L1InfotreeIndex)
		// New L2 block raw
		newBlockRaw := state.L2BlockRaw{}

		// Add L2 block
		wipBatchRaw.Blocks = append(wipBatchRaw.Blocks, newBlockRaw)
		// Get current L2 block
		_, blockRaw := s.getWipL2Block()
		if blockRaw == nil {
			log.Debugf("wip block %d not found!")
			return
		}

		// Fill in data
		blockRaw.DeltaTimestamp = l2Block.DeltaTimestamp
		blockRaw.IndexL1InfoTree = l2Block.L1InfotreeIndex
	}

	s.mutexSequence.Unlock()
}

// addNewBlockTx adds a new Tx to the current L2 block
func (s *SequenceSender) addNewBlockTx(l2Tx *datastream.Transaction) {
	s.mutexSequence.Lock()
	log.Debugf("........new tx, length %d EGP %d SR %x..", len(l2Tx.Encoded), l2Tx.EffectiveGasPricePercentage, l2Tx.ImStateRoot[:8])

	// Current L2 block
	_, blockRaw := s.getWipL2Block()

	// New Tx raw
	tx, err := state.DecodeTx(common.Bytes2Hex(l2Tx.Encoded))
	if err != nil {
		log.Fatalf("error decoding tx: %v", err)
		return
	}

	l2TxRaw := state.L2TxRaw{
		EfficiencyPercentage: uint8(l2Tx.EffectiveGasPricePercentage),
		TxAlreadyEncoded:     false,
		Tx:                   tx,
	}

	// Add Tx
	blockRaw.Transactions = append(blockRaw.Transactions, l2TxRaw)
	s.mutexSequence.Unlock()
}

// getWipL2Block returns index of the array and pointer to the current L2 block (helper func)
func (s *SequenceSender) getWipL2Block() (uint64, *state.L2BlockRaw) {
	// Current batch
	var wipBatchRaw *state.BatchRawV2
	if s.sequenceData[s.wipBatch] != nil {
		wipBatchRaw = s.sequenceData[s.wipBatch].batchRaw
	}

	// Current wip block
	if len(wipBatchRaw.Blocks) > 0 {
		blockIndex := uint64(len(wipBatchRaw.Blocks)) - 1
		return blockIndex, &wipBatchRaw.Blocks[blockIndex]
	} else {
		return 0, nil
	}
}

// updateLatestVirtualBatch queries the value in L1 and updates the latest virtual batch field
func (s *SequenceSender) updateLatestVirtualBatch() error {
	// Get latest virtual state batch from L1
	var err error

	s.latestVirtualBatch, err = s.etherman.GetLatestBatchNumber()
	if err != nil {
		log.Errorf("error getting latest virtual batch, error: %v", err)
		return errors.New("fail to get latest virtual batch")
	} else {
		log.Infof("latest virtual batch is %d", s.latestVirtualBatch)
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
	log.Errorf(template, args...)
	log.Errorf("sequence sending stopped.")
	for {
		time.Sleep(1 * time.Second)
	}
}

// printSequenceBatches prints data from slice of type sequence batches
func printSequenceBatches(sequence *etherman.SequenceBanana) {
	for i, b := range sequence.Batches {
		log.Debugf("// sequence(%d): batch: %d, ts: %v, lenData: %d, GER: %x..",
			i, b.BatchNumber, b.LastL2BLockTimestamp, len(b.L2Data), b.GlobalExitRoot[:8])
	}
}

// printBatch prints data from batch raw V2
func printBatch(raw *state.BatchRawV2, showBlock bool, showTx bool) {
	// Total amount of L2 tx in the batch
	totalL2Txs := 0
	for k := 0; k < len(raw.Blocks); k++ {
		totalL2Txs += len(raw.Blocks[k].Transactions)
	}

	log.Debugf("// #blocks: %d, #L2txs: %d", len(raw.Blocks), totalL2Txs)

	// Blocks info
	if showBlock {
		numBlocks := len(raw.Blocks)
		var firstBlock *state.L2BlockRaw
		var lastBlock *state.L2BlockRaw
		if numBlocks > 0 {
			firstBlock = &raw.Blocks[0]
		}
		if numBlocks > 1 {
			lastBlock = &raw.Blocks[numBlocks-1]
		}
		if firstBlock != nil {
			log.Debugf("//    block first (indL1info: %d, delta-timestamp: %d, #L2txs: %d)", firstBlock.IndexL1InfoTree, firstBlock.DeltaTimestamp, len(firstBlock.Transactions))
			// Tx info
			if showTx {
				for iTx, tx := range firstBlock.Transactions {
					v, r, s := tx.Tx.RawSignatureValues()
					log.Debugf("//       tx(%d) effPct: %d, encoded: %t, v: %v, r: %v, s: %v", iTx, tx.EfficiencyPercentage, tx.TxAlreadyEncoded, v, r, s)
				}
			}
		}
		if lastBlock != nil {
			log.Debugf("//    block last (indL1info: %d, delta-timestamp: %d, #L2txs: %d)", lastBlock.DeltaTimestamp, lastBlock.DeltaTimestamp, len(lastBlock.Transactions))
		}
	}
}
