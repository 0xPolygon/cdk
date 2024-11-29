package migrations

import (
	"context"
	"path"
	"testing"

	"github.com/0xPolygon/cdk/db"
	"github.com/stretchr/testify/require"
)

func Test001(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "bridgesyncTest001.sqlite")

	err := RunMigrations(dbPath)
	require.NoError(t, err)
	db, err := db.NewSQLiteDB(dbPath)
	require.NoError(t, err)

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	_, err = tx.Exec(`
		INSERT INTO block (num) VALUES (1);

		INSERT INTO bridge (
			block_num,
			block_pos,
			leaf_type,
			origin_network,
			origin_address,
			destination_network,
			destination_address,
			amount,
			metadata,
			deposit_count
		) VALUES (1, 0, 0, 0, '0x0000', 0, '0x0000', 0, NULL, 0);

		INSERT INTO claim (
			block_num,
			block_pos,
    		global_index,
			origin_network,
			origin_address,
			destination_address,
			amount,
			proof_local_exit_root,
			proof_rollup_exit_root,
			mainnet_exit_root,
			rollup_exit_root,
			global_exit_root,
			destination_network,
			metadata,
			is_message
		) VALUES (1, 0, 0, 0, '0x0000', '0x0000', 0, '0x000,0x000', '0x000,0x000', '0x000', '0x000', '0x0', 0, NULL, FALSE);
	`)
	require.NoError(t, err)
	err = tx.Commit()
	require.NoError(t, err)
}
