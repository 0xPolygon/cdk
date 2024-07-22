package txbuilder

import (
	"context"
	"math/big"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type TxBuilderElderberryZKEVM struct {
	TxBuilderElderberryBase
	condNewSeq CondNewSequence
}

func NewTxBuilderElderberryZKEVM(zkevm contracts.RollupElderberryType, opts bind.TransactOpts, sender common.Address, maxTxSizeForL1 uint64) *TxBuilderElderberryZKEVM {
	return &TxBuilderElderberryZKEVM{
		TxBuilderElderberryBase: *NewTxBuilderElderberryBase(
			zkevm, opts,
		),
		condNewSeq: &NewSequenceConditionalMaxSize{
			maxTxSizeForL1: maxTxSizeForL1,
		},
	}
}

func (t *TxBuilderElderberryZKEVM) BuildSequenceBatchesTx(ctx context.Context, sender common.Address, sequences seqsendertypes.Sequence) (*ethtypes.Transaction, error) {

	newopts := t.opts
	newopts.NoSend = true

	// force nonce, gas limit and gas price to avoid querying it from the chain
	newopts.Nonce = big.NewInt(1)
	newopts.GasLimit = uint64(1)
	newopts.GasPrice = big.NewInt(1)

	return t.sequenceBatchesRollup(newopts, sequences)
}

func (t *TxBuilderElderberryZKEVM) sequenceBatchesRollup(opts bind.TransactOpts, sequences seqsendertypes.Sequence) (*types.Transaction, error) {
	batches := make([]polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData, sequences.Len())
	for i, seq := range sequences.Batches() {
		var ger common.Hash
		if seq.ForcedBatchTimestamp() > 0 {
			ger = seq.GlobalExitRoot()
		}

		batches[i] = polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData{
			Transactions:         seq.L2Data(),
			ForcedGlobalExitRoot: ger,
			ForcedTimestamp:      uint64(seq.ForcedBatchTimestamp()),
			// TODO: Check that is ok to use ForcedBlockHashL1 instead PrevBlockHash
			ForcedBlockHashL1: seq.ForcedBlockHashL1(),
		}
	}
	lastSequencedBatchNumber := getLastSequencedBatchNumber(sequences)
	ZkEVM := t.rollupContract.Contract()
	tx, err := ZkEVM.SequenceBatches(&opts, batches, sequences.MaxSequenceTimestamp(), lastSequencedBatchNumber, sequences.L2Coinbase())
	if err != nil {
		t.warningMessage(batches, sequences.L2Coinbase(), &opts)
		if parsedErr, ok := etherman.TryParseError(err); ok {
			err = parsedErr
		}
	}

	return tx, err
}

func (t *TxBuilderElderberryZKEVM) warningMessage(batches []polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData, l2Coinbase common.Address, opts *bind.TransactOpts) {
	log.Warnf("Sequencer address: ", opts.From, "l2CoinBase: ", l2Coinbase, " Batches to send: %+v", batches)
}

func getLastSequencedBatchNumber(sequences seqsendertypes.Sequence) uint64 {
	if sequences.Len() == 0 {
		return 0
	}
	return sequences.FirstBatch().BatchNumber() - 1
}

func (t *TxBuilderElderberryZKEVM) String() string {
	return "Elderberry/ZKEVM"
}
