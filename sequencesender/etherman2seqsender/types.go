package etherman2seqsender

import (
	ethmanTypes "github.com/0xPolygon/cdk/etherman/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type BuildSequenceBatchesParams struct {
	Sequences                []ethmanTypes.Sequence
	MaxSequenceTimestamp     uint64
	LastSequencedBatchNumber uint64
	L2Coinbase               common.Address
}

type DataAvailabilityMessageType = []byte

type Eth2SequenceBatchesTxZKEVMBuilder interface {
	BuildSequenceBatchesTxZKEVM(auth *bind.TransactOpts, params BuildSequenceBatchesParams) (*types.Transaction, error)
}

type Eth2SequenceBatchesTxValidiumBuilder interface {
	BuildSequenceBatchesTxValidium(auth *bind.TransactOpts, params BuildSequenceBatchesParams, dataAvailabilityMessage DataAvailabilityMessageType) (*types.Transaction, error)
}
