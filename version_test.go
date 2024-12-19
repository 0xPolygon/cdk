package cdk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	data := GetVersion()
	require.NotEmpty(t, data.Version)
	require.NotEmpty(t, data.GitRev)
	require.NotEmpty(t, data.GitBranch)
	require.NotEmpty(t, data.BuildDate)
	require.NotEmpty(t, data.GoVersion)
	require.NotEmpty(t, data.OS)
	require.NotEmpty(t, data.Arch)
}
