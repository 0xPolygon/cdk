package l1infotreesync

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk/l1infotreesync/mocks"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBuildAppenderErrorOnBadContractAddr(t *testing.T) {
	l1Client := mocks.NewEthClienter(t)
	globalExitRoot := common.HexToAddress("0x1")
	rollupManager := common.HexToAddress("0x2")
	l1Client.EXPECT().CallContract(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("test-error"))
	flags := FlagNone
	_, err := buildAppender(l1Client, globalExitRoot, rollupManager, flags)
	require.Error(t, err)
}

func TestBuildAppenderBypassBadContractAddr(t *testing.T) {
	l1Client := mocks.NewEthClienter(t)
	globalExitRoot := common.HexToAddress("0x1")
	rollupManager := common.HexToAddress("0x2")
	l1Client.EXPECT().CallContract(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("test-error"))
	flags := FlagAllowWrongContractsAddrs
	_, err := buildAppender(l1Client, globalExitRoot, rollupManager, flags)
	require.NoError(t, err)
}

func TestBuildAppenderVerifiedContractAddr(t *testing.T) {
	l1Client := mocks.NewEthClienter(t)
	globalExitRoot := common.HexToAddress("0x1")
	rollupManager := common.HexToAddress("0x2")

	smcAbi, err := abi.JSON(strings.NewReader(polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2ABI))
	require.NoError(t, err)
	bigInt := big.NewInt(1)
	returnGER, err := smcAbi.Methods["depositCount"].Outputs.Pack(bigInt)
	require.NoError(t, err)
	l1Client.EXPECT().CallContract(mock.Anything, mock.Anything, mock.Anything).Return(returnGER, nil).Once()
	v := common.HexToAddress("0x1234")
	returnRM, err := smcAbi.Methods["bridgeAddress"].Outputs.Pack(v)
	require.NoError(t, err)
	l1Client.EXPECT().CallContract(mock.Anything, mock.Anything, mock.Anything).Return(returnRM, nil).Once()
	flags := FlagNone
	_, err = buildAppender(l1Client, globalExitRoot, rollupManager, flags)
	require.NoError(t, err)
}
