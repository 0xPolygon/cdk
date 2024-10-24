package config

import (
	ethermanconfig "github.com/0xPolygon/cdk/etherman/config"
	"github.com/0xPolygon/cdk/state"
)

// NetworkConfig is the configuration struct for the different environments
type NetworkConfig struct {
	// L1: Configuration related to L1
	L1Config ethermanconfig.L1Config `mapstructure:"L1"`
	// L1: Genesis of the rollup, first block number and root
	Genesis state.Genesis
}

type network string
type leafType uint8

const (
	custom network = "custom"
	// LeafTypeBalance specifies that leaf stores Balance
	LeafTypeBalance leafType = 0
	// LeafTypeNonce specifies that leaf stores Nonce
	LeafTypeNonce leafType = 1
	// LeafTypeCode specifies that leaf stores Code
	LeafTypeCode leafType = 2
	// LeafTypeStorage specifies that leaf stores Storage Value
	LeafTypeStorage leafType = 3
	// LeafTypeSCLength specifies that leaf stores Storage Value
	LeafTypeSCLength leafType = 4
)

// GenesisFromJSON is the config file for network_custom
type GenesisFromJSON struct {
	// L1: root hash of the genesis block
	Root string `json:"root"`
	// L1: block number of the genesis block
	GenesisBlockNum uint64 `json:"genesisBlockNumber"`
	// L2:  List of states contracts used to populate merkle tree at initial state
	Genesis []genesisAccountFromJSON `json:"genesis"`
	// L1: configuration of the network
	L1Config ethermanconfig.L1Config
}

type genesisAccountFromJSON struct {
	// Address of the account
	Balance string `json:"balance"`
	// Nonce of the account
	Nonce string `json:"nonce"`
	// Address of the contract
	Address string `json:"address"`
	// Byte code of the contract
	Bytecode string `json:"bytecode"`
	// Initial storage of the contract
	Storage map[string]string `json:"storage"`
	// Name of the contract in L1 (e.g. "PolygonZkEVMDeployer", "PolygonZkEVMBridge",...)
	ContractName string `json:"contractName"`
}
