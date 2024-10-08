package txbuilder

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk/dataavailability"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type TxBuilderElderberryValidium struct {
	TxBuilderElderberryBase
	da             dataavailability.SequenceSenderElderberry
	condNewSeq     CondNewSequence
	rollupContract rollupElderberryValidiumContractor
}

type rollupElderberryValidiumContractor interface {
	SequenceBatchesValidium(
		opts *bind.TransactOpts,
		batches []polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData,
		maxSequenceTimestamp uint64,
		initSequencedBatch uint64,
		l2Coinbase common.Address,
		dataAvailabilityMessage []byte,
	) (*types.Transaction, error)
}

func NewTxBuilderElderberryValidium(
	logger *log.Logger,
	zkevm contracts.RollupElderberryType,
	da dataavailability.SequenceSenderElderberry,
	opts bind.TransactOpts, maxBatchesForL1 uint64) *TxBuilderElderberryValidium {
	return &TxBuilderElderberryValidium{
		da:                      da,
		TxBuilderElderberryBase: *NewTxBuilderElderberryBase(logger, opts),
		condNewSeq:              NewConditionalNewSequenceNumBatches(maxBatchesForL1),
		rollupContract:          zkevm,
	}
}
func (t *TxBuilderElderberryValidium) NewSequenceIfWorthToSend(
	ctx context.Context, sequenceBatches []seqsendertypes.Batch, l2Coinbase common.Address, batchNumber uint64,
) (seqsendertypes.Sequence, error) {
	return t.condNewSeq.NewSequenceIfWorthToSend(ctx, t, sequenceBatches, l2Coinbase)
}

// SetCondNewSeq allow to override the default conditional for new sequence
func (t *TxBuilderElderberryValidium) SetCondNewSeq(cond CondNewSequence) CondNewSequence {
	previous := t.condNewSeq
	t.condNewSeq = cond
	return previous
}

func (t *TxBuilderElderberryValidium) BuildSequenceBatchesTx(
	ctx context.Context, sequences seqsendertypes.Sequence,
) (*types.Transaction, error) {
	if sequences == nil || sequences.Len() == 0 {
		return nil, fmt.Errorf("can't sequence an empty sequence")
	}
	batchesData := convertToBatchesData(sequences)
	dataAvailabilityMessage, err := t.da.PostSequenceElderberry(ctx, batchesData)
	if err != nil {
		t.logger.Error("error posting sequences to the data availability protocol: ", err)
		return nil, err
	}
	if dataAvailabilityMessage == nil {
		err := fmt.Errorf("data availability message is nil")
		t.logger.Error("error posting sequences to the data availability protocol: ", err.Error())
		return nil, err
	}
	newopts := t.opts
	newopts.NoSend = true

	// force nonce, gas limit and gas price to avoid querying it from the chain
	newopts.Nonce = big.NewInt(1)
	newopts.GasLimit = uint64(1)
	newopts.GasPrice = big.NewInt(1)

	return t.buildSequenceBatchesTxValidium(&newopts, sequences, dataAvailabilityMessage)
}

func (t *TxBuilderElderberryValidium) buildSequenceBatchesTxValidium(opts *bind.TransactOpts,
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
			ForcedTimestamp:      seq.ForcedBatchTimestamp(),
			ForcedBlockHashL1:    seq.ForcedBlockHashL1(),
		}
	}
	lastSequencedBatchNumber := getLastSequencedBatchNumber(sequences)
	t.logger.Infof("SequenceBatchesValidium(from=%s, len(batches)=%d, MaxSequenceTimestamp=%d, "+
		"lastSequencedBatchNumber=%d, L2Coinbase=%s, dataAvailabilityMessage=%s)",
		t.opts.From.String(), len(batches), sequences.MaxSequenceTimestamp(), lastSequencedBatchNumber,
		sequences.L2Coinbase().String(), hex.EncodeToString(dataAvailabilityMessage),
	)
	tx, err := t.rollupContract.SequenceBatchesValidium(opts, batches, sequences.MaxSequenceTimestamp(),
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
