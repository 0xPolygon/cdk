package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAggsenderStatusSetLastError(t *testing.T) {
	sut := AggsenderStatus{}
	sut.SetLastError(nil)
	require.Equal(t, "", sut.LastError)
	sut.SetLastError(errors.New("error"))
	require.Equal(t, "error", sut.LastError)
}
