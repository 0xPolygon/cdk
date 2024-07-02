package etherman2seqsender

import (
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk/etherman2"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type Eth2Elderberry struct {
	ZkEVM etherman2.Etherman2ContractRollupElderberry
}

func NewEth2Elderberry(zkEVM etherman2.Etherman2ContractRollupElderberry) *Eth2Elderberry {
	return &Eth2Elderberry{
		ZkEVM: zkEVM,
	}
}

// func (etherMan *Eth2Elderberry) BuildSequenceBatchesTxZKEVM(opts bind.TransactOpts, sequences []ethmanTypes.Sequence, maxSequenceTimestamp uint64, lastSequencedBatchNumber uint64, l2Coinbase common.Address) (*types.Transaction, error) {
func (etherMan *Eth2Elderberry) BuildSequenceBatchesTxZKEVM(opts *bind.TransactOpts, params BuildSequenceBatchesParams) (*types.Transaction, error) {
	batches := make([]polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData, len(params.Sequences))
	for i, seq := range params.Sequences {
		var ger common.Hash
		if seq.ForcedBatchTimestamp > 0 {
			ger = seq.GlobalExitRoot
		}

		batches[i] = polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData{
			Transactions:         seq.BatchL2Data,
			ForcedGlobalExitRoot: ger,
			ForcedTimestamp:      uint64(seq.ForcedBatchTimestamp),
			ForcedBlockHashL1:    seq.PrevBlockHash,
		}
	}
	ZkEVM := etherMan.ZkEVM.GetContract()
	tx, err := ZkEVM.SequenceBatches(opts, batches, params.MaxSequenceTimestamp, params.LastSequencedBatchNumber, params.L2Coinbase)
	if err != nil {
		etherMan.warningMessage(batches, params.L2Coinbase, opts)
		if parsedErr, ok := etherman2.TryParseError(err); ok {
			err = parsedErr
		}
	}

	return tx, err
}

func (etherMan *Eth2Elderberry) BuildSequenceBatchesTxValidium(opts *bind.TransactOpts,
	params BuildSequenceBatchesParams, dataAvailabilityMessage DataAvailabilityMessageType) (*types.Transaction, error) {
	batches := make([]polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData, len(params.Sequences))
	for i, seq := range params.Sequences {
		var ger common.Hash
		if seq.ForcedBatchTimestamp > 0 {
			ger = seq.GlobalExitRoot
		}
		batches[i] = polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData{
			TransactionsHash:     crypto.Keccak256Hash(seq.BatchL2Data),
			ForcedGlobalExitRoot: ger,
			ForcedTimestamp:      uint64(seq.ForcedBatchTimestamp),
			ForcedBlockHashL1:    seq.PrevBlockHash,
		}
	}
	ZkEVM := etherMan.ZkEVM.GetContract()
	tx, err := ZkEVM.SequenceBatchesValidium(opts, batches, params.MaxSequenceTimestamp,
		params.LastSequencedBatchNumber, params.L2Coinbase, dataAvailabilityMessage)
	if err != nil {
		if parsedErr, ok := etherman2.TryParseError(err); ok {
			err = parsedErr
		}
	}

	return tx, err
}

func (etherMan *Eth2Elderberry) warningMessage(batches []polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData, l2Coinbase common.Address, opts *bind.TransactOpts) {
	log.Debugf("Batches to send: %+v", batches)
	log.Debug("l2CoinBase: ", l2Coinbase)
	log.Debug("Sequencer address: ", opts.From)
}
