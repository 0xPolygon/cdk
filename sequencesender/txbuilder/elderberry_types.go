package txbuilder

import (
	"fmt"
	"log"

	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes"
	"github.com/ethereum/go-ethereum/common"
)

type ElderberrySequence struct {
	l2Coinbase common.Address
	batches    []seqsendertypes.Batch
}

func NewElderberrySequence(batches []seqsendertypes.Batch, l2Coinbase common.Address) *ElderberrySequence {
	return &ElderberrySequence{
		l2Coinbase: l2Coinbase,
		batches:    batches,
	}
}

func (b *ElderberrySequence) IndexL1InfoRoot() uint32 {
	log.Fatal("Elderberry Sequence does not have IndexL1InfoRoot")
	return 0
}

func (b *ElderberrySequence) MaxSequenceTimestamp() uint64 {
	return b.LastBatch().LastL2BLockTimestamp()
}

func (b *ElderberrySequence) L1InfoRoot() common.Hash {
	log.Fatal("Elderberry Sequence does not have L1InfoRoot")
	return common.Hash{}
}

func (b *ElderberrySequence) Batches() []seqsendertypes.Batch {
	return b.batches
}

func (b *ElderberrySequence) FirstBatch() seqsendertypes.Batch {
	return b.batches[0]
}

func (b *ElderberrySequence) LastBatch() seqsendertypes.Batch {
	return b.batches[b.Len()-1]
}

func (b *ElderberrySequence) Len() int {
	return len(b.batches)
}

func (b *ElderberrySequence) L2Coinbase() common.Address {
	return b.l2Coinbase
}
func (b *ElderberrySequence) String() string {
	res := fmt.Sprintf("Seq/Elderberry: L2Coinbase: %s, Batches: %d", b.l2Coinbase.String(), len(b.batches))
	for i, batch := range b.Batches() {
		res += fmt.Sprintf("\n\tBatch %d: %s", i, batch.String())
	}
	return res
}
