package agglayer

import (
	"math/big"

	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type LeafType uint8

func (l LeafType) Uint8() uint8 {
	return uint8(l)
}

const (
	LeafTypeAsset   LeafType = 0
	LeafTypeMessage LeafType = 1
)

// Certificate is the data structure that will be sent to the aggLayer
type Certificate struct {
	NetworkID           uint32                `json:"network_id"`
	Height              uint64                `json:"height"`
	PrevLocalExitRoot   common.Hash           `json:"prev_local_exit_root"`
	NewLocalExitRoot    common.Hash           `json:"new_local_exit_root"`
	BridgeExits         []*BridgeExit         `json:"bridge_exits"`
	ImportedBridgeExits []*ImportedBridgeExit `json:"imported_bridge_exits"`
}

// Hash returns a hash that uniquely identifies the certificate
func (c *Certificate) Hash() common.Hash {
	importedBridgeHashes := make([][]byte, 0, len(c.ImportedBridgeExits))
	for _, claim := range c.ImportedBridgeExits {
		importedBridgeHashes = append(importedBridgeHashes, claim.Hash().Bytes())
	}

	incomeCommitment := crypto.Keccak256Hash(importedBridgeHashes...)
	outcomeCommitment := c.NewLocalExitRoot

	return crypto.Keccak256Hash(outcomeCommitment.Bytes(), incomeCommitment.Bytes())
}

// SignedCertificate is the struct that contains the certificate and the signature of the signer
type SignedCertificate struct {
	*Certificate
	Signature []byte `json:"signature"`
}

// TokenInfo encapsulates the information to uniquely identify a token on the origin network.
type TokenInfo struct {
	OriginNetwork      uint32         `json:"origin_network"`
	OriginTokenAddress common.Address `json:"origin_token_address"`
}

// GlobalIndex represents the global index of an imported bridge exit
type GlobalIndex struct {
	MainnetFlag bool   `json:"mainnet_flag"`
	RollupIndex uint32 `json:"rollup_index"`
	LeafIndex   uint32 `json:"leaf_index"`
}

// BridgeExit represents a token bridge exit
type BridgeExit struct {
	LeafType           LeafType       `json:"leaf_type"`
	TokenInfo          *TokenInfo     `json:"token_info"`
	DestinationNetwork uint32         `json:"destination_network"`
	DestinationAddress common.Address `json:"destination_address"`
	Amount             *big.Int       `json:"amount"`
	Metadata           []byte         `json:"metadata"`
}

// Hash returns a hash that uniquely identifies the bridge exit
func (c *BridgeExit) Hash() common.Hash {
	if c.Amount == nil {
		c.Amount = big.NewInt(0)
	}

	return crypto.Keccak256Hash(
		[]byte{c.LeafType.Uint8()},
		cdkcommon.Uint32ToBytes(c.TokenInfo.OriginNetwork),
		c.TokenInfo.OriginTokenAddress.Bytes(),
		cdkcommon.Uint32ToBytes(c.DestinationNetwork),
		c.DestinationAddress.Bytes(),
		c.Amount.Bytes(),
		crypto.Keccak256(c.Metadata),
	)
}

// ImportedBridgeExit represents a token bridge exit originating on another network but claimed on the current network.
type ImportedBridgeExit struct {
	BridgeExit            *BridgeExit                      `json:"bridge_exit"`
	ImportedLocalExitRoot common.Hash                      `json:"imported_local_exit_root"`
	InclusionProof        [types.DefaultHeight]common.Hash `json:"inclusion_proof"`
	InclusionProofRER     [types.DefaultHeight]common.Hash `json:"inclusion_proof_rer"`
	GlobalIndex           *GlobalIndex                     `json:"global_index"`
}

// Hash returns a hash that uniquely identifies the imported bridge exit
func (c *ImportedBridgeExit) Hash() common.Hash {
	return c.BridgeExit.Hash()
}
