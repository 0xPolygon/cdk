package txbuilder

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type TxBuilderElderberryZKEVM struct {
	TxBuilderElderberryBase
	condNewSeq     CondNewSequence
	rollupContract rollupElderberryZKEVMContractor
}

type rollupElderberryZKEVMContractor interface {
	SequenceBatches(
		opts *bind.TransactOpts,
		batches []polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData,
		maxSequenceTimestamp uint64,
		initSequencedBatch uint64,
		l2Coinbase common.Address,
	) (*types.Transaction, error)
}

func NewTxBuilderElderberryZKEVM(
	logger *log.Logger, zkevm rollupElderberryZKEVMContractor,
	opts bind.TransactOpts, maxTxSizeForL1 uint64,
) *TxBuilderElderberryZKEVM {
	return &TxBuilderElderberryZKEVM{
		TxBuilderElderberryBase: *NewTxBuilderElderberryBase(logger, opts),
		condNewSeq:              NewConditionalNewSequenceMaxSize(maxTxSizeForL1),
		rollupContract:          zkevm,
	}
}

func (t *TxBuilderElderberryZKEVM) NewSequenceIfWorthToSend(
	ctx context.Context, sequenceBatches []seqsendertypes.Batch, l2Coinbase common.Address, batchNumber uint64,
) (seqsendertypes.Sequence, error) {
	return t.condNewSeq.NewSequenceIfWorthToSend(ctx, t, sequenceBatches, l2Coinbase)
}

// SetCondNewSeq allow to override the default conditional for new sequence
func (t *TxBuilderElderberryZKEVM) SetCondNewSeq(cond CondNewSequence) CondNewSequence {
	previous := t.condNewSeq
	t.condNewSeq = cond
	return previous
}

func (t *TxBuilderElderberryZKEVM) BuildSequenceBatchesTx(
	ctx context.Context, sequences seqsendertypes.Sequence,
) (*types.Transaction, error) {
	newopts := t.opts
	newopts.NoSend = true

	// force nonce, gas limit and gas price to avoid querying it from the chain
	newopts.Nonce = big.NewInt(1)
	newopts.GasLimit = uint64(1)
	newopts.GasPrice = big.NewInt(1)

	return t.sequenceBatchesRollup(newopts, sequences)
}

func (t *TxBuilderElderberryZKEVM) sequenceBatchesRollup(
	opts bind.TransactOpts, sequences seqsendertypes.Sequence,
) (*types.Transaction, error) {
	if sequences == nil || sequences.Len() == 0 {
		return nil, fmt.Errorf("can't sequence an empty sequence")
	}
	batches := make([]polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData, sequences.Len())
	for i, seq := range sequences.Batches() {
		var ger common.Hash
		if seq.ForcedBatchTimestamp() > 0 {
			ger = seq.GlobalExitRoot()
		}

		batches[i] = polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData{
			Transactions:         seq.L2Data(),
			ForcedGlobalExitRoot: ger,
			ForcedTimestamp:      seq.ForcedBatchTimestamp(),
			// TODO: Check that is ok to use ForcedBlockHashL1 instead PrevBlockHash
			ForcedBlockHashL1: seq.ForcedBlockHashL1(),
		}
	}
	lastSequencedBatchNumber := getLastSequencedBatchNumber(sequences)
	tx, err := t.rollupContract.SequenceBatches(
		&opts, batches, sequences.MaxSequenceTimestamp(), lastSequencedBatchNumber, sequences.L2Coinbase(),
	)
	if err != nil {
		t.warningMessage(batches, sequences.L2Coinbase(), &opts)
		if parsedErr, ok := etherman.TryParseError(err); ok {
			err = parsedErr
		}
	}

	return tx, err
}

func (t *TxBuilderElderberryZKEVM) warningMessage(
	batches []polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData, l2Coinbase common.Address, opts *bind.TransactOpts) {
	t.logger.Warnf("Sequencer address: ", opts.From, "l2CoinBase: ", l2Coinbase, " Batches to send: %+v", batches)
}

func (t *TxBuilderElderberryZKEVM) String() string {
	return "Elderberry/ZKEVM"
}
