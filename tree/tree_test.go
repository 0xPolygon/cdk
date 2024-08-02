package tree

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/0xPolygon/cdk/tree/testvectors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
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
			dbPrefix := "foo"
			tableCfgFunc := func(defaultBuckets kv.TableCfg) kv.TableCfg {
				cfg := kv.TableCfg{}
				AddTables(cfg, dbPrefix)
				return cfg
			}
			db, err := mdbx.NewMDBX(nil).
				Path(path).
				WithTableCfg(tableCfgFunc).
				Open()
			require.NoError(t, err)
			tree, err := NewAppendOnly(context.Background(), db, dbPrefix)
			require.NoError(t, err)

			// Add exisiting leaves
			leaves := []Leaf{}
			for i, leaf := range testVector.ExistingLeaves {
				leaves = append(leaves, Leaf{
					Index: uint32(i),
					Hash:  common.HexToHash(leaf),
				})
			}
			tx, err := db.BeginRw(ctx)
			require.NoError(t, err)
			_, err = tree.AddLeaves(tx, leaves)
			require.NoError(t, err)
			require.NoError(t, tx.Commit())
			if len(testVector.ExistingLeaves) > 0 {
				txRo, err := tree.db.BeginRo(ctx)
				require.NoError(t, err)
				_, actualRoot, err := tree.getLastIndexAndRootWithTx(txRo)
				txRo.Rollback()
				require.NoError(t, err)
				require.Equal(t, common.HexToHash(testVector.CurrentRoot), actualRoot)
			}

			// Add new bridge
			tx, err = db.BeginRw(ctx)
			require.NoError(t, err)
			_, err = tree.AddLeaves(tx, []Leaf{{
				Index: uint32(len(testVector.ExistingLeaves)),
				Hash:  common.HexToHash(testVector.NewLeaf.CurrentHash),
			}})
			require.NoError(t, err)
			require.NoError(t, tx.Commit())

			txRo, err := tree.db.BeginRo(ctx)
			require.NoError(t, err)
			_, actualRoot, err := tree.getLastIndexAndRootWithTx(txRo)
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
			dbPrefix := "foo"
			tableCfgFunc := func(defaultBuckets kv.TableCfg) kv.TableCfg {
				cfg := kv.TableCfg{}
				AddTables(cfg, dbPrefix)
				return cfg
			}
			db, err := mdbx.NewMDBX(nil).
				Path(path).
				WithTableCfg(tableCfgFunc).
				Open()
			require.NoError(t, err)
			tree, err := NewAppendOnly(context.Background(), db, dbPrefix)
			require.NoError(t, err)

			leaves := []Leaf{}
			for li, leaf := range testVector.Deposits {
				leaves = append(leaves, Leaf{
					Index: uint32(li),
					Hash:  leaf.Hash(),
				})
			}
			tx, err := db.BeginRw(ctx)
			require.NoError(t, err)
			_, err = tree.AddLeaves(tx, leaves)
			require.NoError(t, err)
			require.NoError(t, tx.Commit())

			txRo, err := tree.db.BeginRo(ctx)
			require.NoError(t, err)
			_, actualRoot, err := tree.getLastIndexAndRootWithTx(txRo)
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
