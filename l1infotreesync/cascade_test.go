package l1infotreesync

import (
	"context"
	"path"
	"strconv"
	"testing"

	"github.com/0xPolygon/cdk/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/russross/meddler"
	"github.com/stretchr/testify/require"
)

func TestCascade(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "TestCascade.sqlite")
	p, err := newProcessor(dbPath)
	require.NoError(t, err)
	ctx := context.Background()
	for i := 1; i < 10_000; i++ {
		// insert block and info
		tx, err := db.NewTx(ctx, p.db)
		require.NoError(t, err)
		_, err = tx.Exec(`INSERT INTO block (num) VALUES ($1)`, i)
		require.NoError(t, err)
		info := &L1InfoTreeLeaf{
			BlockNumber:       uint64(i),
			BlockPosition:     0,
			L1InfoTreeIndex:   0,
			PreviousBlockHash: common.Hash{},
			Timestamp:         0,
			MainnetExitRoot:   common.Hash{},
			RollupExitRoot:    common.Hash{},
			GlobalExitRoot:    common.HexToHash(strconv.Itoa(i)), // ger needs to be unique
		}
		info.Hash = info.hash()
		err = meddler.Insert(tx, "l1info_leaf", info)
		require.NoError(t, err)
		require.NoError(t, tx.Commit())

		// simulate reorg every 3 iterations
		if i%3 == 0 {
			tx, err = db.NewTx(ctx, p.db)
			require.NoError(t, err)
			_, err = tx.Exec(`delete from block where num >= $1;`, i)
			require.NoError(t, err)
			require.NoError(t, tx.Commit())
			// assert that info table is empty
			info, err = p.GetLastInfo()
			require.NoError(t, err)
			require.Equal(t, info.BlockNumber, uint64(i-1))
		}
	}
}
