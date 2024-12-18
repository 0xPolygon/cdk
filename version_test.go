package cdk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	data := GetVersion()
	require.True(t, len(data.Version) > 0)
	require.True(t, len(data.GitRev) > 0)
	require.True(t, len(data.GitBranch) > 0)
	require.True(t, len(data.BuildDate) > 0)
	require.True(t, len(data.GoVersion) > 0)
	require.True(t, len(data.OS) > 0)
	require.True(t, len(data.Arch) > 0)
}
