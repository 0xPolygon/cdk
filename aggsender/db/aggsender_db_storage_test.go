package db

import (
	"context"
	"database/sql"
	"errors"
	"path"
	"testing"

	"github.com/0xPolygon/cdk/agglayer"
	"github.com/0xPolygon/cdk/aggsender/db/migrations"
	"github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func Test_SaveLastSentCertificate(t *testing.T) {
	ctx := context.Background()

	path := path.Join(t.TempDir(), "file::memory:?cache=shared")
	log.Debugf("sqlite path: %s", path)
	require.NoError(t, migrations.RunMigrations(path))

	storage, err := NewAggSenderSQLStorage(log.WithFields("aggsender-db"), path)
	require.NoError(t, err)

	t.Run("SaveLastSentCertificate", func(t *testing.T) {
		certificate := types.CertificateInfo{
			Height:           1,
			CertificateID:    common.HexToHash("0x1"),
			NewLocalExitRoot: common.HexToHash("0x2"),
			FromBlock:        1,
			ToBlock:          2,
			Status:           agglayer.Settled,
		}
		require.NoError(t, storage.SaveLastSentCertificate(ctx, certificate))

		certificateFromDB, err := storage.GetCertificateByHeight(ctx, certificate.Height)
		require.NoError(t, err)

		require.Equal(t, certificate, certificateFromDB)
	})

	t.Run("DeleteCertificate", func(t *testing.T) {
		certificate := types.CertificateInfo{
			Height:           2,
			CertificateID:    common.HexToHash("0x3"),
			NewLocalExitRoot: common.HexToHash("0x4"),
			FromBlock:        3,
			ToBlock:          4,
			Status:           agglayer.Settled,
		}
		require.NoError(t, storage.SaveLastSentCertificate(ctx, certificate))

		require.NoError(t, storage.DeleteCertificate(ctx, certificate.CertificateID))

		certificateFromDB, err := storage.GetCertificateByHeight(ctx, certificate.Height)
		require.Error(t, err)
		require.True(t, errors.Is(err, sql.ErrNoRows))
		require.Equal(t, types.CertificateInfo{}, certificateFromDB)
	})

	t.Run("GetLastSentCertificate", func(t *testing.T) {
		certificate := types.CertificateInfo{
			Height:           3,
			CertificateID:    common.HexToHash("0x5"),
			NewLocalExitRoot: common.HexToHash("0x6"),
			FromBlock:        5,
			ToBlock:          6,
			Status:           agglayer.Settled,
		}
		require.NoError(t, storage.SaveLastSentCertificate(ctx, certificate))

		certificateFromDB, err := storage.GetLastSentCertificate(ctx)
		require.NoError(t, err)

		require.Equal(t, certificate, certificateFromDB)
	})
}
