package txbuilder

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk/dataavailability"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type TxBuilderBananaValidium struct {
	TxBuilderBananaBase
	da             dataavailability.SequenceSenderBanana
	condNewSeq     CondNewSequence
	rollupContract rollupBananaValidiumContractor
}

type rollupBananaValidiumContractor interface {
	rollupBananaBaseContractor
	SequenceBatchesValidium(opts *bind.TransactOpts, batches []polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData, indexL1InfoRoot uint32, maxSequenceTimestamp uint64, expectedFinalAccInputHash [32]byte, l2Coinbase common.Address, dataAvailabilityMessage []byte) (*types.Transaction, error)
}

func NewTxBuilderBananaValidium(
	rollupContract rollupBananaValidiumContractor,
	gerContract globalExitRootBananaContractor,
	da dataavailability.SequenceSenderBanana, opts bind.TransactOpts, maxBatchesForL1 uint64,
	l1InfoTree l1InfoSyncer,
	ethClient l1Client,
	blockFinality *big.Int,
) *TxBuilderBananaValidium {
	return &TxBuilderBananaValidium{
		TxBuilderBananaBase: *NewTxBuilderBananaBase(rollupContract, gerContract, l1InfoTree, ethClient, blockFinality, opts),
		da:                  da,
		condNewSeq:          NewConditionalNewSequenceNumBatches(maxBatchesForL1),
		rollupContract:      rollupContract,
	}
}

func (t *TxBuilderBananaValidium) NewSequenceIfWorthToSend(ctx context.Context, sequenceBatches []seqsendertypes.Batch, l2Coinbase common.Address, batchNumber uint64) (seqsendertypes.Sequence, error) {
	return t.condNewSeq.NewSequenceIfWorthToSend(ctx, t, sequenceBatches, l2Coinbase)
}

// SetCondNewSeq allow to override the default conditional for new sequence
func (t *TxBuilderBananaValidium) SetCondNewSeq(cond CondNewSequence) CondNewSequence {
	previous := t.condNewSeq
	t.condNewSeq = cond
	return previous
}

func (t *TxBuilderBananaValidium) BuildSequenceBatchesTx(ctx context.Context, sequences seqsendertypes.Sequence) (*types.Transaction, error) {
	// TODO: param sender
	// Post sequences to DA backend
	var dataAvailabilityMessage []byte
	var err error
	ethseq, err := convertToSequenceBanana(sequences)
	if err != nil {
		log.Error("error converting sequences to etherman: ", err)
		return nil, err
	}

	dataAvailabilityMessage, err = t.da.PostSequenceBanana(ctx, ethseq)
	if err != nil {
		log.Error("error posting sequences to the data availability protocol: ", err)
		return nil, err
	}
	if dataAvailabilityMessage == nil {
		err := fmt.Errorf("data availability message is nil")
		log.Error("error posting sequences to the data availability protocol: ", err.Error())
		return nil, err
	}

	// Build sequence data
	tx, err := t.internalBuildSequenceBatchesTx(ethseq, dataAvailabilityMessage)
	if err != nil {
		log.Errorf("[SeqSender] error estimating new sequenceBatches to add to ethtxmanager: ", err)
		return nil, err
	}
	return tx, nil
}

// BuildSequenceBatchesTx builds a tx to be sent to the PoE SC method SequenceBatches.
func (t *TxBuilderBananaValidium) internalBuildSequenceBatchesTx(sequence etherman.SequenceBanana,
	dataAvailabilityMessage []byte) (*types.Transaction, error) {
	newopts := t.opts
	newopts.NoSend = true

	// force nonce, gas limit and gas price to avoid querying it from the chain
	newopts.Nonce = big.NewInt(1)
	newopts.GasLimit = uint64(1)
	newopts.GasPrice = big.NewInt(1)

	return t.sequenceBatchesValidium(newopts, sequence, dataAvailabilityMessage)
}

func (t *TxBuilderBananaValidium) sequenceBatchesValidium(opts bind.TransactOpts, sequence etherman.SequenceBanana, dataAvailabilityMessage []byte) (*types.Transaction, error) {
	batches := make([]polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData, len(sequence.Batches))
	for i, batch := range sequence.Batches {
		var ger common.Hash
		if batch.ForcedBatchTimestamp > 0 {
			ger = batch.ForcedGlobalExitRoot
		}

		batches[i] = polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData{
			TransactionsHash:     crypto.Keccak256Hash(batch.L2Data),
			ForcedGlobalExitRoot: ger,
			ForcedTimestamp:      batch.ForcedBatchTimestamp,
			ForcedBlockHashL1:    batch.ForcedBlockHashL1,
		}
	}

	log.Infof("building banana sequence tx. AccInputHash: %s", sequence.AccInputHash.Hex())
	tx, err := t.rollupContract.SequenceBatchesValidium(&opts, batches, sequence.CounterL1InfoRoot, sequence.MaxSequenceTimestamp, sequence.AccInputHash, sequence.L2Coinbase, dataAvailabilityMessage)
	if err != nil {
		log.Debugf("Batches to send: %+v", batches)
		log.Debug("l2CoinBase: ", sequence.L2Coinbase)
		log.Debug("Sequencer address: ", opts.From)

	}

	return tx, err
}

func (t *TxBuilderBananaValidium) String() string {
	return "Banana/Validium"
}
