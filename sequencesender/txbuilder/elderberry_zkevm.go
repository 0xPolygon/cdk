package txbuilder

import (
	"context"
	"math/big"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type TxBuilderElderberryZKEVM struct {
	opts       bind.TransactOpts
	zkevm      contracts.ContractRollupElderberry
	condNewSeq CondNewSequence
}

func NewTxBuilderElderberryZKEVM(zkevm contracts.ContractRollupElderberry, opts bind.TransactOpts, sender common.Address, maxTxSizeForL1 uint64) *TxBuilderElderberryZKEVM {
	return &TxBuilderElderberryZKEVM{
		opts:  opts,
		zkevm: zkevm,
		condNewSeq: &NewSequenceConditionalMaxSize{
			maxTxSizeForL1: maxTxSizeForL1,
		},
	}
}

func (t *TxBuilderElderberryZKEVM) NewSequence(batches []seqsendertypes.Batch, coinbase common.Address) (seqsendertypes.Sequence, error) {
	seq := ElderberrySequence{
		l2Coinbase: coinbase,
		batches:    batches,
	}
	return &seq, nil
}

func (t *TxBuilderElderberryZKEVM) NewSequenceIfWorthToSend(ctx context.Context, sequenceBatches []seqsendertypes.Batch, l2Coinbase common.Address, batchNumber uint64) (seqsendertypes.Sequence, error) {
	return t.condNewSeq.NewSequenceIfWorthToSend(ctx, t, sequenceBatches, t.opts.From, l2Coinbase, batchNumber)
}

func (t *TxBuilderElderberryZKEVM) NewBatchFromL2Block(l2Block *datastream.L2Block) seqsendertypes.Batch {
	batch := &etherman.Batch{
		LastL2BLockTimestamp: l2Block.Timestamp,
		BatchNumber:          l2Block.BatchNumber,
		L1InfoTreeIndex:      l2Block.L1InfotreeIndex,
		LastCoinbase:         common.BytesToAddress(l2Block.Coinbase),
		GlobalExitRoot:       common.BytesToHash(l2Block.GlobalExitRoot),
	}
	return NewBananaBatch(batch)
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
	ZkEVM := t.zkevm.GetContract()
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
