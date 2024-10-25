package agglayer

import (
	"crypto/ecdsa"

	"github.com/0xPolygon/cdk-rpc/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ZKP is the struct that contains the zero-knowledge proof
type ZKP struct {
	NewStateRoot     common.Hash    `json:"newStateRoot"`
	NewLocalExitRoot common.Hash    `json:"newLocalExitRoot"`
	Proof            types.ArgBytes `json:"proof"`
}

// Tx is the struct that contains the verified batch transaction
type Tx struct {
	RollupID          uint32
	LastVerifiedBatch types.ArgUint64 `json:"lastVerifiedBatch"`
	NewVerifiedBatch  types.ArgUint64 `json:"newVerifiedBatch"`
	ZKP               ZKP             `json:"ZKP"`
}

// Hash returns a hash that uniquely identifies the tx
func (t *Tx) Hash() common.Hash {
	return common.BytesToHash(crypto.Keccak256(
		[]byte(t.LastVerifiedBatch.Hex()),
		[]byte(t.NewVerifiedBatch.Hex()),
		t.ZKP.NewStateRoot[:],
		t.ZKP.NewLocalExitRoot[:],
		[]byte(t.ZKP.Proof.Hex()),
	))
}

// Sign returns a signed batch by the private key
func (t *Tx) Sign(privateKey *ecdsa.PrivateKey) (*SignedTx, error) {
	hashToSign := t.Hash()
	sig, err := crypto.Sign(hashToSign.Bytes(), privateKey)
	if err != nil {
		return nil, err
	}

	return &SignedTx{
		Tx:        *t,
		Signature: sig,
	}, nil
}

// SignedTx is the struct that contains the signed batch transaction
type SignedTx struct {
	Tx        Tx             `json:"tx"`
	Signature types.ArgBytes `json:"signature"`
}

// Signer returns the address of the signer
func (s *SignedTx) Signer() (common.Address, error) {
	pubKey, err := crypto.SigToPub(s.Tx.Hash().Bytes(), s.Signature)
	if err != nil {
		return common.Address{}, err
	}

	return crypto.PubkeyToAddress(*pubKey), nil
}
