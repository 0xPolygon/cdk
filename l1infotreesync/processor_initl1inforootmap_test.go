package l1infotreesync

import (
	"context"
	"path"
	"testing"

	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestInitL1InfoRootMap(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "l1infotreesyncTestInitL1InfoRootMap.sqlite")
	sut, err := newProcessor(dbPath)
	require.NoError(t, err)
	ctx := context.TODO()
	event := InitL1InfoRootMap{
		LeafCount:         1,
		CurrentL1InfoRoot: common.HexToHash("beef"),
	}
	block := sync.Block{
		Num: 1,
		Events: []interface{}{
			Event{InitL1InfoRootMap: &event},
		},
	}

	err = sut.ProcessBlock(ctx, block)
	require.NoError(t, err)

	info, err := sut.GetInitL1InfoRootMap(nil)
	require.NoError(t, err)
	require.NotNil(t, info)
	require.Equal(t, event.LeafCount, info.LeafCount)
	require.Equal(t, event.CurrentL1InfoRoot, info.L1InfoRoot)
	require.Equal(t, block.Num, info.BlockNumber)
}

func TestInitL1InfoRootMapDontAllow2Rows(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "l1infotreesyncTestInitL1InfoRootMapDontAllow2Rows.sqlite")
	sut, err := newProcessor(dbPath)
	require.NoError(t, err)
	ctx := context.TODO()
	block := sync.Block{
		Num: 1,
		Events: []interface{}{
			Event{InitL1InfoRootMap: &InitL1InfoRootMap{
				LeafCount:         1,
				CurrentL1InfoRoot: common.HexToHash("beef"),
			}},
		},
	}
	err = sut.ProcessBlock(ctx, block)
	require.NoError(t, err)
	block.Num = 2
	err = sut.ProcessBlock(ctx, block)
	require.Error(t, err, "should not allow to insert a second row")
}

func TestGetInitL1InfoRootMap(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "l1infotreesyncTestGetInitL1InfoRootMap.sqlite")
	sut, err := newProcessor(dbPath)
	require.NoError(t, err)
	info, err := sut.GetInitL1InfoRootMap(nil)
	require.NoError(t, err, "should return no error if no row is present, because it returns data=nil")
	require.Nil(t, info, "should return nil if no row is present")
}
