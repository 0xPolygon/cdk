package agglayer

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk/bridgesync"
	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type CertificateStatus int

const (
	Pending CertificateStatus = iota
	InError
	Settled
)

// String representation of the enum
func (c CertificateStatus) String() string {
	return [...]string{"Pending", "InError", "Settled"}[c]
}

type LeafType uint8

func (l LeafType) Uint8() uint8 {
	return uint8(l)
}

const (
	LeafTypeAsset LeafType = iota
	LeafTypeMessage
)

// Certificate is the data structure that will be sent to the agglayer
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
	bridgeExitsHashes := make([][]byte, len(c.BridgeExits))
	for i, bridgeExit := range c.BridgeExits {
		bridgeExitsHashes[i] = bridgeExit.Hash().Bytes()
	}

	importedBridgeExitsHashes := make([][]byte, len(c.ImportedBridgeExits))
	for i, importedBridgeExit := range c.ImportedBridgeExits {
		importedBridgeExitsHashes[i] = importedBridgeExit.Hash().Bytes()
	}

	bridgeExitsPart := crypto.Keccak256(bridgeExitsHashes...)
	importedBridgeExitsPart := crypto.Keccak256(importedBridgeExitsHashes...)

	return crypto.Keccak256Hash(
		cdkcommon.Uint32ToBytes(c.NetworkID),
		cdkcommon.Uint64ToBytes(c.Height),
		c.PrevLocalExitRoot.Bytes(),
		c.NewLocalExitRoot.Bytes(),
		bridgeExitsPart,
		importedBridgeExitsPart,
	)
}

// SignedCertificate is the struct that contains the certificate and the signature of the signer
type SignedCertificate struct {
	*Certificate
	Signature *Signature `json:"signature"`
}

// Signature is the data structure that will hold the signature of the given certificate
type Signature struct {
	R         common.Hash `json:"r"`
	S         common.Hash `json:"s"`
	OddParity bool        `json:"odd_y_parity"`
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

func (g *GlobalIndex) Hash() common.Hash {
	return crypto.Keccak256Hash(
		bridgesync.GenerateGlobalIndex(g.MainnetFlag, g.RollupIndex, g.LeafIndex).Bytes())
}

// BridgeExit represents a token bridge exit
type BridgeExit struct {
	LeafType           LeafType       `json:"leaf_type"`
	TokenInfo          *TokenInfo     `json:"token_info"`
	DestinationNetwork uint32         `json:"dest_network"`
	DestinationAddress common.Address `json:"dest_address"`
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

// MerkleProof represents an inclusion proof of a leaf in a Merkle tree
type MerkleProof struct {
	Root  common.Hash                      `json:"root"`
	Proof [types.DefaultHeight]common.Hash `json:"proof"`
}

// MarshalJSON is the implementation of the json.Marshaler interface
func (m *MerkleProof) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Root  common.Hash                                 `json:"root"`
		Proof map[string][types.DefaultHeight]common.Hash `json:"proof"`
	}{
		Root: m.Root,
		Proof: map[string][types.DefaultHeight]common.Hash{
			"siblings": m.Proof,
		},
	})
}

// Hash returns the hash of the Merkle proof struct
func (m *MerkleProof) Hash() common.Hash {
	proofsAsSingleSlice := make([]byte, 0)

	for _, proof := range m.Proof {
		proofsAsSingleSlice = append(proofsAsSingleSlice, proof.Bytes()...)
	}

	return crypto.Keccak256Hash(
		m.Root.Bytes(),
		proofsAsSingleSlice,
	)
}

// L1InfoTreeLeafInner represents the inner part of the L1 info tree leaf
type L1InfoTreeLeafInner struct {
	GlobalExitRoot common.Hash `json:"global_exit_root"`
	BlockHash      common.Hash `json:"block_hash"`
	Timestamp      uint64      `json:"timestamp"`
}

// Hash returns the hash of the L1InfoTreeLeafInner struct
func (l *L1InfoTreeLeafInner) Hash() common.Hash {
	return crypto.Keccak256Hash(
		l.GlobalExitRoot.Bytes(),
		l.BlockHash.Bytes(),
		cdkcommon.Uint64ToBytes(l.Timestamp),
	)
}

// L1InfoTreeLeaf represents the leaf of the L1 info tree
type L1InfoTreeLeaf struct {
	L1InfoTreeIndex uint32               `json:"l1_info_tree_index"`
	RollupExitRoot  common.Hash          `json:"rer"`
	MainnetExitRoot common.Hash          `json:"mer"`
	Inner           *L1InfoTreeLeafInner `json:"inner"`
}

// Hash returns the hash of the L1InfoTreeLeaf struct
func (l *L1InfoTreeLeaf) Hash() common.Hash {
	return l.Inner.Hash()
}

// Claim is the interface that will be implemented by the different types of claims
type Claim interface {
	Type() string
	Hash() common.Hash
	MarshalJSON() ([]byte, error)
}

// ClaimFromMainnnet represents a claim originating from the mainnet
type ClaimFromMainnnet struct {
	ProofLeafMER     *MerkleProof    `json:"proof_leaf_mer"`
	ProofGERToL1Root *MerkleProof    `json:"proof_ger_l1root"`
	L1Leaf           *L1InfoTreeLeaf `json:"l1_leaf"`
}

// Type is the implementation of Claim interface
func (c ClaimFromMainnnet) Type() string {
	return "Mainnet"
}

// MarshalJSON is the implementation of Claim interface
func (c *ClaimFromMainnnet) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Child map[string]interface{} `json:"Mainnet"`
	}{
		Child: map[string]interface{}{
			"proof_leaf_mer":   c.ProofLeafMER,
			"proof_ger_l1root": c.ProofGERToL1Root,
			"l1_leaf":          c.L1Leaf,
		},
	})
}

// Hash is the implementation of Claim interface
func (c *ClaimFromMainnnet) Hash() common.Hash {
	return crypto.Keccak256Hash(
		c.ProofLeafMER.Hash().Bytes(),
		c.ProofGERToL1Root.Hash().Bytes(),
		c.L1Leaf.Hash().Bytes(),
	)
}

// ClaimFromRollup represents a claim originating from a rollup
type ClaimFromRollup struct {
	ProofLeafLER     *MerkleProof    `json:"proof_leaf_ler"`
	ProofLERToRER    *MerkleProof    `json:"proof_ler_rer"`
	ProofGERToL1Root *MerkleProof    `json:"proof_ger_l1root"`
	L1Leaf           *L1InfoTreeLeaf `json:"l1_leaf"`
}

// Type is the implementation of Claim interface
func (c ClaimFromRollup) Type() string {
	return "Rollup"
}

// MarshalJSON is the implementation of Claim interface
func (c *ClaimFromRollup) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Child map[string]interface{} `json:"Rollup"`
	}{
		Child: map[string]interface{}{
			"proof_leaf_ler":   c.ProofLeafLER,
			"proof_ler_rer":    c.ProofLERToRER,
			"proof_ger_l1root": c.ProofGERToL1Root,
			"l1_leaf":          c.L1Leaf,
		},
	})
}

// Hash is the implementation of Claim interface
func (c *ClaimFromRollup) Hash() common.Hash {
	return crypto.Keccak256Hash(
		c.ProofLeafLER.Hash().Bytes(),
		c.ProofLERToRER.Hash().Bytes(),
		c.ProofGERToL1Root.Hash().Bytes(),
		c.L1Leaf.Hash().Bytes(),
	)
}

// ImportedBridgeExit represents a token bridge exit originating on another network but claimed on the current network.
type ImportedBridgeExit struct {
	BridgeExit  *BridgeExit  `json:"bridge_exit"`
	ClaimData   Claim        `json:"claim_data"`
	GlobalIndex *GlobalIndex `json:"global_index"`
}

// Hash returns a hash that uniquely identifies the imported bridge exit
func (c *ImportedBridgeExit) Hash() common.Hash {
	return crypto.Keccak256Hash(
		c.BridgeExit.Hash().Bytes(),
		c.ClaimData.Hash().Bytes(),
		c.GlobalIndex.Hash().Bytes(),
	)
}

// CertificateHeader is the structure returned by the interop_getCertificateHeader RPC call
type CertificateHeader struct {
	NetworkID        uint32            `json:"network_id"`
	Height           uint64            `json:"height"`
	EpochNumber      *uint64           `json:"epoch_number"`
	CertificateIndex *uint64           `json:"certificate_index"`
	CertificateID    common.Hash       `json:"certificate_id"`
	NewLocalExitRoot common.Hash       `json:"new_local_exit_root"`
	Status           CertificateStatus `json:"status"`
}

func (c CertificateHeader) String() string {
	return fmt.Sprintf("Height: %d, CertificateID: %s, NewLocalExitRoot: %s",
		c.Height, c.CertificateID.String(), c.NewLocalExitRoot.String())
}
