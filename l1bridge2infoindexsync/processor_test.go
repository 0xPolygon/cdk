package l1bridge2infoindexsync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDuplicatedKey(t *testing.T) {
	dbPath := t.TempDir()
	p, err := newProcessor(dbPath)
	require.NoError(t, err)
	ctx := context.Background()
	err = p.processUntilBlock(ctx, 5, []bridge2L1InfoRelation{{bridgeIndex: 2, l1InfoTreeIndex: 2}})
	require.NoError(t, err)
	err = p.processUntilBlock(ctx, 7, []bridge2L1InfoRelation{{bridgeIndex: 2, l1InfoTreeIndex: 3}})
	require.NoError(t, err)
	l1InfoTreeIndex, err := p.getL1InfoTreeIndexByBridgeIndex(ctx, 2)
	require.NoError(t, err)
	require.Equal(t, uint32(2), l1InfoTreeIndex)
}
