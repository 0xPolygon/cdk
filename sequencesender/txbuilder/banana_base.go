package txbuilder

import (
	"context"
	"errors"
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
	GetInitL1InfoRootMap(ctx context.Context) (*l1infotreesync.L1InfoTreeInitial, error)
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

func getHighestL1InfoIndex(batches []etherman.Batch) uint32 {
	var highestL1Index uint32
	for _, b := range batches {
		if highestL1Index < b.L1InfoTreeIndex {
			highestL1Index = b.L1InfoTreeIndex
		}
	}
	return highestL1Index
}

// Returns CounterL1InfoRoot to use for this batch
func (t *TxBuilderBananaBase) GetCounterL1InfoRoot(ctx context.Context, highestL1IndexInBatch uint32) (uint32, error) {
	header, err := t.ethClient.HeaderByNumber(ctx, t.blockFinality)
	if err != nil {
		return 0, fmt.Errorf("error calling HeaderByNumber, with block finality %d: %w", t.blockFinality.Int64(), err)
	}
	var resL1InfoCounter uint32

	info, err := t.l1InfoTree.GetLatestInfoUntilBlock(ctx, header.Number.Uint64())
	if err == nil {
		resL1InfoCounter = info.L1InfoTreeIndex + 1
	}
	if errors.Is(err, l1infotreesync.ErrNotFound) {
		// There are no L1 Info tree leaves yet, so we can try to use L1InfoRootMap event
		l1infotreeInitial, err := t.l1InfoTree.GetInitL1InfoRootMap(ctx)
		if err != nil {
			return 0, fmt.Errorf("error no leaves on L1InfoTree yet and GetInitL1InfoRootMap fails: %w", err)
		}
		// We use this leaf as first one
		resL1InfoCounter = l1infotreeInitial.LeafCount
	} else if err != nil {
		return 0, fmt.Errorf("error calling GetLatestInfoUntilBlock with block num %d: %w", header.Number.Uint64(), err)
	}
	// This is a very rare case, but it can happen if there are no leaves in L1InfoTree yet, so the batch can use any of them and set 0
	if resL1InfoCounter == 0 && highestL1IndexInBatch == 0 {
		log.Infof("No L1 Info tree leaves yet, batch don't use any leaf (index=0), so we use CounterL1InfoRoot=0 that is the empty tree")
		return resL1InfoCounter, nil
	}
	if resL1InfoCounter > highestL1IndexInBatch {
		return resL1InfoCounter, nil
	}

	return 0, fmt.Errorf(
		"sequence contained an L1 Info tree index (%d) that is greater than the one synced with the desired finality (%d)",
		highestL1IndexInBatch, resL1InfoCounter,
	)
}

func (t *TxBuilderBananaBase) NewSequence(
	ctx context.Context, batches []seqsendertypes.Batch, coinbase common.Address,
) (seqsendertypes.Sequence, error) {
	ethBatches := toEthermanBatches(batches)
	sequence := etherman.NewSequenceBanana(ethBatches, coinbase)
	greatestL1Index := getHighestL1InfoIndex(sequence.Batches)

	counterL1InfoRoot, err := t.GetCounterL1InfoRoot(ctx, greatestL1Index)
	if err != nil {
		return nil, err
	}
	sequence.CounterL1InfoRoot = counterL1InfoRoot
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

	err = SequenceSanityCheck(sequence)
	if err != nil {
		return nil, fmt.Errorf("sequenceSanityCheck fails. Err: %w", err)
	}
	res := NewBananaSequence(*sequence)
	return res, nil
}

func SequenceSanityCheck(seq *etherman.SequenceBanana) error {
	maxL1InfoIndex, err := calculateMaxL1InfoTreeIndexInsideSequence(seq)
	if err != nil {
		return err
	}
	if seq.CounterL1InfoRoot < maxL1InfoIndex+1 {
		return fmt.Errorf("wrong CounterL1InfoRoot(%d): BatchL2Data (max=%d) ", seq.CounterL1InfoRoot, maxL1InfoIndex)
	}
	return nil
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
