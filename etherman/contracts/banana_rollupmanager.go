package contracts

import (
	"math/big"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonrollupmanager"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type ContractRollupManangerBanana struct {
	*ContractBase[polygonrollupmanager.Polygonrollupmanager]
}

func NewContractRollupManangerBanana(address common.Address, backend bind.ContractBackend) (*ContractRollupManangerBanana, error) {
	base, err := NewContractBase(polygonrollupmanager.NewPolygonrollupmanager, address, backend, ContractNameRollupManager, VersionBanana)
	if err != nil {
		return nil, err
	}
	return &ContractRollupManangerBanana{
		ContractBase: base,
	}, nil
}

func (e *ContractRollupManangerBanana) RollupAddressToID(rollupAddress common.Address) (uint32, error) {
	return e.GetContract().RollupAddressToID(&bind.CallOpts{Pending: false}, rollupAddress)
}

func (e *ContractRollupManangerBanana) RollupIDToRollupData(rollupID uint32) (*RollupData, error) {
	rollupData, err := e.GetContract().RollupIDToRollupData(&bind.CallOpts{Pending: false}, rollupID)
	if err != nil {
		return nil, err
	}
	return &RollupData{
		RollupContract:                 rollupData.RollupContract,
		ChainID:                        rollupData.ChainID,
		Verifier:                       rollupData.Verifier,
		ForkID:                         rollupData.ForkID,
		LastLocalExitRoot:              rollupData.LastLocalExitRoot,
		LastBatchSequenced:             rollupData.LastBatchSequenced,
		LastVerifiedBatch:              rollupData.LastVerifiedBatch,
		LastPendingState:               rollupData.LastPendingState,
		LastPendingStateConsolidated:   rollupData.LastPendingStateConsolidated,
		LastVerifiedBatchBeforeUpgrade: rollupData.LastVerifiedBatchBeforeUpgrade,
		RollupTypeID:                   rollupData.RollupTypeID,
		RollupCompatibilityID:          rollupData.RollupCompatibilityID,
	}, nil
}

func (e *ContractRollupManangerBanana) GetBatchFee() (*big.Int, error) {
	return e.GetContract().GetBatchFee(&bind.CallOpts{Pending: false})
}

func (e *ContractRollupManangerBanana) VerifyBatchesTrustedAggregator(opts *bind.TransactOpts, rollupID uint32, pendingStateNum uint64, initNumBatch uint64, finalNewBatch uint64, newLocalExitRoot [32]byte, newStateRoot [32]byte, beneficiary common.Address, proof [24][32]byte) (*types.Transaction, error) {
	return e.GetContract().VerifyBatchesTrustedAggregator(opts, rollupID, pendingStateNum, initNumBatch, finalNewBatch, newLocalExitRoot, newStateRoot, beneficiary, proof)
}

func (e *ContractRollupManangerBanana) GetRollupSequencedBatches(rollupID uint32, batchNum uint64) (StateVariablesSequencedBatchData, error) {
	res, err := e.GetContract().GetRollupSequencedBatches(&bind.CallOpts{Pending: false}, rollupID, batchNum)
	if err != nil {
		return StateVariablesSequencedBatchData{}, err
	}
	return StateVariablesSequencedBatchData{
		AccInputHash:               res.AccInputHash,
		SequencedTimestamp:         res.SequencedTimestamp,
		PreviousLastBatchSequenced: res.PreviousLastBatchSequenced,
	}, nil
}
