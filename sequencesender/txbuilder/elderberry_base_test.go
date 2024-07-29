package txbuilder

import (
	"testing"

	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestElderberryNewSequence(t *testing.T) {
	zkevmContract := contracts.RollupElderberryType{}
	opts := bind.TransactOpts{}
	sut := NewTxBuilderElderberryBase(zkevmContract, opts)
	require.NotNil(t, sut)
	seq, err := sut.NewSequence(nil, common.Address{})
	require.NotNil(t, seq)
	require.NoError(t, err)
}
