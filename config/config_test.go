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
	tmpFile.Write([]byte(DefaultValues))
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
	tmpFile.Write([]byte(configWithUnexpectedFields))
	cfg, err := LoadFile(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

const configWithForbiddenFields = `
[aggregator.synchronizer.db]
name = "value"
`

func TestLoadConfigWithForbiddenFields(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ut_config")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte(configWithForbiddenFields))
	cfg, err := LoadFile(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, cfg)
}
