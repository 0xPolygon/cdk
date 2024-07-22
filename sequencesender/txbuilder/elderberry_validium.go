package txbuilder

import (
	"context"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk/dataavailability"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type TxBuilderElderberryValidium struct {
	TxBuilderElderberryBase
	da         dataavailability.SequenceSenderElderberry
	condNewSeq CondNewSequence
}

func NewTxBuilderElderberryValidium(zkevm contracts.RollupElderberryType,
	da dataavailability.SequenceSenderElderberry,
	opts bind.TransactOpts, sender common.Address, maxBatchesForL1 uint64) *TxBuilderElderberryValidium {
	return &TxBuilderElderberryValidium{
		da: da,
		TxBuilderElderberryBase: *NewTxBuilderElderberryBase(
			zkevm, opts,
		),
		condNewSeq: &NewSequenceConditionalNumBatches{
			maxBatchesForL1: maxBatchesForL1,
		},
	}
}
func (t *TxBuilderElderberryValidium) NewSequenceIfWorthToSend(ctx context.Context, sequenceBatches []seqsendertypes.Batch, l2Coinbase common.Address, batchNumber uint64) (seqsendertypes.Sequence, error) {
	return t.condNewSeq.NewSequenceIfWorthToSend(ctx, t, sequenceBatches, t.opts.From, l2Coinbase, batchNumber)
}

func (t *TxBuilderElderberryValidium) BuildSequenceBatchesTx(ctx context.Context, sender common.Address, sequences seqsendertypes.Sequence) (*ethtypes.Transaction, error) {

	batchesData := convertToBatchesData(sequences)
	dataAvailabilityMessage, err := t.da.PostSequenceElderberry(ctx, batchesData)
	if err != nil {
		log.Error("error posting sequences to the data availability protocol: ", err)
		return nil, err
	}
	return t.buildSequenceBatchesTxValidium(sequences, dataAvailabilityMessage)
}

func (t *TxBuilderElderberryValidium) buildSequenceBatchesTxValidium(
	sequences seqsendertypes.Sequence, dataAvailabilityMessage []byte) (*types.Transaction, error) {
	batches := make([]polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData, sequences.Len())
	for i, seq := range sequences.Batches() {
		var ger common.Hash
		if seq.ForcedBatchTimestamp() > 0 {
			ger = seq.GlobalExitRoot()
		}
		batches[i] = polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData{
			TransactionsHash:     crypto.Keccak256Hash(seq.L2Data()),
			ForcedGlobalExitRoot: ger,
			ForcedTimestamp:      uint64(seq.ForcedBatchTimestamp()),
			ForcedBlockHashL1:    seq.ForcedBlockHashL1(),
		}
	}
	lastSequencedBatchNumber := getLastSequencedBatchNumber(sequences)
	ZkEVM := t.rollupContract.Contract()
	tx, err := ZkEVM.SequenceBatchesValidium(&t.opts, batches, sequences.MaxSequenceTimestamp(),
		lastSequencedBatchNumber, sequences.L2Coinbase(), dataAvailabilityMessage)
	if err != nil {
		if parsedErr, ok := etherman.TryParseError(err); ok {
			err = parsedErr
		}
	}

	return tx, err
}

func (t *TxBuilderElderberryValidium) String() string {
	return "Elderberry/Validium"
}

func convertToBatchesData(sequences seqsendertypes.Sequence) [][]byte {
	batches := make([][]byte, sequences.Len())
	for i, batch := range sequences.Batches() {
		batches[i] = batch.L2Data()
	}
	return batches
}
