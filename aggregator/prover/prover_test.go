package prover_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/0xPolygon/cdk/aggregator/prover"
	"github.com/0xPolygon/cdk/log"
	"github.com/stretchr/testify/require"
)

const (
	dir                 = "../../test/vectors/proofs"
	stateRootStartIndex = 19
	stateRootFinalIndex = stateRootStartIndex + 8
)

type TestStateRoot struct {
	Publics []string `mapstructure:"publics"`
}

func TestCalculateStateRoots(t *testing.T) {
	var expectedStateRoots = map[string]string{
		"1871.json": "0x0ed594d8bc0bb38f3190ff25fb1e5b4fe1baf0e2e0c1d7bf3307f07a55d3a60f",
		"1872.json": "0xb6aac97ebb0eb2d4a3bdd40cfe49b6a22d42fe7deff1a8fae182a9c11cc8a7b1",
		"1873.json": "0x6f88be87a2ad2928a655bbd38c6f1b59ca8c0f53fd8e9e9d5806e90783df701f",
		"1874.json": "0x6f88be87a2ad2928a655bbd38c6f1b59ca8c0f53fd8e9e9d5806e90783df701f",
		"1875.json": "0xf4a439c5642a182d9e27c8ab82c64b44418ba5fa04c175a013bed452c19908c9"}

	// Read all files in the directory
	files, err := os.ReadDir(dir)
	require.NoError(t, err)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Read the file
		data, err := os.ReadFile(fmt.Sprintf("%s/%s", dir, file.Name()))
		require.NoError(t, err)

		// Get the state root from the batch proof
		fileStateRoot, err := prover.GetSanityCheckHashFromProof(log.GetDefaultLogger(), string(data), stateRootStartIndex, stateRootFinalIndex)
		require.NoError(t, err)

		// Get the expected state root
		expectedStateRoot, ok := expectedStateRoots[file.Name()]
		require.True(t, ok, "Expected state root not found")

		// Compare the state roots
		require.Equal(t, expectedStateRoot, fileStateRoot.String(), "State roots do not match")
	}
}
