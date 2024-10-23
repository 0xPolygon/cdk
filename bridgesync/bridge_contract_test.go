package bridgesync

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestLastUpdatedDepositCount(t *testing.T) {
	_, err := NewBridgeContract(common.Address{}, nil)
	require.NoError(t, err)
}
