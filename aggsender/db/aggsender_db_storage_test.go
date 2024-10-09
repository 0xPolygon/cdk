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
		require.NoError(t, err)
		require.Equal(t, types.CertificateInfo{}, certificateFromDB)

		// try getting a certificate that exists
		certificate := types.CertificateInfo{
			Height:           3,
			CertificateID:    common.HexToHash("0x5"),
			NewLocalExitRoot: common.HexToHash("0x6"),
			FromBlock:        5,
			ToBlock:          6,
			Status:           agglayer.Pending,
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

	t.Run("GetCertificatesByStatus", func(t *testing.T) {
		// Insert some certificates with different statuses
		certificates := []*types.CertificateInfo{
			{
				Height:           7,
				CertificateID:    common.HexToHash("0x7"),
				NewLocalExitRoot: common.HexToHash("0x8"),
				FromBlock:        7,
				ToBlock:          8,
				Status:           agglayer.Settled,
			},
			{
				Height:           9,
				CertificateID:    common.HexToHash("0x9"),
				NewLocalExitRoot: common.HexToHash("0xA"),
				FromBlock:        9,
				ToBlock:          10,
				Status:           agglayer.Pending,
			},
			{
				Height:           11,
				CertificateID:    common.HexToHash("0xB"),
				NewLocalExitRoot: common.HexToHash("0xC"),
				FromBlock:        11,
				ToBlock:          12,
				Status:           agglayer.InError,
			},
		}

		for _, cert := range certificates {
			require.NoError(t, storage.SaveLastSentCertificate(ctx, *cert))
		}

		// Test fetching certificates with status Settled
		statuses := []agglayer.CertificateStatus{agglayer.Settled}
		certificatesFromDB, err := storage.GetCertificatesByStatus(ctx, statuses)
		require.NoError(t, err)
		require.Len(t, certificatesFromDB, 1)
		require.ElementsMatch(t, []*types.CertificateInfo{certificates[0]}, certificatesFromDB)

		// Test fetching certificates with status Pending
		statuses = []agglayer.CertificateStatus{agglayer.Pending}
		certificatesFromDB, err = storage.GetCertificatesByStatus(ctx, statuses)
		require.NoError(t, err)
		require.Len(t, certificatesFromDB, 1)
		require.ElementsMatch(t, []*types.CertificateInfo{certificates[1]}, certificatesFromDB)

		// Test fetching certificates with status InError
		statuses = []agglayer.CertificateStatus{agglayer.InError}
		certificatesFromDB, err = storage.GetCertificatesByStatus(ctx, statuses)
		require.NoError(t, err)
		require.Len(t, certificatesFromDB, 1)
		require.ElementsMatch(t, []*types.CertificateInfo{certificates[2]}, certificatesFromDB)

		// Test fetching certificates with status InError and Pending
		statuses = []agglayer.CertificateStatus{agglayer.InError, agglayer.Pending}
		certificatesFromDB, err = storage.GetCertificatesByStatus(ctx, statuses)
		require.NoError(t, err)
		require.Len(t, certificatesFromDB, 2)
		require.ElementsMatch(t, []*types.CertificateInfo{certificates[1], certificates[2]}, certificatesFromDB)

		require.NoError(t, storage.clean())
	})
}
