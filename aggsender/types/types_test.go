package types

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestMetadataConversions_toBlock_Only(t *testing.T) {
	toBlock := uint64(123567890)
	hash := common.BigToHash(new(big.Int).SetUint64(toBlock))
	extractBlock := NewCertificateMetadataFromHash(hash)
	require.Equal(t, toBlock, extractBlock.ToBlock)
}

func TestMetadataConversions(t *testing.T) {
	fromBlock := uint64(123567890)
	toBlock := uint64(123567890)
	createdAt := uint64(0)
	meta := NewCertificateMetadata(fromBlock, toBlock, createdAt)
	c := meta.ToHash()
	extractBlock := NewCertificateMetadataFromHash(c)
	require.Equal(t, fromBlock, extractBlock.FromBlock)
	require.Equal(t, toBlock, extractBlock.ToBlock)
	require.Equal(t, createdAt, extractBlock.CreatedAt)
}
