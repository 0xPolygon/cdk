package txbuilder

import (
	"context"

	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type TxBuilder interface {
	BuildSequenceBatchesTx(ctx context.Context, sender common.Address, sequences seqsendertypes.Sequence) (*ethtypes.Transaction, error)
	NewSequence(batches []seqsendertypes.Batch, coinbase common.Address) (seqsendertypes.Sequence, error)
	NewSequenceIfWorthToSend(ctx context.Context, sequenceBatches []seqsendertypes.Batch, l2Coinbase common.Address, batchNumber uint64) (seqsendertypes.Sequence, error)
	NewBatchFromL2Block(l2Block *datastream.L2Block) seqsendertypes.Batch
	String() string
}

type CondNewSequence interface {
	//NewSequenceIfWorthToSend  Return nil, nil if the sequence is not worth sending
	NewSequenceIfWorthToSend(ctx context.Context, txBuilder TxBuilder, sequenceBatches []seqsendertypes.Batch, senderAddress, l2Coinbase common.Address) (seqsendertypes.Sequence, error)
}
