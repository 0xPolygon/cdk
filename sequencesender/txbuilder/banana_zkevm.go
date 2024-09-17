package txbuilder

import (
	"context"
	"math/big"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type TxBuilderBananaZKEVM struct {
	TxBuilderBananaBase
	condNewSeq     CondNewSequence
	rollupContract rollupBananaZKEVMContractor
}

type rollupBananaZKEVMContractor interface {
	rollupBananaBaseContractor
	SequenceBatches(
		opts *bind.TransactOpts,
		batches []polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData,
		indexL1InfoRoot uint32,
		maxSequenceTimestamp uint64,
		expectedFinalAccInputHash [32]byte,
		l2Coinbase common.Address,
	) (*types.Transaction, error)
}

type globalExitRootBananaZKEVMContractor interface {
	globalExitRootBananaContractor
}

func NewTxBuilderBananaZKEVM(
	logger *log.Logger,
	rollupContract rollupBananaZKEVMContractor,
	gerContract globalExitRootBananaZKEVMContractor,
	opts bind.TransactOpts,
	maxTxSizeForL1 uint64,
	l1InfoTree l1InfoSyncer,
	ethClient l1Client,
	blockFinality *big.Int,
) *TxBuilderBananaZKEVM {
	txBuilderBase := *NewTxBuilderBananaBase(logger, rollupContract,
		gerContract, l1InfoTree, ethClient, blockFinality, opts)

	return &TxBuilderBananaZKEVM{
		TxBuilderBananaBase: txBuilderBase,
		condNewSeq:          NewConditionalNewSequenceMaxSize(maxTxSizeForL1),
		rollupContract:      rollupContract,
	}
}

func (t *TxBuilderBananaZKEVM) NewSequenceIfWorthToSend(
	ctx context.Context, sequenceBatches []seqsendertypes.Batch, l2Coinbase common.Address, batchNumber uint64,
) (seqsendertypes.Sequence, error) {
	return t.condNewSeq.NewSequenceIfWorthToSend(ctx, t, sequenceBatches, l2Coinbase)
}

// SetCondNewSeq allow to override the default conditional for new sequence
func (t *TxBuilderBananaZKEVM) SetCondNewSeq(cond CondNewSequence) CondNewSequence {
	previous := t.condNewSeq
	t.condNewSeq = cond
	return previous
}

func (t *TxBuilderBananaZKEVM) BuildSequenceBatchesTx(
	ctx context.Context, sequences seqsendertypes.Sequence,
) (*types.Transaction, error) {
	var err error
	ethseq, err := convertToSequenceBanana(sequences)
	if err != nil {
		t.logger.Error("error converting sequences to etherman: ", err)
		return nil, err
	}
	newopts := t.opts
	newopts.NoSend = true

	// force nonce, gas limit and gas price to avoid querying it from the chain
	newopts.Nonce = big.NewInt(1)
	newopts.GasLimit = uint64(1)
	newopts.GasPrice = big.NewInt(1)
	// Build sequence data
	tx, err := t.sequenceBatchesRollup(newopts, ethseq)
	if err != nil {
		t.logger.Errorf("error estimating new sequenceBatches to add to ethtxmanager: ", err)
		return nil, err
	}
	return tx, nil
}

func (t *TxBuilderBananaZKEVM) sequenceBatchesRollup(
	opts bind.TransactOpts, sequence etherman.SequenceBanana,
) (*types.Transaction, error) {
	batches := make([]polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData, len(sequence.Batches))
	for i, batch := range sequence.Batches {
		var ger common.Hash
		if batch.ForcedBatchTimestamp > 0 {
			ger = batch.ForcedGlobalExitRoot
		}

		batches[i] = polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData{
			Transactions:         batch.L2Data,
			ForcedGlobalExitRoot: ger,
			ForcedTimestamp:      batch.ForcedBatchTimestamp,
			ForcedBlockHashL1:    batch.ForcedBlockHashL1,
		}
	}

	tx, err := t.rollupContract.SequenceBatches(
		&opts, batches, sequence.CounterL1InfoRoot, sequence.MaxSequenceTimestamp, sequence.AccInputHash, sequence.L2Coinbase,
	)
	if err != nil {
		t.logger.Debugf("Batches to send: %+v", batches)
		t.logger.Debug("l2CoinBase: ", sequence.L2Coinbase)
		t.logger.Debug("Sequencer address: ", opts.From)
	}

	return tx, err
}

func (t *TxBuilderBananaZKEVM) String() string {
	return "Banana/ZKEVM"
}
