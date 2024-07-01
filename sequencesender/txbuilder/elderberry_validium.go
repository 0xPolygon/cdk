package txbuilder

import (
	"context"
	"fmt"

	"github.com/0xPolygon/cdk/dataavailability"
	"github.com/0xPolygon/cdk/etherman/types"
	"github.com/0xPolygon/cdk/sequencesender/etherman2seqsender"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type BuildSequenceBatchesTxValidium struct {
	da         *dataavailability.DataAvailability
	etherman   etherman2seqsender.Eth2SequenceBatchesTxValidiumBuilder
	L2Coinbase common.Address
	opts       *bind.TransactOpts
}

func NewBuildSequenceBatchesTxValidium(da *dataavailability.DataAvailability, etherman etherman2seqsender.Eth2SequenceBatchesTxValidiumBuilder, l2Coinbase common.Address, opts *bind.TransactOpts) *BuildSequenceBatchesTxValidium {
	return &BuildSequenceBatchesTxValidium{
		da:         da,
		etherman:   etherman,
		L2Coinbase: l2Coinbase,
		opts:       opts,
	}
}

func (b *BuildSequenceBatchesTxValidium) BuildSequenceBatchesTx(ctx context.Context, sequences []types.Sequence, lastSequence types.Sequence, firstSequence types.Sequence) (*ethtypes.Transaction, error) {
	dataAvailabilityMessage, err := b.da.PostSequence(ctx, sequences)
	if err != nil {
		err = fmt.Errorf("error posting sequences to the data availability protocol: Err:%w ", err)
		return nil, err
	}

	// Build sequence data
	params := etherman2seqsender.BuildSequenceBatchesParams{
		Sequences:                sequences,
		MaxSequenceTimestamp:     lastSequence.LastL2BLockTimestamp,
		LastSequencedBatchNumber: firstSequence.BatchNumber - 1,
		L2Coinbase:               b.L2Coinbase,
	}
	tx, err := b.etherman.BuildSequenceBatchesTxValidium(b.opts, params, dataAvailabilityMessage)
	if err != nil {
		err = fmt.Errorf("[SeqSender] error estimating new sequenceBatches to add to ethtxmanager: Err:%w ", err)
		return nil, err
	}
	return tx, nil
}
