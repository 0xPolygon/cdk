package txbuilder

import (
	"context"
	"errors"
	"fmt"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/ethereum/go-ethereum/common"
)

var (
	// ErrOversizedData when transaction input data is greater than a limit (DOS protection)
	ErrOversizedData       = errors.New("oversized data")
	MaxTxSizeForL1Disabled = uint64(0)
)

type ConditionalNewSequenceMaxSize struct {
	maxTxSizeForL1 uint64 // cfg.MaxTxSizeForL1
}

func NewConditionalNewSequenceMaxSize(maxTxSizeForL1 uint64) *ConditionalNewSequenceMaxSize {
	return &ConditionalNewSequenceMaxSize{
		maxTxSizeForL1: maxTxSizeForL1,
	}
}

func (c *ConditionalNewSequenceMaxSize) NewSequenceIfWorthToSend(
	ctx context.Context, txBuilder TxBuilder, sequenceBatches []seqsendertypes.Batch, l2Coinbase common.Address,
) (seqsendertypes.Sequence, error) {
	if c.maxTxSizeForL1 == MaxTxSizeForL1Disabled {
		log.Debugf("maxTxSizeForL1 is %d, so is disabled", MaxTxSizeForL1Disabled)
		return nil, nil
	}
	sequence, err := txBuilder.NewSequence(ctx, sequenceBatches, l2Coinbase)
	if err != nil {
		return nil, err
	}
	if sequence == nil {
		err = fmt.Errorf("error txBuilder.NewSequence, returns sequence=nil and err==nil, is not expected")
		log.Errorf(err.Error())
		return nil, err
	}

	// Check if can be sent
	tx, err := txBuilder.BuildSequenceBatchesTx(ctx, sequence)
	if tx == nil && err == nil {
		err = fmt.Errorf("error txBuilder.BuildSequenceBatchesTx, returns tx=nil and err==nil, is not expected")
		log.Errorf(err.Error())
		return nil, err
	}
	if err == nil && tx != nil && tx.Size() > c.maxTxSizeForL1 {
		log.Infof("Oversized Data on TX oldHash %s (txSize %d > %d)", tx.Hash(), tx.Size(), c.maxTxSizeForL1)
		err = ErrOversizedData
	}

	if err != nil {
		log.Debugf("Handling estimate gas send sequence error: %v", err)
		sequenceBatches, err = handleEstimateGasSendSequenceErr(sequence.Batches(), err)
		if sequenceBatches != nil {
			// Handling the error gracefully, re-processing the sequence as a sanity check
			//sequence, err = s.newSequenceBanana(sequenceBatches, s.cfg.L2Coinbase)
			sequence, err = txBuilder.NewSequence(ctx, sequenceBatches, l2Coinbase)
			if err != nil {
				return nil, err
			}

			txReduced, err := txBuilder.BuildSequenceBatchesTx(ctx, sequence)
			log.Debugf("After reducing batches:  (txSize %d -> %d)", tx.Size(), txReduced.Size())
			if err == nil && txReduced != nil && txReduced.Size() > c.maxTxSizeForL1 {
				log.Warnf("After reducing batches:  (txSize %d -> %d) is still too big > %d",
					tx.Size(), txReduced.Size(), c.maxTxSizeForL1,
				)
			}
			return sequence, err
		}

		return sequence, err
	}
	log.Debugf(
		"Current size:%d < max_size:%d  num_batches: %d, no sequence promoted yet",
		tx.Size(), c.maxTxSizeForL1, sequence.Len(),
	)
	return nil, nil
}

// handleEstimateGasSendSequenceErr handles an error on the estimate gas.
// Results: (nil,nil)=requires waiting, (nil,error)=no handled gracefully, (seq,nil) handled gracefully
func handleEstimateGasSendSequenceErr(
	sequenceBatches []seqsendertypes.Batch, err error,
) ([]seqsendertypes.Batch, error) {
	// Insufficient allowance
	if errors.Is(err, etherman.ErrInsufficientAllowance) {
		return nil, err
	}
	errMsg := fmt.Sprintf("due to unknown error: %v", err)
	if isDataForEthTxTooBig(err) {
		errMsg = fmt.Sprintf("caused the L1 tx to be too big: %v", err)
	}
	var adjustMsg string
	if len(sequenceBatches) > 1 {
		lastPrevious := sequenceBatches[len(sequenceBatches)-1].BatchNumber()
		sequenceBatches = sequenceBatches[:len(sequenceBatches)-1]
		lastCurrent := sequenceBatches[len(sequenceBatches)-1].BatchNumber()
		adjustMsg = fmt.Sprintf(
			"removing last batch: old  BatchNumber:%d ->  %d, new length: %d",
			lastPrevious, lastCurrent, len(sequenceBatches),
		)
	} else {
		sequenceBatches = nil
		adjustMsg = "removing all batches"
		log.Warnf("No more batches to remove, sequence is empty... it could be a deadlock situation")
	}
	log.Infof("Adjusted sequence, %s, because %s", adjustMsg, errMsg)
	return sequenceBatches, nil
}

// isDataForEthTxTooBig checks if tx oversize error
func isDataForEthTxTooBig(err error) bool {
	return errors.Is(err, etherman.ErrGasRequiredExceedsAllowance) ||
		errors.Is(err, ErrOversizedData) ||
		errors.Is(err, etherman.ErrContentLengthTooLarge)
}
