package txbuilder

import (
	"context"

	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/ethereum/go-ethereum/common"
)

var MaxBatchesForL1Disabled = uint64(0)

type ConditionalNewSequenceNumBatches struct {
	maxBatchesForL1 uint64 // cfg.MaxBatchesForL1
}

func NewConditionalNewSequenceNumBatches(maxBatchesForL1 uint64) *ConditionalNewSequenceNumBatches {
	return &ConditionalNewSequenceNumBatches{
		maxBatchesForL1: maxBatchesForL1,
	}
}

func (c *ConditionalNewSequenceNumBatches) NewSequenceIfWorthToSend(
	ctx context.Context, txBuilder TxBuilder, sequenceBatches []seqsendertypes.Batch, l2Coinbase common.Address,
) (seqsendertypes.Sequence, error) {
	if c.maxBatchesForL1 != MaxBatchesForL1Disabled && uint64(len(sequenceBatches)) >= c.maxBatchesForL1 {
		log.Infof(
			"sequence should be sent to L1, because MaxBatchesForL1 (%d) has been reached",
			c.maxBatchesForL1,
		)
		return txBuilder.NewSequence(ctx, sequenceBatches, l2Coinbase)
	}
	return nil, nil
}
