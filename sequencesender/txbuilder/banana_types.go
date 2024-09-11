package txbuilder

import (
	"fmt"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/ethereum/go-ethereum/common"
)

type BananaBatch struct {
	etherman.Batch
}

type BananaSequence struct {
	etherman.SequenceBanana
}

func NewBananaBatch(batch *etherman.Batch) *BananaBatch {
	return &BananaBatch{*batch}
}

func NewBananaSequence(ult etherman.SequenceBanana) *BananaSequence {
	return &BananaSequence{ult}
}

func (b *BananaSequence) IndexL1InfoRoot() uint32 {
	return b.SequenceBanana.CounterL1InfoRoot
}

func (b *BananaSequence) MaxSequenceTimestamp() uint64 {
	return b.SequenceBanana.MaxSequenceTimestamp
}

func (b *BananaSequence) L1InfoRoot() common.Hash {
	return b.SequenceBanana.L1InfoRoot
}

func (b *BananaSequence) Batches() []seqsendertypes.Batch {
	res := make([]seqsendertypes.Batch, len(b.SequenceBanana.Batches))
	for i, batch := range b.SequenceBanana.Batches {
		res[i] = &BananaBatch{batch}
	}
	return res
}

func (b *BananaSequence) FirstBatch() seqsendertypes.Batch {
	return &BananaBatch{b.SequenceBanana.Batches[0]}
}

func (b *BananaSequence) LastBatch() seqsendertypes.Batch {
	return &BananaBatch{b.SequenceBanana.Batches[b.Len()-1]}
}

func (b *BananaSequence) Len() int {
	return len(b.SequenceBanana.Batches)
}

func (b *BananaSequence) String() string {
	res := fmt.Sprintf(
		"Seq/Banana: L2Coinbase: %s, OldAccInputHash: %x, AccInputHash: %x, L1InfoRoot: %x, "+
			"MaxSequenceTimestamp: %d, IndexL1InfoRoot: %d",
		b.L2Coinbase().String(), b.OldAccInputHash.String(), b.AccInputHash.String(), b.L1InfoRoot().String(),
		b.MaxSequenceTimestamp(), b.IndexL1InfoRoot(),
	)

	for i, batch := range b.Batches() {
		res += fmt.Sprintf("\n\tBatch %d: %s", i, batch.String())
	}
	return res
}

func (b *BananaSequence) L2Coinbase() common.Address {
	return b.SequenceBanana.L2Coinbase
}

func (b *BananaBatch) LastCoinbase() common.Address {
	return b.Batch.LastCoinbase
}

func (b *BananaBatch) ForcedBatchTimestamp() uint64 {
	return b.Batch.ForcedBatchTimestamp
}

func (b *BananaBatch) ForcedGlobalExitRoot() common.Hash {
	return b.Batch.ForcedGlobalExitRoot
}

func (b *BananaBatch) ForcedBlockHashL1() common.Hash {
	return b.Batch.ForcedBlockHashL1
}

func (b *BananaBatch) L2Data() []byte {
	return b.Batch.L2Data
}

func (b *BananaBatch) LastL2BLockTimestamp() uint64 {
	return b.Batch.LastL2BLockTimestamp
}

func (b *BananaBatch) BatchNumber() uint64 {
	return b.Batch.BatchNumber
}

func (b BananaBatch) DeepCopy() seqsendertypes.Batch {
	return &BananaBatch{b.Batch}
}

func (b *BananaBatch) SetL2Data(data []byte) {
	b.Batch.L2Data = data
}

func (b *BananaBatch) SetLastCoinbase(address common.Address) {
	b.Batch.LastCoinbase = address
}

func (b *BananaBatch) SetLastL2BLockTimestamp(ts uint64) {
	b.Batch.LastL2BLockTimestamp = ts
}

func (b *BananaBatch) SetL1InfoTreeIndex(index uint32) {
	b.Batch.L1InfoTreeIndex = index
}

func (b *BananaBatch) GlobalExitRoot() common.Hash {
	return b.Batch.GlobalExitRoot
}

func (b *BananaBatch) L1InfoTreeIndex() uint32 {
	return b.Batch.L1InfoTreeIndex
}

func (b *BananaBatch) String() string {
	return fmt.Sprintf("Batch/Banana: LastCoinbase: %s, ForcedBatchTimestamp: %d, ForcedGlobalExitRoot: %x, "+
		"ForcedBlockHashL1: %x, L2Data: %x, LastL2BLockTimestamp: %d, BatchNumber: %d, "+
		"GlobalExitRoot: %x, L1InfoTreeIndex: %d",
		b.LastCoinbase().String(), b.ForcedBatchTimestamp(), b.ForcedGlobalExitRoot().String(),
		b.ForcedBlockHashL1().String(), b.L2Data(), b.LastL2BLockTimestamp(), b.BatchNumber(),
		b.GlobalExitRoot().String(), b.L1InfoTreeIndex(),
	)
}

func (b *BananaSequence) LastVirtualBatchNumber() uint64 {
	return b.SequenceBanana.LastVirtualBatchNumber
}

func (b *BananaSequence) SetLastVirtualBatchNumber(batchNumber uint64) {
	b.SequenceBanana.LastVirtualBatchNumber = batchNumber
}
