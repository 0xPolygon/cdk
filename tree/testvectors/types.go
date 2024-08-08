package testvectors

import (
	"encoding/binary"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/keccak256"
)

// DepositVectorRaw represents the deposit vector
type DepositVectorRaw struct {
	OriginalNetwork    uint32 `json:"originNetwork"`
	TokenAddress       string `json:"tokenAddress"`
	Amount             string `json:"amount"`
	DestinationNetwork uint32 `json:"destinationNetwork"`
	DestinationAddress string `json:"destinationAddress"`
	ExpectedHash       string `json:"leafValue"`
	CurrentHash        string `json:"currentLeafValue"`
	Metadata           string `json:"metadata"`
}

func (d *DepositVectorRaw) Hash() common.Hash {
	origNet := make([]byte, 4) //nolint:gomnd
	binary.BigEndian.PutUint32(origNet, uint32(d.OriginalNetwork))
	destNet := make([]byte, 4) //nolint:gomnd
	binary.BigEndian.PutUint32(destNet, uint32(d.DestinationNetwork))

	metaHash := keccak256.Hash(common.FromHex(d.Metadata))
	hash := common.Hash{}
	var buf [32]byte //nolint:gomnd
	amount, _ := big.NewInt(0).SetString(d.Amount, 0)
	origAddrBytes := common.HexToAddress(d.TokenAddress)
	destAddrBytes := common.HexToAddress(d.DestinationAddress)
	copy(
		hash[:],
		keccak256.Hash(
			[]byte{0},
			origNet,
			origAddrBytes[:],
			destNet,
			destAddrBytes[:],
			amount.FillBytes(buf[:]),
			metaHash,
		),
	)
	return hash
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