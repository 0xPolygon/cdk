package testvectors

import (
	"encoding/binary"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/keccak256"
)

// DepositVectorRaw represents the deposit vector
type DepositVectorRaw struct {
	OriginNetwork      uint32 `json:"originNetwork"`
	TokenAddress       string `json:"tokenAddress"`
	Amount             string `json:"amount"`
	DestinationNetwork uint32 `json:"destinationNetwork"`
	DestinationAddress string `json:"destinationAddress"`
	ExpectedHash       string `json:"leafValue"`
	CurrentHash        string `json:"currentLeafValue"`
	Metadata           string `json:"metadata"`
}

func (d *DepositVectorRaw) Hash() common.Hash {
	origNet := make([]byte, 4) //nolint:mnd
	binary.BigEndian.PutUint32(origNet, d.OriginNetwork)
	destNet := make([]byte, 4) //nolint:mnd
	binary.BigEndian.PutUint32(destNet, d.DestinationNetwork)

	metaHash := keccak256.Hash(common.FromHex(d.Metadata))
	var buf [32]byte
	amount, _ := big.NewInt(0).SetString(d.Amount, 0)
	origAddrBytes := common.HexToAddress(d.TokenAddress)
	destAddrBytes := common.HexToAddress(d.DestinationAddress)
	return common.BytesToHash(keccak256.Hash(
		[]byte{0}, // LeafType
		origNet,
		origAddrBytes[:],
		destNet,
		destAddrBytes[:],
		amount.FillBytes(buf[:]),
		metaHash,
	))
}

// MTClaimVectorRaw represents the merkle proof
type MTClaimVectorRaw struct {
	Deposits     []DepositVectorRaw `json:"leafs"`
	Index        uint32             `json:"index"`
	MerkleProof  []string           `json:"proof"`
	ExpectedRoot string             `json:"root"`
}

// MTRootVectorRaw represents the root of Merkle Tree
type MTRootVectorRaw struct {
	ExistingLeaves []string         `json:"previousLeafsValues"`
	CurrentRoot    string           `json:"currentRoot"`
	NewLeaf        DepositVectorRaw `json:"newLeaf"`
	NewRoot        string           `json:"newRoot"`
}
