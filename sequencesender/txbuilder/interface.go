package txbuilder

import (
	"context"
	"fmt"

	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type TxBuilder interface {
	// Stringer interface
	fmt.Stringer

	// BuildSequenceBatchesTx  Builds a sequence of batches transaction
	BuildSequenceBatchesTx(ctx context.Context, sequences seqsendertypes.Sequence) (*ethtypes.Transaction, error)
	// NewSequence  Creates a new sequence
	NewSequence(
		ctx context.Context, batches []seqsendertypes.Batch, coinbase common.Address,
	) (seqsendertypes.Sequence, error)
	// NewSequenceIfWorthToSend  Creates a new sequence if it is worth sending
	NewSequenceIfWorthToSend(
		ctx context.Context, sequenceBatches []seqsendertypes.Batch, l2Coinbase common.Address, batchNumber uint64,
	) (seqsendertypes.Sequence, error)
	// NewBatchFromL2Block  Creates a new batch from the L2 block from a datastream
	NewBatchFromL2Block(l2Block *datastream.L2Block) seqsendertypes.Batch
	// SetCondNewSeq  Allows to override the condition to send a new sequence, returns previous one
	SetCondNewSeq(cond CondNewSequence) CondNewSequence
}

type CondNewSequence interface {
	//NewSequenceIfWorthToSend  Return nil, nil if the sequence is not worth sending
	NewSequenceIfWorthToSend(
		ctx context.Context, txBuilder TxBuilder, sequenceBatches []seqsendertypes.Batch, l2Coinbase common.Address,
	) (seqsendertypes.Sequence, error)
}
