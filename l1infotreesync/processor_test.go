package l1infotreesync

import (
	"testing"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/sync"
	"github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestGetInfo(t *testing.T) {
	dbPath := "file:TestGetInfo?mode=memory&cache=shared"
	p, err := newProcessor(dbPath)
	require.NoError(t, err)
	ctx := context.Background()

	// Test ErrNotFound returned correctly on all methods
	_, err = p.GetFirstL1InfoWithRollupExitRoot(common.Hash{})
	require.Equal(t, db.ErrNotFound, err)
	_, err = p.GetLastInfo()
	require.Equal(t, db.ErrNotFound, err)
	_, err = p.GetFirstInfo()
	require.Equal(t, db.ErrNotFound, err)
	_, err = p.GetFirstInfoAfterBlock(0)
	require.Equal(t, db.ErrNotFound, err)
	_, err = p.GetInfoByGlobalExitRoot(common.Hash{})
	require.Equal(t, db.ErrNotFound, err)

	// First insert
	info1 := &UpdateL1InfoTree{
		MainnetExitRoot: common.HexToHash("beef"),
		RollupExitRoot:  common.HexToHash("5ca1e"),
		ParentHash:      common.HexToHash("1010101"),
		Timestamp:       420,
	}
	expected1 := L1InfoTreeLeaf{
		BlockNumber:       1,
		L1InfoTreeIndex:   0,
		PreviousBlockHash: info1.ParentHash,
		Timestamp:         info1.Timestamp,
		MainnetExitRoot:   info1.MainnetExitRoot,
		RollupExitRoot:    info1.RollupExitRoot,
	}
	expected1.GlobalExitRoot = expected1.globalExitRoot()
	expected1.Hash = expected1.hash()
	err = p.ProcessBlock(ctx, sync.Block{
		Num: 1,
		Events: []interface{}{
			Event{UpdateL1InfoTree: info1},
		},
	})
	require.NoError(t, err)
	actual, err := p.GetFirstL1InfoWithRollupExitRoot(info1.RollupExitRoot)
	require.NoError(t, err)
	require.Equal(t, expected1, *actual)
	actual, err = p.GetLastInfo()
	require.NoError(t, err)
	require.Equal(t, expected1, *actual)
	actual, err = p.GetFirstInfo()
	require.NoError(t, err)
	require.Equal(t, expected1, *actual)
	actual, err = p.GetFirstInfoAfterBlock(0)
	require.NoError(t, err)
	require.Equal(t, expected1, *actual)
	actual, err = p.GetInfoByGlobalExitRoot(expected1.GlobalExitRoot)
	require.NoError(t, err)
	require.Equal(t, expected1, *actual)

	// Second insert
	info2 := &UpdateL1InfoTree{
		MainnetExitRoot: common.HexToHash("b055"),
		RollupExitRoot:  common.HexToHash("5ca1e"),
		ParentHash:      common.HexToHash("1010101"),
		Timestamp:       420,
	}
	expected2 := L1InfoTreeLeaf{
		BlockNumber:       2,
		L1InfoTreeIndex:   1,
		PreviousBlockHash: info2.ParentHash,
		Timestamp:         info2.Timestamp,
		MainnetExitRoot:   info2.MainnetExitRoot,
		RollupExitRoot:    info2.RollupExitRoot,
	}
	expected2.GlobalExitRoot = expected2.globalExitRoot()
	expected2.Hash = expected2.hash()
	err = p.ProcessBlock(ctx, sync.Block{
		Num: 2,
		Events: []interface{}{
			Event{UpdateL1InfoTree: info2},
		},
	})
	require.NoError(t, err)
	actual, err = p.GetFirstL1InfoWithRollupExitRoot(info2.RollupExitRoot)
	require.NoError(t, err)
	require.Equal(t, expected1, *actual)
	actual, err = p.GetLastInfo()
	require.NoError(t, err)
	require.Equal(t, expected2, *actual)
	actual, err = p.GetFirstInfo()
	require.NoError(t, err)
	require.Equal(t, expected1, *actual)
	actual, err = p.GetFirstInfoAfterBlock(2)
	require.NoError(t, err)
	require.Equal(t, expected2, *actual)
	actual, err = p.GetInfoByGlobalExitRoot(expected2.GlobalExitRoot)
	require.NoError(t, err)
	require.Equal(t, expected2, *actual)
}

func TestGetLatestInfoUntilBlockIfNotFoundReturnsErrNotFound(t *testing.T) {
	dbPath := "file:TestGetLatestInfoUntilBlock?mode=memory&cache=shared"
	sut, err := newProcessor(dbPath)
	require.NoError(t, err)
	ctx := context.Background()
	// Fake block 1
	_, err = sut.db.Exec(`INSERT INTO block (num) VALUES ($1)`, 1)
	require.NoError(t, err)

	_, err = sut.GetLatestInfoUntilBlock(ctx, 1)
	require.Equal(t, db.ErrNotFound, err)
}

func Test_processor_GetL1InfoTreeMerkleProof(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name         string
		getProcessor func(t *testing.T) *processor
		idx          uint32
		expectedRoot types.Root
		expectedErr  error
	}{
		{
			name: "empty tree",
			getProcessor: func(t *testing.T) *processor {
				t.Helper()

				p, err := newProcessor("file:Test_processor_GetL1InfoTreeMerkleProof_1?mode=memory&cache=shared")
				require.NoError(t, err)

				return p
			},
			idx:         0,
			expectedErr: db.ErrNotFound,
		},
		{
			name: "single leaf tree",
			getProcessor: func(t *testing.T) *processor {
				t.Helper()

				p, err := newProcessor("file:Test_processor_GetL1InfoTreeMerkleProof_2?mode=memory&cache=shared")
				require.NoError(t, err)

				info := &UpdateL1InfoTree{
					MainnetExitRoot: common.HexToHash("beef"),
					RollupExitRoot:  common.HexToHash("5ca1e"),
					ParentHash:      common.HexToHash("1010101"),
					Timestamp:       420,
				}
				err = p.ProcessBlock(context.Background(), sync.Block{
					Num: 1,
					Events: []interface{}{
						Event{UpdateL1InfoTree: info},
					},
				})
				require.NoError(t, err)

				return p
			},
			idx: 0,
			expectedRoot: types.Root{
				Hash:          common.HexToHash("beef"),
				Index:         0,
				BlockNum:      1,
				BlockPosition: 0,
			},
		},
	}

	for _, tt := range testTable {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := tt.getProcessor(t)
			proof, root, err := p.GetL1InfoTreeMerkleProof(context.Background(), tt.idx)
			if tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, proof)
				require.NotEmpty(t, root.Hash)
				require.Equal(t, tt.expectedRoot.Index, root.Index)
				require.Equal(t, tt.expectedRoot.BlockNum, root.BlockNum)
				require.Equal(t, tt.expectedRoot.BlockPosition, root.BlockPosition)
			}
		})
	}
}

func Test_processor_Reorg(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name         string
		getProcessor func(t *testing.T) *processor
		reorgBlock   uint64
		expectedErr  error
	}{
		{
			name: "empty tree",
			getProcessor: func(t *testing.T) *processor {
				t.Helper()

				p, err := newProcessor("file:Test_processor_Reorg_1?mode=memory&cache=shared")
				require.NoError(t, err)
				return p
			},
			reorgBlock:  0,
			expectedErr: nil,
		},
		{
			name: "single leaf tree",
			getProcessor: func(t *testing.T) *processor {
				t.Helper()

				p, err := newProcessor("file:Test_processor_Reorg_2?mode=memory&cache=shared")
				require.NoError(t, err)

				info := &UpdateL1InfoTree{
					MainnetExitRoot: common.HexToHash("beef"),
					RollupExitRoot:  common.HexToHash("5ca1e"),
					ParentHash:      common.HexToHash("1010101"),
					Timestamp:       420,
				}
				err = p.ProcessBlock(context.Background(), sync.Block{
					Num: 1,
					Events: []interface{}{
						Event{UpdateL1InfoTree: info},
					},
				})
				require.NoError(t, err)

				return p
			},
			reorgBlock: 1,
		},
	}

	for _, tt := range testTable {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := tt.getProcessor(t)
			err := p.Reorg(context.Background(), tt.reorgBlock)
			if tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
