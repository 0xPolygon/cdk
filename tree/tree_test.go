package tree_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/tree"
	"github.com/0xPolygon/cdk/tree/migrations"
	"github.com/0xPolygon/cdk/tree/testvectors"
	"github.com/0xPolygon/cdk/tree/types"
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
			dbPath := path.Join(t.TempDir(), "file::memory:?cache=shared")
			log.Debug("DB created at: ", dbPath)
			err := migrations.RunMigrations(dbPath)
			require.NoError(t, err)
			treeDB, err := db.NewSQLiteDB(dbPath)
			require.NoError(t, err)
			_, err = treeDB.Exec(`select * from root`)
			require.NoError(t, err)
			merkletree := tree.NewAppendOnlyTree(treeDB, "")

			// Add exisiting leaves
			tx, err := db.NewTx(ctx, treeDB)
			require.NoError(t, err)
			for i, leaf := range testVector.ExistingLeaves {
				err = merkletree.AddLeaf(tx, uint64(i), 0, types.Leaf{
					Index: uint32(i),
					Hash:  common.HexToHash(leaf),
				})
				require.NoError(t, err)
			}
			require.NoError(t, tx.Commit())
			if len(testVector.ExistingLeaves) > 0 {
				root, err := merkletree.GetLastRoot(nil)
				require.NoError(t, err)
				require.Equal(t, common.HexToHash(testVector.CurrentRoot), root.Hash)
			}

			// Add new bridge
			tx, err = db.NewTx(ctx, treeDB)
			require.NoError(t, err)
			err = merkletree.AddLeaf(tx, uint64(len(testVector.ExistingLeaves)), 0, types.Leaf{
				Index: uint32(len(testVector.ExistingLeaves)),
				Hash:  common.HexToHash(testVector.NewLeaf.CurrentHash),
			})
			require.NoError(t, err)
			require.NoError(t, tx.Commit())

			root, err := merkletree.GetLastRoot(nil)
			require.NoError(t, err)
			require.Equal(t, common.HexToHash(testVector.NewRoot), root.Hash)
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
			dbPath := path.Join(t.TempDir(), "file::memory:?cache=shared")
			err := migrations.RunMigrations(dbPath)
			require.NoError(t, err)
			treeDB, err := db.NewSQLiteDB(dbPath)
			require.NoError(t, err)
			tre := tree.NewAppendOnlyTree(treeDB, "")

			tx, err := db.NewTx(ctx, treeDB)
			require.NoError(t, err)
			for li, leaf := range testVector.Deposits {
				err = tre.AddLeaf(tx, uint64(li), 0, types.Leaf{
					Index: uint32(li),
					Hash:  leaf.Hash(),
				})
				require.NoError(t, err)
			}
			require.NoError(t, tx.Commit())

			root, err := tre.GetLastRoot(nil)
			require.NoError(t, err)
			expectedRoot := common.HexToHash(testVector.ExpectedRoot)
			require.Equal(t, expectedRoot, root.Hash)

			proof, err := tre.GetProof(ctx, testVector.Index, expectedRoot)
			require.NoError(t, err)
			for i, sibling := range testVector.MerkleProof {
				require.Equal(t, common.HexToHash(sibling), proof[i])
			}
		})
	}
}

func createTreeDBForTest(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := "file::memory:?cache=shared"
	err := migrations.RunMigrations(dbPath)
	require.NoError(t, err)
	treeDB, err := db.NewSQLiteDB(dbPath)
	require.NoError(t, err)
	return treeDB
}
