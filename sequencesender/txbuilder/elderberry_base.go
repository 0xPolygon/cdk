package txbuilder

import (
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type TxBuilderElderberryBase struct {
	opts bind.TransactOpts
}

func NewTxBuilderElderberryBase(opts bind.TransactOpts) *TxBuilderElderberryBase {
	return &TxBuilderElderberryBase{
		opts: opts,
	}
}

// SetAuth sets the auth for the tx builder
func (t *TxBuilderElderberryBase) SetAuth(auth *bind.TransactOpts) {
	t.opts = *auth
}

func (t *TxBuilderElderberryBase) NewSequence(batches []seqsendertypes.Batch, coinbase common.Address) (seqsendertypes.Sequence, error) {
	seq := ElderberrySequence{
		l2Coinbase: coinbase,
		batches:    batches,
	}
	return &seq, nil
}

func (t *TxBuilderElderberryBase) NewBatchFromL2Block(l2Block *datastream.L2Block) seqsendertypes.Batch {
	batch := &etherman.Batch{
		LastL2BLockTimestamp: l2Block.Timestamp,
		BatchNumber:          l2Block.BatchNumber,
		L1InfoTreeIndex:      l2Block.L1InfotreeIndex,
		LastCoinbase:         common.BytesToAddress(l2Block.Coinbase),
		GlobalExitRoot:       common.BytesToHash(l2Block.GlobalExitRoot),
	}
	return NewBananaBatch(batch)
}

func getLastSequencedBatchNumber(sequences seqsendertypes.Sequence) uint64 {
	if sequences.Len() == 0 {
		return 0
	}
	if sequences.FirstBatch().BatchNumber() == 0 {
		panic("First batch number is 0, that is not allowed!")
	}
	if sequences.LastVirtualBatchNumber() != 0 {
		return sequences.LastVirtualBatchNumber()
	}
	return sequences.FirstBatch().BatchNumber() - 1
}
