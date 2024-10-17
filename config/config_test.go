package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDeafaultConfig(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ut_config")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.Write([]byte(DefaultValues))
	require.NoError(t, err)
	cfg, err := LoadFile(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

const configWithUnexpectedFields = `
[UnknownField]
Field = "value"
`

func TestLoadConfigWithUnexpectedFields(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ut_config")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.Write([]byte(configWithUnexpectedFields))
	require.NoError(t, err)
	cfg, err := LoadFile(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestLoadConfigWithForbiddenFields(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{
			name: "[Aggregator.Synchronizer] DB",
			input: `[aggregator.synchronizer.db]
						name = "value"`,
		},
		{
			name: "[SequenceSender.EthTxManager] PersistenceFilename",
			input: `[SequenceSender.EthTxManager]
						PersistenceFilename = "foo.json"`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "ut_config")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())
			_, err = tmpFile.Write([]byte(c.input))
			require.NoError(t, err)
			cfg, err := LoadFile(tmpFile.Name())
			require.NoError(t, err)
			require.NotNil(t, cfg)
		})
	}
}
