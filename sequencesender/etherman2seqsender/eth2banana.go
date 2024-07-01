package etherman2seqsender

import (
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	ethmanTypes "github.com/0xPolygon/cdk/etherman/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Eth2Banana struct {
	ZkEVM     *polygonvalidiumetrog.Polygonvalidiumetrog
	ZkEVMAddr common.Address
}

func (etherMan *Eth2Banana) BuildSequenceBatchesTxZKEVM(opts *bind.TransactOpts, sequences []ethmanTypes.Sequence, maxSequenceTimestamp uint64, lastSequencedBatchNumber uint64, l2Coinbase common.Address) (*types.Transaction, error) {
	panic("implement me")
}

func (etherMan *Eth2Banana) BuildSequenceBatchesTxValidium(opts *bind.TransactOpts, sequences []ethmanTypes.Sequence, maxSequenceTimestamp uint64, lastSequencedBatchNumber uint64, l2Coinbase common.Address) (*types.Transaction, error) {
	panic("implement me")
}
