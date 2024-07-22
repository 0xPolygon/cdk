package state

import (
	"github.com/ethereum/go-ethereum/common"
)

// Genesis contains the information to populate state on creation
type Genesis struct {
	// BlockNumber is the block number where the polygonZKEVM smc was deployed on L1
	BlockNumber uint64
	// Root hash of the genesis block
	Root common.Hash
	// Actions is the data to populate into the state trie
	Actions []*GenesisAction
}

// GenesisAction represents one of the values set on the SMT during genesis.
type GenesisAction struct {
	Address         string `json:"address"`
	Type            int    `json:"type"`
	StoragePosition string `json:"storagePosition"`
	Bytecode        string `json:"bytecode"`
	Key             string `json:"key"`
	Value           string `json:"value"`
	Root            string `json:"root"`
}
