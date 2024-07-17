package txbuilder

import (
	"context"
	"errors"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/ethereum/go-ethereum/common"
)

var (
	// ErrOversizedData when transaction input data is greater than a limit (DOS protection)
	ErrOversizedData = errors.New("oversized data")
)

type NewSequenceConditionalMaxSize struct {
	maxTxSizeForL1 uint64 // cfg.MaxTxSizeForL1
}

func (c *NewSequenceConditionalMaxSize) NewSequenceIfWorthToSend(ctx context.Context, txBuilder TxBuilder, sequenceBatches []seqsendertypes.Batch, senderAddress, l2Coinbase common.Address, batchNumber uint64) (seqsendertypes.Sequence, error) {
	sequence, err := txBuilder.NewSequence(sequenceBatches, l2Coinbase)
	if err != nil {
		return nil, err
	}

	// Check if can be sent
	tx, err := txBuilder.BuildSequenceBatchesTx(ctx, senderAddress, sequence)
	if err == nil && tx.Size() > c.maxTxSizeForL1 {
		log.Infof("[SeqSender] oversized Data on TX oldHash %s (txSize %d > %d)", tx.Hash(), tx.Size(), c.maxTxSizeForL1)
		err = ErrOversizedData
	}

	if err != nil {
		log.Infof("[SeqSender] handling estimate gas send sequence error: %v", err)
		sequenceBatches, err = handleEstimateGasSendSequenceErr(sequence.Batches(), batchNumber, err)
		if sequenceBatches != nil {
			// Handling the error gracefully, re-processing the sequence as a sanity check
			//sequence, err = s.newSequenceBanana(sequenceBatches, s.cfg.L2Coinbase)
			sequence, err = txBuilder.NewSequence(sequenceBatches, l2Coinbase)
			if err != nil {
				return nil, err
			}

			_, err = txBuilder.BuildSequenceBatchesTx(ctx, senderAddress, sequence)
			return sequence, err
		}

		return sequence, err
	}
	return nil, nil
}

// handleEstimateGasSendSequenceErr handles an error on the estimate gas. Results: (nil,nil)=requires waiting, (nil,error)=no handled gracefully, (seq,nil) handled gracefully
func handleEstimateGasSendSequenceErr(sequenceBatches []seqsendertypes.Batch, currentBatchNumToSequence uint64, err error) ([]seqsendertypes.Batch, error) {
	// Insufficient allowance
	if errors.Is(err, etherman.ErrInsufficientAllowance) {
		return nil, err
	}
	if isDataForEthTxTooBig(err) {
		// Remove the latest item and send the sequences
		log.Infof("Done building sequences, selected batches to %d. Batch %d caused the L1 tx to be too big: %v", currentBatchNumToSequence-1, currentBatchNumToSequence, err)
	} else {
		// Remove the latest item and send the sequences
		log.Infof("Done building sequences, selected batches to %d. Batch %d excluded due to unknown error: %v", currentBatchNumToSequence, currentBatchNumToSequence+1, err)
	}

	if len(sequenceBatches) > 1 {
		sequenceBatches = sequenceBatches[:len(sequenceBatches)-1]
	} else {
		sequenceBatches = nil
	}

	return sequenceBatches, nil
}

// isDataForEthTxTooBig checks if tx oversize error
func isDataForEthTxTooBig(err error) bool {
	return errors.Is(err, etherman.ErrGasRequiredExceedsAllowance) ||
		errors.Is(err, ErrOversizedData) ||
		errors.Is(err, etherman.ErrContentLengthTooLarge)
}
