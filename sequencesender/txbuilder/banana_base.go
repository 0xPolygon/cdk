package txbuilder

import (
	"fmt"

	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type TxBuilderBananaBase struct {
	zkevm                  contracts.ContractRollupBanana
	globalExitRootContract contracts.ContractGlobalExitRootBanana
	opts                   bind.TransactOpts
	SenderAddress          common.Address
}

func convertToSequenceBanana(sequences seqsendertypes.Sequence) (etherman.SequenceBanana, error) {
	seqEth, ok := sequences.(*BananaSequence)
	if !ok {
		log.Error("sequences is not a BananaSequence")
		return etherman.SequenceBanana{}, fmt.Errorf("sequences is not a BananaSequence")
	}
	seqEth.SequenceBanana.Batches = make([]etherman.Batch, len(sequences.Batches()))
	for _, batch := range sequences.Batches() {
		ethBatch, err := convertToEthermanBatch(batch)
		if err != nil {
			return etherman.SequenceBanana{}, err
		}
		seqEth.SequenceBanana.Batches = append(seqEth.SequenceBanana.Batches, ethBatch)
	}
	return seqEth.SequenceBanana, nil
}

func convertToEthermanBatch(batch seqsendertypes.Batch) (etherman.Batch, error) {
	return etherman.Batch{
		L2Data:               batch.L2Data(),
		LastCoinbase:         batch.LastCoinbase(),
		ForcedGlobalExitRoot: batch.ForcedGlobalExitRoot(),
		ForcedBlockHashL1:    batch.ForcedBlockHashL1(),
		ForcedBatchTimestamp: batch.ForcedBatchTimestamp(),
		BatchNumber:          batch.BatchNumber(),
		L1InfoTreeIndex:      batch.L1InfoTreeIndex(),
		LastL2BLockTimestamp: batch.LastL2BLockTimestamp(),
		GlobalExitRoot:       batch.GlobalExitRoot(),
	}, nil
	// cast, ok := batch.(*BananaBatch)
	// if !ok {
	// 	log.Error("Batch is not a BananaBatch")
	// 	return etherman.Batch{}, fmt.Errorf("Batch is not a BananaBatch")
	// }
	// return cast.Batch, nil
}

func convertToEthermanBatches(batch []seqsendertypes.Batch) ([]etherman.Batch, error) {
	result := make([]etherman.Batch, len(batch))
	for i, b := range batch {
		var err error
		result[i], err = convertToEthermanBatch(b)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (t *TxBuilderBananaBase) NewBatchFromL2Block(l2Block *datastream.L2Block) seqsendertypes.Batch {
	batch := &etherman.Batch{
		LastL2BLockTimestamp: l2Block.Timestamp,
		BatchNumber:          l2Block.BatchNumber,
		L1InfoTreeIndex:      l2Block.L1InfotreeIndex,
		LastCoinbase:         common.BytesToAddress(l2Block.Coinbase),
		GlobalExitRoot:       common.BytesToHash(l2Block.GlobalExitRoot),
	}
	return NewBananaBatch(batch)
}

func (t *TxBuilderBananaBase) NewSequence(batches []seqsendertypes.Batch, coinbase common.Address) (seqsendertypes.Sequence, error) {
	ethBatches, err := convertToEthermanBatches(batches)
	if err != nil {
		return nil, err
	}
	sequence := etherman.NewSequenceBanana(ethBatches, coinbase)

	l1InfoRoot, err := t.globalExitRootContract.GetL1InfoRoot(sequence.IndexL1InfoRoot)
	if err != nil {
		return nil, err
	}

	sequence.L1InfoRoot = l1InfoRoot

	accInputHash, err := t.zkevm.LastAccInputHash()
	if err != nil {
		return nil, err
	}

	oldAccInputHash := common.BytesToHash(accInputHash.Bytes()) //copy it

	for _, batch := range sequence.Batches {
		infoRootHash := sequence.L1InfoRoot
		timestamp := sequence.MaxSequenceTimestamp
		blockHash := common.Hash{}

		if batch.ForcedBatchTimestamp > 0 {
			infoRootHash = batch.ForcedGlobalExitRoot
			timestamp = batch.ForcedBatchTimestamp
			blockHash = batch.ForcedBlockHashL1
		}

		accInputHash, err = cdkcommon.CalculateAccInputHash(accInputHash, batch.L2Data, infoRootHash, timestamp, batch.LastCoinbase, blockHash)
		if err != nil {
			return nil, err
		}
	}

	sequence.OldAccInputHash = oldAccInputHash
	sequence.AccInputHash = accInputHash
	res := NewBananaSequence(*sequence)
	return res, nil
}
