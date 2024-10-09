package db

import (
	"context"
	"path"
	"testing"

	"github.com/0xPolygon/cdk/agglayer"
	"github.com/0xPolygon/cdk/aggsender/db/migrations"
	"github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func Test_Storage(t *testing.T) {
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
		require.NoError(t, storage.clean())
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
		require.ErrorIs(t, err, db.ErrNotFound)
		require.Equal(t, types.CertificateInfo{}, certificateFromDB)
		require.NoError(t, storage.clean())
	})

	t.Run("GetLastSentCertificate", func(t *testing.T) {
		// try getting a certificate that doesn't exist
		certificateFromDB, err := storage.GetLastSentCertificate(ctx)
		require.ErrorIs(t, err, db.ErrNotFound)
		require.Equal(t, types.CertificateInfo{}, certificateFromDB)

		// try getting a certificate that exists
		certificate := types.CertificateInfo{
			Height:           3,
			CertificateID:    common.HexToHash("0x5"),
			NewLocalExitRoot: common.HexToHash("0x6"),
			FromBlock:        5,
			ToBlock:          6,
			Status:           agglayer.Settled,
		}
		require.NoError(t, storage.SaveLastSentCertificate(ctx, certificate))

		certificateFromDB, err = storage.GetLastSentCertificate(ctx)
		require.NoError(t, err)

		require.Equal(t, certificate, certificateFromDB)
		require.NoError(t, storage.clean())
	})

	t.Run("GetCertificateByHeight", func(t *testing.T) {
		// try getting height 0
		certificateFromDB, err := storage.GetCertificateByHeight(ctx, 0)
		require.NoError(t, err)
		require.Equal(t, types.CertificateInfo{}, certificateFromDB)

		// try getting a certificate that doesn't exist
		certificateFromDB, err = storage.GetCertificateByHeight(ctx, 4)
		require.ErrorIs(t, err, db.ErrNotFound)
		require.Equal(t, types.CertificateInfo{}, certificateFromDB)

		// try getting a certificate that exists
		certificate := types.CertificateInfo{
			Height:           11,
			CertificateID:    common.HexToHash("0x17"),
			NewLocalExitRoot: common.HexToHash("0x18"),
			FromBlock:        17,
			ToBlock:          18,
			Status:           agglayer.Pending,
		}
		require.NoError(t, storage.SaveLastSentCertificate(ctx, certificate))

		certificateFromDB, err = storage.GetCertificateByHeight(ctx, certificate.Height)
		require.NoError(t, err)

		require.Equal(t, certificate, certificateFromDB)
		require.NoError(t, storage.clean())
	})
}
