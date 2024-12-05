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
	meta := NewCertificateMetadataFromHash(hash)
	require.Equal(t, toBlock, meta.ToBlock)
}

func TestMetadataConversions(t *testing.T) {
	fromBlock := uint64(123567890)
	offset := uint32(1000)
	createdAt := uint32(0)
	meta := NewCertificateMetadata(fromBlock, offset, createdAt)
	c := meta.ToHash()
	extractBlock := NewCertificateMetadataFromHash(c)
	require.Equal(t, fromBlock, extractBlock.FromBlock)
	require.Equal(t, offset, extractBlock.Offset)
	require.Equal(t, createdAt, extractBlock.CreatedAt)
}
