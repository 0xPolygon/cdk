package claimsponsor

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonzkevmbridgev2"
	configTypes "github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/ethtxmanager"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	// LeafTypeAsset represents a bridge asset
	LeafTypeAsset uint8 = 0
	// LeafTypeMessage represents a bridge message
	LeafTypeMessage       uint8 = 1
	gasTooHighErrTemplate       = "Claim tx estimated to consume more gas than the maximum allowed by the service. " +
		"Estimated %d, maximum allowed: %d"
)

type EthClienter interface {
	ethereum.GasEstimator
	bind.ContractBackend
}

type EthTxManager interface {
	Remove(ctx context.Context, id common.Hash) error
	ResultsByStatus(ctx context.Context, statuses []ethtxmanager.MonitoredTxStatus,
	) ([]ethtxmanager.MonitoredTxResult, error)
	Result(ctx context.Context, id common.Hash) (ethtxmanager.MonitoredTxResult, error)
	Add(ctx context.Context, to *common.Address, forcedNonce *uint64, value *big.Int, data []byte,
		gasOffset uint64, sidecar *types.BlobTxSidecar) (common.Hash, error)
}

type EVMClaimSponsor struct {
	*ClaimSponsor
	l2Client       EthClienter
	bridgeABI      *abi.ABI
	bridgeAddr     common.Address
	bridgeContract *polygonzkevmbridgev2.Polygonzkevmbridgev2
	ethTxManager   EthTxManager
	sender         common.Address
	gasOffest      uint64
	maxGas         uint64
}

type EVMClaimSponsorConfig struct {
	// DBPath path of the DB
	DBPath string `mapstructure:"DBPath"`
	// Enabled indicates if the sponsor should be run or not
	Enabled bool `mapstructure:"Enabled"`
	// SenderAddr is the address that will be used to send the claim txs
	SenderAddr common.Address `mapstructure:"SenderAddr"`
	// BridgeAddrL2 is the address of the bridge smart contract on L2
	BridgeAddrL2 common.Address `mapstructure:"BridgeAddrL2"`
	// MaxGas is the max gas (limit) allowed for a claim to be sponsored
	MaxGas uint64 `mapstructure:"MaxGas"`
	// RetryAfterErrorPeriod is the time that will be waited when an unexpected error happens before retry
	RetryAfterErrorPeriod configTypes.Duration `mapstructure:"RetryAfterErrorPeriod"`
	// MaxRetryAttemptsAfterError is the maximum number of consecutive attempts that will happen before panicing.
	// Any number smaller than zero will be considered as unlimited retries
	MaxRetryAttemptsAfterError int `mapstructure:"MaxRetryAttemptsAfterError"`
	// WaitTxToBeMinedPeriod is the period that will be used to ask if a given tx has been mined (or failed)
	WaitTxToBeMinedPeriod configTypes.Duration `mapstructure:"WaitTxToBeMinedPeriod"`
	// WaitOnEmptyQueue is the time that will be waited before trying to send the next claim of the queue
	// if the queue is empty
	WaitOnEmptyQueue configTypes.Duration `mapstructure:"WaitOnEmptyQueue"`
	// EthTxManager is the configuration of the EthTxManager to be used by the claim sponsor
	EthTxManager ethtxmanager.Config `mapstructure:"EthTxManager"`
	// GasOffset is the gas to add on top of the estimated gas when sending the claim txs
	GasOffset uint64 `mapstructure:"GasOffset"`
}

func NewEVMClaimSponsor(
	logger *log.Logger,
	dbPath string,
	l2Client EthClienter,
	bridge common.Address,
	sender common.Address,
	maxGas, gasOffset uint64,
	ethTxManager EthTxManager,
	retryAfterErrorPeriod time.Duration,
	maxRetryAttemptsAfterError int,
	waitTxToBeMinedPeriod time.Duration,
	waitOnEmptyQueue time.Duration,
) (*ClaimSponsor, error) {
	contract, err := polygonzkevmbridgev2.NewPolygonzkevmbridgev2(bridge, l2Client)
	if err != nil {
		return nil, err
	}
	abi, err := polygonzkevmbridgev2.Polygonzkevmbridgev2MetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	evmSponsor := &EVMClaimSponsor{
		l2Client:       l2Client,
		bridgeABI:      abi,
		bridgeAddr:     bridge,
		bridgeContract: contract,
		sender:         sender,
		gasOffest:      gasOffset,
		maxGas:         maxGas,
		ethTxManager:   ethTxManager,
	}
	baseSponsor, err := newClaimSponsor(
		logger,
		dbPath,
		evmSponsor,
		retryAfterErrorPeriod,
		maxRetryAttemptsAfterError,
		waitTxToBeMinedPeriod,
		waitOnEmptyQueue,
	)
	if err != nil {
		return nil, err
	}
	evmSponsor.ClaimSponsor = baseSponsor

	return baseSponsor, nil
}

func (c *EVMClaimSponsor) checkClaim(ctx context.Context, claim *Claim) error {
	data, err := c.buildClaimTxData(claim)
	if err != nil {
		return err
	}
	gas, err := c.l2Client.EstimateGas(ctx, ethereum.CallMsg{
		From: c.sender,
		To:   &c.bridgeAddr,
		Data: data,
	})
	if err != nil {
		return err
	}
	if gas > c.maxGas {
		return fmt.Errorf(gasTooHighErrTemplate, gas, c.maxGas)
	}

	return nil
}

func (c *EVMClaimSponsor) sendClaim(ctx context.Context, claim *Claim) (string, error) {
	data, err := c.buildClaimTxData(claim)
	if err != nil {
		return "", err
	}
	id, err := c.ethTxManager.Add(ctx, &c.bridgeAddr, nil, big.NewInt(0), data, c.gasOffest, nil)
	if err != nil {
		return "", err
	}

	return id.Hex(), nil
}

func (c *EVMClaimSponsor) claimStatus(ctx context.Context, id string) (ClaimStatus, error) {
	res, err := c.ethTxManager.Result(ctx, common.HexToHash(id))
	if err != nil {
		return "", err
	}
	switch res.Status {
	case ethtxmanager.MonitoredTxStatusCreated,
		ethtxmanager.MonitoredTxStatusSent:
		return WIPStatus, nil
	case ethtxmanager.MonitoredTxStatusFailed:
		return FailedClaimStatus, nil
	case ethtxmanager.MonitoredTxStatusMined,
		ethtxmanager.MonitoredTxStatusSafe,
		ethtxmanager.MonitoredTxStatusFinalized:
		return SuccessClaimStatus, nil
	default:
		return "", fmt.Errorf("unexpected tx status: %v", res.Status)
	}
}

func (c *EVMClaimSponsor) buildClaimTxData(claim *Claim) ([]byte, error) {
	switch claim.LeafType {
	case LeafTypeAsset:
		return c.bridgeABI.Pack(
			"claimAsset",
			claim.ProofLocalExitRoot,  // bytes32[32] smtProofLocalExitRoot
			claim.ProofRollupExitRoot, // bytes32[32] smtProofRollupExitRoot
			claim.GlobalIndex,         // uint256 globalIndex
			claim.MainnetExitRoot,     // bytes32 mainnetExitRoot
			claim.RollupExitRoot,      // bytes32 rollupExitRoot
			claim.OriginNetwork,       // uint32 originNetwork
			claim.OriginTokenAddress,  // address originTokenAddress,
			claim.DestinationNetwork,  // uint32 destinationNetwork
			claim.DestinationAddress,  // address destinationAddress
			claim.Amount,              // uint256 amount
			claim.Metadata,            // bytes metadata
		)
	case LeafTypeMessage:
		return c.bridgeABI.Pack(
			"claimMessage",
			claim.ProofLocalExitRoot,  // bytes32[32] smtProofLocalExitRoot
			claim.ProofRollupExitRoot, // bytes32[32] smtProofRollupExitRoot
			claim.GlobalIndex,         // uint256 globalIndex
			claim.MainnetExitRoot,     // bytes32 mainnetExitRoot
			claim.RollupExitRoot,      // bytes32 rollupExitRoot
			claim.OriginNetwork,       // uint32 originNetwork
			claim.OriginTokenAddress,  // address originTokenAddress,
			claim.DestinationNetwork,  // uint32 destinationNetwork
			claim.DestinationAddress,  // address destinationAddress
			claim.Amount,              // uint256 amount
			claim.Metadata,            // bytes metadata
		)
	default:
		return nil, fmt.Errorf("unexpected leaf type %d", claim.LeafType)
	}
}
