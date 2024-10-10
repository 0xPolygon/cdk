package sequencesender

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/zkevm-ethtx-manager/ethtxmanager"
	"github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum/common"
)

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
	Gas             uint64                              `json:"gas"`
}

type ethTxAdditionalData struct {
	GasPrice      *big.Int `json:"gasPrice,omitempty"`
	RevertMessage string   `json:"revertMessage,omitempty"`
}

// sendTx adds transaction to the ethTxManager to send it to L1
func (s *SequenceSender) sendTx(ctx context.Context, resend bool, txOldHash *common.Hash, to *common.Address,
	fromBatch uint64, toBatch uint64, data []byte, gas uint64) error {
	// Params if new tx to send or resend a previous tx
	var (
		paramTo        *common.Address
		paramData      []byte
		valueFromBatch uint64
		valueToBatch   uint64
		valueToAddress common.Address
	)

	if !resend {
		paramTo = to
		paramData = data
		valueFromBatch = fromBatch
		valueToBatch = toBatch
	} else {
		if txOldHash == nil {
			log.Errorf("trying to resend a tx with nil hash")
			return errors.New("resend tx with nil hash monitor id")
		}
		oldEthTx := s.ethTransactions[*txOldHash]
		paramTo = &oldEthTx.To
		paramData = s.ethTxData[*txOldHash]
		valueFromBatch = oldEthTx.FromBatch
		valueToBatch = oldEthTx.ToBatch
	}
	if paramTo != nil {
		valueToAddress = *paramTo
	}

	// Add sequence tx
	txHash, err := s.ethTxManager.AddWithGas(ctx, paramTo, big.NewInt(0), paramData, s.cfg.GasOffset, nil, gas)
	if err != nil {
		log.Errorf("error adding sequence to ethtxmanager: %v", err)
		return err
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
		Gas:             gas,
	}

	// Add tx to internal structure
	s.mutexEthTx.Lock()
	s.ethTransactions[txHash] = &txData
	txResults := make(map[common.Hash]types.TxResult, 0)
	s.copyTxData(txHash, paramData, txResults)
	err = s.getResultAndUpdateEthTx(ctx, txHash)
	if err != nil {
		log.Errorf("error getting result for tx %v: %v", txHash, err)
	}
	if !resend {
		atomic.StoreUint64(&s.latestSentToL1Batch, valueToBatch)
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

// purgeEthTx purges transactions from memory structures
func (s *SequenceSender) purgeEthTx(ctx context.Context) {
	// If sequence sending is stopped, do not purge
	if s.IsStopped() {
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

		if !data.OnMonitor || data.Status == types.MonitoredTxStatusFinalized.String() {
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
	var (
		txPending uint64
		txSync    uint64
	)
	for hash, tx := range s.ethTransactions {
		if tx.Status == types.MonitoredTxStatusFinalized.String() {
			continue
		}

		err := s.getResultAndUpdateEthTx(ctx, hash)
		if err != nil {
			log.Errorf("error getting result for tx %v: %v", hash, err)
			return 0, err
		}

		txSync++
		txStatus := types.MonitoredTxStatus(tx.Status)
		// Count if it is not in a final state
		if tx.OnMonitor &&
			txStatus != types.MonitoredTxStatusFailed &&
			txStatus != types.MonitoredTxStatusSafe &&
			txStatus != types.MonitoredTxStatusFinalized {
			txPending++
		}
	}
	s.mutexEthTx.Unlock()

	// Save updated sequences transactions
	err := s.saveSentSequencesTransactions(ctx)
	if err != nil {
		log.Errorf("error saving tx sequence, error: %v", err)
		return 0, err
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
			log.Debugf("transaction %v missing in memory structure. Adding it", result.ID)
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
func (s *SequenceSender) copyTxData(
	txHash common.Hash, txData []byte, txsResults map[common.Hash]types.TxResult,
) {
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
func (s *SequenceSender) updateEthTxResult(txData *ethTxData, txResult types.MonitoredTxResult) {
	if txData.Status != txResult.Status.String() {
		log.Infof("update transaction %v to state %s", txResult.ID, txResult.Status.String())
		txData.StatusTimestamp = time.Now()
		stTrans := txData.StatusTimestamp.Format("2006-01-02T15:04:05.000-07:00") +
			", " + txData.Status + ", " + txResult.Status.String()
		txData.Status = txResult.Status.String()
		txData.StateHistory = append(txData.StateHistory, stTrans)

		// Manage according to the state
		statusConsolidated := txData.Status == types.MonitoredTxStatusSafe.String() ||
			txData.Status == types.MonitoredTxStatusFinalized.String()
		if txData.Status == types.MonitoredTxStatusFailed.String() {
			s.logFatalf("transaction %v result failed!")
		} else if statusConsolidated && txData.ToBatch >= atomic.LoadUint64(&s.latestVirtualBatchNumber) {
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
		s.logger.Errorf("transaction %v not found in memory", txHash)
		return errors.New("transaction not found in memory structure")
	}

	txResult, err := s.ethTxManager.Result(ctx, txHash)
	switch {
	case errors.Is(err, ethtxmanager.ErrNotFound):
		s.logger.Infof("transaction %v does not exist in ethtxmanager. Marking it", txHash)
		txData.OnMonitor = false
		// Resend tx
		errSend := s.sendTx(ctx, true, &txHash, nil, 0, 0, nil, txData.Gas)
		if errSend == nil {
			txData.OnMonitor = false
		}

	case err != nil:
		s.logger.Errorf("error getting result for tx %v: %v", txHash, err)
		return err

	default:
		s.updateEthTxResult(txData, txResult)
	}

	return nil
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

	// Create file
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

// IsStopped returns true in case seqSendingStopped is set to 1, otherwise false
func (s *SequenceSender) IsStopped() bool {
	return atomic.LoadUint32(&s.seqSendingStopped) == 1
}
