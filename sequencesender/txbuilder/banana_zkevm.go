package txbuilder

import (
	"context"
	"math/big"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type TxBuilderBananaZKEVM struct {
	TxBuilderBananaBase
}

func NewTxBuilderBananaZKEVM(zkevm contracts.ContractRollupBanana, opts bind.TransactOpts, sender common.Address) *TxBuilderBananaZKEVM {
	return &TxBuilderBananaZKEVM{
		TxBuilderBananaBase: TxBuilderBananaBase{
			zkevm:         zkevm,
			opts:          opts,
			SenderAddress: sender,
		},
	}
}

func (t *TxBuilderBananaZKEVM) BuildSequenceBatchesTx(ctx context.Context, sender common.Address, sequences seqsendertypes.Sequence) (*ethtypes.Transaction, error) {
	var err error
	ethseq := convertToSequenceBanana(sequences)
	newopts := t.opts
	newopts.NoSend = true

	// force nonce, gas limit and gas price to avoid querying it from the chain
	newopts.Nonce = big.NewInt(1)
	newopts.GasLimit = uint64(1)
	newopts.GasPrice = big.NewInt(1)
	// Build sequence data
	tx, err := t.sequenceBatchesRollup(newopts, ethseq)
	if err != nil {
		log.Errorf("[SeqSender] error estimating new sequenceBatches to add to ethtxmanager: ", err)
		return nil, err
	}
	return tx, nil
}

func (t *TxBuilderBananaZKEVM) sequenceBatchesRollup(opts bind.TransactOpts, sequence etherman.SequenceBanana) (*ethtypes.Transaction, error) {
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

	tx, err := t.zkevm.GetContract().SequenceBatches(&opts, batches, sequence.IndexL1InfoRoot, sequence.MaxSequenceTimestamp, sequence.AccInputHash, sequence.L2Coinbase)
	if err != nil {
		log.Debugf("Batches to send: %+v", batches)
		log.Debug("l2CoinBase: ", sequence.L2Coinbase)
		log.Debug("Sequencer address: ", opts.From)

	}

	return tx, err
}

func (t *TxBuilderBananaZKEVM) String() string {
	return "Banana/ZKEVM"
}
