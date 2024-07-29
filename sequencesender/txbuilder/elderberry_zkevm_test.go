package txbuilder

import (
	"strings"
	"testing"

	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestElderberryZkevmName(t *testing.T) {
	zkevmContract := contracts.RollupElderberryType{}
	opts := bind.TransactOpts{}
	sender := common.Address{}
	sut := NewTxBuilderElderberryZKEVM(zkevmContract, opts, sender, 100)
	require.NotNil(t, sut)
	require.True(t, strings.Contains(sut.String(), "Elderberry"))
	require.True(t, strings.Contains(sut.String(), "ZKEVM"))
}

func TestElderberryZkevmNewSequence(t *testing.T) {
	zkevmContract := contracts.RollupElderberryType{}
	opts := bind.TransactOpts{}
	sender := common.Address{}
	sut := NewTxBuilderElderberryZKEVM(zkevmContract, opts, sender, 100)
	require.NotNil(t, sut)
	seq, err := sut.NewSequence(nil, common.Address{})
	require.NoError(t, err)
	require.NotNil(t, seq)
}
