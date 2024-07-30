package bridgesync

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// MTRootVectorRaw represents the root of Merkle Tree
type MTRootVectorRaw struct {
	ExistingLeaves []string         `json:"previousLeafsValues"`
	CurrentRoot    string           `json:"currentRoot"`
	NewLeaf        DepositVectorRaw `json:"newLeaf"`
	NewRoot        string           `json:"newRoot"`
}

func TestMTAddLeaf(t *testing.T) {
	data, err := os.ReadFile("testvectors/root-vectors.json")
	require.NoError(t, err)

	var mtTestVectors []MTRootVectorRaw
	err = json.Unmarshal(data, &mtTestVectors)
	require.NoError(t, err)
	ctx := context.Background()

	for ti, testVector := range mtTestVectors {
		t.Run(fmt.Sprintf("Test vector %d", ti), func(t *testing.T) {

			path := t.TempDir()
			p, err := newProcessor(context.Background(), path, "foo")
			require.NoError(t, err)

			// Add exisiting leaves
			for i, leaf := range testVector.ExistingLeaves {
				tx, err := p.db.BeginRw(ctx)
				require.NoError(t, err)
				err = p.tree.addLeaf(tx, uint32(i), common.HexToHash(leaf))
				require.NoError(t, err)
				err = tx.Commit()
				require.NoError(t, err)
			}
			if len(testVector.ExistingLeaves) > 0 {
				txRo, err := p.db.BeginRo(ctx)
				require.NoError(t, err)
				_, actualRoot, err := p.tree.getLastDepositCountAndRoot(txRo)
				txRo.Rollback()
				require.NoError(t, err)
				require.Equal(t, common.HexToHash(testVector.CurrentRoot), actualRoot)
			}

			// Add new bridge
			amount, result := big.NewInt(0).SetString(testVector.NewLeaf.Amount, 0)
			require.True(t, result)
			bridge := Bridge{
				OriginNetwork:      testVector.NewLeaf.OriginalNetwork,
				OriginAddress:      common.HexToAddress(testVector.NewLeaf.TokenAddress),
				Amount:             amount,
				DestinationNetwork: testVector.NewLeaf.DestinationNetwork,
				DestinationAddress: common.HexToAddress(testVector.NewLeaf.DestinationAddress),
				DepositCount:       uint32(len(testVector.ExistingLeaves)),
				Metadata:           common.FromHex(testVector.NewLeaf.Metadata),
			}
			tx, err := p.db.BeginRw(ctx)
			require.NoError(t, err)
			require.Equal(t, common.HexToHash(testVector.NewLeaf.CurrentHash), bridge.Hash())
			err = p.tree.addLeaf(tx, bridge.DepositCount, bridge.Hash())
			require.NoError(t, err)
			err = tx.Commit()
			txRo, err := p.db.BeginRo(ctx)
			require.NoError(t, err)
			_, actualRoot, err := p.tree.getLastDepositCountAndRoot(txRo)
			txRo.Rollback()
			require.NoError(t, err)
			require.Equal(t, common.HexToHash(testVector.NewRoot), actualRoot)
		})
	}
}

// MTClaimVectorRaw represents the merkle proof
type MTClaimVectorRaw struct {
	Deposits     []DepositVectorRaw `json:"leafs"`
	Index        uint32             `json:"index"`
	MerkleProof  []string           `json:"proof"`
	ExpectedRoot string             `json:"root"`
}

func TestMTGetProof(t *testing.T) {
	data, err := os.ReadFile("testvectors/claim-vectors.json")
	require.NoError(t, err)

	var mtTestVectors []MTClaimVectorRaw
	err = json.Unmarshal(data, &mtTestVectors)
	require.NoError(t, err)
	ctx := context.Background()

	for ti, testVector := range mtTestVectors {
		t.Run(fmt.Sprintf("Test vector %d", ti), func(t *testing.T) {
			path := t.TempDir()
			p, err := newProcessor(context.Background(), path, "foo")
			require.NoError(t, err)

			for li, leaf := range testVector.Deposits {
				amount, result := big.NewInt(0).SetString(leaf.Amount, 0)
				require.True(t, result)
				bridge := &Bridge{
					OriginNetwork:      leaf.OriginalNetwork,
					OriginAddress:      common.HexToAddress(leaf.TokenAddress),
					Amount:             amount,
					DestinationNetwork: leaf.DestinationNetwork,
					DestinationAddress: common.HexToAddress(leaf.DestinationAddress),
					DepositCount:       uint32(li),
					Metadata:           common.FromHex(leaf.Metadata),
				}
				tx, err := p.db.BeginRw(ctx)
				require.NoError(t, err)
				err = p.tree.addLeaf(tx, bridge.DepositCount, bridge.Hash())
				require.NoError(t, err)
				err = tx.Commit()
				require.NoError(t, err)
			}
			txRo, err := p.db.BeginRo(ctx)
			require.NoError(t, err)
			_, actualRoot, err := p.tree.getLastDepositCountAndRoot(txRo)
			txRo.Rollback()
			expectedRoot := common.HexToHash(testVector.ExpectedRoot)
			require.Equal(t, expectedRoot, actualRoot)

			proof, err := p.tree.getProof(ctx, testVector.Index, expectedRoot)
			require.NoError(t, err)
			for i, sibling := range testVector.MerkleProof {
				require.Equal(t, common.HexToHash(sibling), proof[i])
			}
		})
	}
}
