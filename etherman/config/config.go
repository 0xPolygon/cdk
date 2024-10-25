package config

import (
	"github.com/0xPolygon/zkevm-ethtx-manager/etherman"
	"github.com/ethereum/go-ethereum/common"
)

// Config represents the configuration of the etherman
type Config struct {
	// URL is the URL of the Ethereum node for L1
	URL string `mapstructure:"URL"`

	EthermanConfig etherman.Config

	// ForkIDChunkSize is the max interval for each call to L1 provider to get the forkIDs
	ForkIDChunkSize uint64 `mapstructure:"ForkIDChunkSize"`
}

// L1Config represents the configuration of the network used in L1
type L1Config struct {
	// Chain ID of the L1 network
	L1ChainID uint64 `json:"chainId"`
	// ZkEVMAddr Address of the L1 contract polygonZkEVMAddress
	ZkEVMAddr common.Address `json:"polygonZkEVMAddress"`
	// RollupManagerAddr Address of the L1 contract
	RollupManagerAddr common.Address `json:"polygonRollupManagerAddress"`
	// PolAddr Address of the L1 Pol token Contract
	PolAddr common.Address `json:"polTokenAddress"`
	// GlobalExitRootManagerAddr Address of the L1 GlobalExitRootManager contract
	GlobalExitRootManagerAddr common.Address `json:"polygonZkEVMGlobalExitRootAddress"`
}
