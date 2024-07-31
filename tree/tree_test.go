package tree

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/0xPolygon/cdk/tree/testvectors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestMTAddLeaf(t *testing.T) {
	data, err := os.ReadFile("testvectors/root-vectors.json")
	require.NoError(t, err)

	var mtTestVectors []testvectors.MTRootVectorRaw
	err = json.Unmarshal(data, &mtTestVectors)
	require.NoError(t, err)
	ctx := context.Background()

	for ti, testVector := range mtTestVectors {
		t.Run(fmt.Sprintf("Test vector %d", ti), func(t *testing.T) {

			path := t.TempDir()
			tree, err := NewAppendOnly(context.Background(), path, "foo")
			require.NoError(t, err)

			// Add exisiting leaves
			leaves := []Leaf{}
			for i, leaf := range testVector.ExistingLeaves {
				leaves = append(leaves, Leaf{
					Index: uint32(i),
					Hash:  common.HexToHash(leaf),
				})
			}
			err = tree.AddLeaves(ctx, leaves)
			require.NoError(t, err)
			if len(testVector.ExistingLeaves) > 0 {
				txRo, err := tree.db.BeginRo(ctx)
				require.NoError(t, err)
				_, actualRoot, err := tree.getLastIndexAndRoot(txRo)
				txRo.Rollback()
				require.NoError(t, err)
				require.Equal(t, common.HexToHash(testVector.CurrentRoot), actualRoot)
			}

			// Add new bridge
			err = tree.AddLeaves(ctx, []Leaf{{
				Index: uint32(len(testVector.ExistingLeaves)),
				Hash:  common.HexToHash(testVector.NewLeaf.CurrentHash),
			}})
			require.NoError(t, err)
			txRo, err := tree.db.BeginRo(ctx)
			require.NoError(t, err)
			_, actualRoot, err := tree.getLastIndexAndRoot(txRo)
			txRo.Rollback()
			require.NoError(t, err)
			require.Equal(t, common.HexToHash(testVector.NewRoot), actualRoot)
		})
	}
}

func TestMTGetProof(t *testing.T) {
	data, err := os.ReadFile("testvectors/claim-vectors.json")
	require.NoError(t, err)

	var mtTestVectors []testvectors.MTClaimVectorRaw
	err = json.Unmarshal(data, &mtTestVectors)
	require.NoError(t, err)
	ctx := context.Background()

	for ti, testVector := range mtTestVectors {
		t.Run(fmt.Sprintf("Test vector %d", ti), func(t *testing.T) {
			path := t.TempDir()
			tree, err := NewAppendOnly(context.Background(), path, "foo")
			require.NoError(t, err)
			leaves := []Leaf{}
			for li, leaf := range testVector.Deposits {
				leaves = append(leaves, Leaf{
					Index: uint32(li),
					Hash:  leaf.Hash(),
				})
			}
			err = tree.AddLeaves(ctx, leaves)
			require.NoError(t, err)
			txRo, err := tree.db.BeginRo(ctx)
			require.NoError(t, err)
			_, actualRoot, err := tree.getLastIndexAndRoot(txRo)
			txRo.Rollback()
			expectedRoot := common.HexToHash(testVector.ExpectedRoot)
			require.Equal(t, expectedRoot, actualRoot)

			proof, err := tree.GetProof(ctx, testVector.Index, expectedRoot)
			require.NoError(t, err)
			for i, sibling := range testVector.MerkleProof {
				require.Equal(t, common.HexToHash(sibling), proof[i])
			}
		})
	}
}
