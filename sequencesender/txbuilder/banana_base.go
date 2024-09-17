package txbuilder

import (
	"context"
	"fmt"
	"math/big"

	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/0xPolygon/cdk/state/datastream"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type rollupBananaBaseContractor interface {
	LastAccInputHash(opts *bind.CallOpts) ([32]byte, error)
}

type globalExitRootBananaContractor interface {
	L1InfoRootMap(opts *bind.CallOpts, index uint32) ([32]byte, error)
	String() string
}

type l1InfoSyncer interface {
	GetLatestInfoUntilBlock(ctx context.Context, blockNum uint64) (*l1infotreesync.L1InfoTreeLeaf, error)
}

type l1Client interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
}

type TxBuilderBananaBase struct {
	logger                 *log.Logger
	rollupContract         rollupBananaBaseContractor
	globalExitRootContract globalExitRootBananaContractor
	l1InfoTree             l1InfoSyncer
	ethClient              l1Client
	blockFinality          *big.Int
	opts                   bind.TransactOpts
}

func NewTxBuilderBananaBase(
	logger *log.Logger,
	rollupContract rollupBananaBaseContractor,
	gerContract globalExitRootBananaContractor,
	l1InfoTree l1InfoSyncer,
	ethClient l1Client,
	blockFinality *big.Int,
	opts bind.TransactOpts,
) *TxBuilderBananaBase {
	return &TxBuilderBananaBase{
		logger:                 logger,
		rollupContract:         rollupContract,
		globalExitRootContract: gerContract,
		l1InfoTree:             l1InfoTree,
		ethClient:              ethClient,
		blockFinality:          blockFinality,
		opts:                   opts,
	}
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

func (t *TxBuilderBananaBase) NewSequence(
	ctx context.Context, batches []seqsendertypes.Batch, coinbase common.Address,
) (seqsendertypes.Sequence, error) {
	ethBatches := toEthermanBatches(batches)
	sequence := etherman.NewSequenceBanana(ethBatches, coinbase)
	var greatestL1Index uint32
	for _, b := range sequence.Batches {
		if greatestL1Index < b.L1InfoTreeIndex {
			greatestL1Index = b.L1InfoTreeIndex
		}
	}
	header, err := t.ethClient.HeaderByNumber(ctx, t.blockFinality)
	if err != nil {
		return nil, fmt.Errorf("error calling HeaderByNumber, with block finality %d: %w", t.blockFinality.Int64(), err)
	}
	info, err := t.l1InfoTree.GetLatestInfoUntilBlock(ctx, header.Number.Uint64())
	if err != nil {
		return nil, fmt.Errorf("error calling GetLatestInfoUntilBlock with block num %d: %w", header.Number.Uint64(), err)
	}
	if info.L1InfoTreeIndex >= greatestL1Index {
		sequence.CounterL1InfoRoot = info.L1InfoTreeIndex + 1
	} else {
		return nil, fmt.Errorf(
			"sequence contained an L1 Info tree index (%d) that is greater than the one synced with the desired finality (%d)",
			greatestL1Index, info.L1InfoTreeIndex,
		)
	}

	l1InfoRoot, err := t.getL1InfoRoot(sequence.CounterL1InfoRoot)
	if err != nil {
		return nil, err
	}

	sequence.L1InfoRoot = l1InfoRoot

	accInputHash, err := t.rollupContract.LastAccInputHash(&bind.CallOpts{Pending: false})
	if err != nil {
		return nil, err
	}

	oldAccInputHash := common.BytesToHash(accInputHash[:]) // copy it

	for _, batch := range sequence.Batches {
		infoRootHash := sequence.L1InfoRoot
		timestamp := sequence.MaxSequenceTimestamp
		blockHash := common.Hash{}

		if batch.ForcedBatchTimestamp > 0 {
			infoRootHash = batch.ForcedGlobalExitRoot
			timestamp = batch.ForcedBatchTimestamp
			blockHash = batch.ForcedBlockHashL1
		}

		accInputHash = cdkcommon.CalculateAccInputHash(
			t.logger, accInputHash, batch.L2Data, infoRootHash, timestamp, batch.LastCoinbase, blockHash,
		)
	}

	sequence.OldAccInputHash = oldAccInputHash
	sequence.AccInputHash = accInputHash
	res := NewBananaSequence(*sequence)
	return res, nil
}

func (t *TxBuilderBananaBase) getL1InfoRoot(counterL1InfoRoot uint32) (common.Hash, error) {
	return t.globalExitRootContract.L1InfoRootMap(&bind.CallOpts{Pending: false}, counterL1InfoRoot)
}

func convertToSequenceBanana(sequences seqsendertypes.Sequence) (etherman.SequenceBanana, error) {
	seqEth, ok := sequences.(*BananaSequence)
	if !ok {
		log.Error("sequences is not a BananaSequence")
		return etherman.SequenceBanana{}, fmt.Errorf("sequences is not a BananaSequence")
	}

	ethermanSequence := etherman.SequenceBanana{
		OldAccInputHash:      seqEth.SequenceBanana.OldAccInputHash,
		AccInputHash:         seqEth.SequenceBanana.AccInputHash,
		L1InfoRoot:           seqEth.SequenceBanana.L1InfoRoot,
		MaxSequenceTimestamp: seqEth.SequenceBanana.MaxSequenceTimestamp,
		CounterL1InfoRoot:    seqEth.SequenceBanana.CounterL1InfoRoot,
		L2Coinbase:           seqEth.SequenceBanana.L2Coinbase,
	}

	for _, batch := range sequences.Batches() {
		ethBatch := toEthermanBatch(batch)
		ethermanSequence.Batches = append(ethermanSequence.Batches, ethBatch)
	}

	return ethermanSequence, nil
}

func toEthermanBatch(batch seqsendertypes.Batch) etherman.Batch {
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
	}
}

func toEthermanBatches(batch []seqsendertypes.Batch) []etherman.Batch {
	result := make([]etherman.Batch, len(batch))
	for i, b := range batch {
		result[i] = toEthermanBatch(b)
	}

	return result
}
