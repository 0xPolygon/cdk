package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXxx(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ut_config")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte(DefaultValues))
	cfg, err := LoadFile(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, cfg)
}
