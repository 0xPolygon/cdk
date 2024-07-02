package txbuilder

import (
	"context"
	"fmt"

	"github.com/0xPolygon/cdk/etherman/types"
	"github.com/0xPolygon/cdk/sequencesender/etherman2seqsender"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type BuildSequenceBatchesTxZKEVM struct {
	etherman   etherman2seqsender.Eth2SequenceBatchesTxZKEVMBuilder
	L2Coinbase common.Address
	opts       *bind.TransactOpts
}

func NewBuildSequenceBatchesTxZKEVM(etherman etherman2seqsender.Eth2SequenceBatchesTxZKEVMBuilder, l2Coinbase common.Address, opts *bind.TransactOpts) *BuildSequenceBatchesTxZKEVM {
	return &BuildSequenceBatchesTxZKEVM{
		etherman:   etherman,
		L2Coinbase: l2Coinbase,
		opts:       opts,
	}
}

func (b *BuildSequenceBatchesTxZKEVM) BuildSequenceBatchesTx(ctx context.Context, sequences []types.Sequence, lastSequence types.Sequence, firstSequence types.Sequence) (*ethtypes.Transaction, error) {

	// Build sequence data
	params := etherman2seqsender.BuildSequenceBatchesParams{

		Sequences:                sequences,
		MaxSequenceTimestamp:     lastSequence.LastL2BLockTimestamp,
		LastSequencedBatchNumber: firstSequence.BatchNumber - 1,
		L2Coinbase:               b.L2Coinbase,
	}
	tx, err := b.etherman.BuildSequenceBatchesTxZKEVM(b.opts, params)
	if err != nil {
		err = fmt.Errorf("[SeqSender] error estimating new sequenceBatches to add to ethtxmanager: Err:%w ", err)
		return nil, err
	}
	return tx, nil
}
