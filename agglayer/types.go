package agglayer

import (
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
	return crypto.Keccak256Hash(
		cdkcommon.Uint32ToBytes(c.NetworkID),
		cdkcommon.Uint64ToBytes(c.Height),
		c.PrevLocalExitRoot.Bytes(),
		c.NewLocalExitRoot.Bytes(),
	)
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

type MerkleProof struct {
	Root  common.Hash                      `json:"root"`
	Proof [types.DefaultHeight]common.Hash `json:"proof"`
}

type L1InfoTreeLeafInner struct {
	GlobalExitRoot common.Hash `json:"global_exit_root"`
	BlockHash      common.Hash `json:"block_hash"`
	Timestamp      uint64      `json:"timestamp"`
}

func (l L1InfoTreeLeafInner) Hash() common.Hash {
	return crypto.Keccak256Hash(l.GlobalExitRoot.Bytes(), l.BlockHash.Bytes(), cdkcommon.Uint64ToBytes(l.Timestamp))
}

type L1InfoTreeLeaf struct {
	L1InfoTreeIndex uint32              `json:"l1_info_tree_index"`
	RollupExitRoot  common.Hash         `json:"rer"`
	MainnetExitRoot common.Hash         `json:"mer"`
	Inner           L1InfoTreeLeafInner `json:"inner"`
}

func (l L1InfoTreeLeaf) Hash() common.Hash {
	return l.Inner.Hash()
}

type Claim interface {
	Type() string
}

type ClaimFromMainnnet struct {
	ProofLeafMER     MerkleProof    `json:"proof_leaf_mer"`
	ProofGERToL1Root MerkleProof    `json:"proof_ger_l1root"`
	L1Leaf           L1InfoTreeLeaf `json:"l1_leaf"`
}

func (c ClaimFromMainnnet) Type() string {
	return "Mainnet"
}

type ClaimFromRollup struct {
	ProofLeafLER     MerkleProof    `json:"proof_leaf_ler"`
	ProofLERToRER    MerkleProof    `json:"proof_ler_rer"`
	ProofGERToL1Root MerkleProof    `json:"proof_ger_l1root"`
	L1Leaf           L1InfoTreeLeaf `json:"l1_leaf"`
}

func (c ClaimFromRollup) Type() string {
	return "Rollup"
}

// ImportedBridgeExit represents a token bridge exit originating on another network but claimed on the current network.
type ImportedBridgeExit struct {
	BridgeExit  *BridgeExit  `json:"bridge_exit"`
	ClaimData   Claim        `json:"claim"`
	GlobalIndex *GlobalIndex `json:"global_index"`
}

// Hash returns a hash that uniquely identifies the imported bridge exit
func (c *ImportedBridgeExit) Hash() common.Hash {
	globalIndexBig := bridgesync.GenerateGlobalIndex(c.GlobalIndex.MainnetFlag,
		c.GlobalIndex.RollupIndex, c.GlobalIndex.LeafIndex)
	return crypto.Keccak256Hash(globalIndexBig.Bytes())
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
