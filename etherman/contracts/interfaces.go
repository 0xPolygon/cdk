package contracts

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type NameType string

const (
	ContractNameDAProtocol     NameType = "daprotocol"
	ContractNameRollupManager  NameType = "rollupmanager"
	ContractNameGlobalExitRoot NameType = "globalexitroot"
	ContractNameRollup         NameType = "rollup"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

type VersionType string

const (
	VersionBanana     VersionType = "banana"
	VersionElderberry VersionType = "elderberry"
)

type RollupContractor interface {
	TrustedSequencer() (common.Address, error)
	TrustedSequencerURL() (string, error)
	LastAccInputHash() (common.Hash, error)
	DataAvailabilityProtocol() (common.Address, error)
}

type GlobalExitRootContractor interface {
	// L1InfoIndexToRoot: get the root for a specific leaf index
	L1InfoIndexToRoot(indexLeaf uint32) (common.Hash, error)
}

type BaseContractor interface {
	GetName() string
	GetVersion() string
	GetAddress() common.Address
	String() string
}

type RollupData struct {
	RollupContract                 common.Address
	ChainID                        uint64
	Verifier                       common.Address
	ForkID                         uint64
	LastLocalExitRoot              [32]byte
	LastBatchSequenced             uint64
	LastVerifiedBatch              uint64
	LastPendingState               uint64
	LastPendingStateConsolidated   uint64
	LastVerifiedBatchBeforeUpgrade uint64
	RollupTypeID                   uint64
	RollupCompatibilityID          uint8
}

type StateVariablesSequencedBatchData struct {
	AccInputHash               [32]byte
	SequencedTimestamp         uint64
	PreviousLastBatchSequenced uint64
}

type RollupManagerContractor interface {
	BaseContractor
	RollupAddressToID(rollupAddress common.Address) (uint32, error)
	GetBatchFee() (*big.Int, error)
	RollupIDToRollupData(rollupID uint32) (*RollupData, error)
	VerifyBatchesTrustedAggregator(opts *bind.TransactOpts, rollupID uint32, pendingStateNum uint64, initNumBatch uint64, finalNewBatch uint64, newLocalExitRoot [32]byte, newStateRoot [32]byte, beneficiary common.Address, proof [24][32]byte) (*types.Transaction, error)
	GetRollupSequencedBatches(rollupID uint32, batchNum uint64) (StateVariablesSequencedBatchData, error)
}
