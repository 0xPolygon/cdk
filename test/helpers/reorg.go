package helpers

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/require"
)

// commitBlocks commits the specified number of blocks with the given client and waits for the specified duration after each block
func CommitBlocks(t *testing.T, client *simulated.Backend, numBlocks int, waitDuration time.Duration) {
	t.Helper()

	for i := 0; i < numBlocks; i++ {
		client.Commit()
		time.Sleep(waitDuration)
	}
}

func Reorg(t *testing.T, client *simulated.Backend, reorgSizeInBlocks uint64) {
	t.Helper()
	ctx := context.Background()
	currentBlockNum, err := client.Client().BlockNumber(ctx)
	require.NoError(t, err)

	block, err := client.Client().BlockByNumber(ctx, big.NewInt(int64(currentBlockNum-reorgSizeInBlocks)))
	log.Debugf("reorging until block %d. Current block %d (before reorg)", block.NumberU64(), currentBlockNum)
	require.NoError(t, err)
	reorgFrom := block.Hash()
	err = client.Fork(reorgFrom)
	require.NoError(t, err)
}
