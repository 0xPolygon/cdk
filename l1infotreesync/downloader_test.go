package l1infotreesync

import (
	"fmt"
	"testing"

	mocks_l1infotreesync "github.com/0xPolygon/cdk/l1infotreesync/mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBuildAppenderErrorOnBadContractAddr(t *testing.T) {
	l1Client := mocks_l1infotreesync.NewEthClienter(t)
	globalExitRoot := common.HexToAddress("0x1")
	rollupManager := common.HexToAddress("0x2")
	l1Client.EXPECT().CallContract(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("test-error"))
	flags := FlagNone
	_, err := buildAppender(l1Client, globalExitRoot, rollupManager, flags)
	require.Error(t, err)
}

func TestBuildAppenderBypassBadContractAddr(t *testing.T) {
	l1Client := mocks_l1infotreesync.NewEthClienter(t)
	globalExitRoot := common.HexToAddress("0x1")
	rollupManager := common.HexToAddress("0x2")
	l1Client.EXPECT().CallContract(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("test-error"))
	flags := FlagAllowWrongContractsAddrs
	_, err := buildAppender(l1Client, globalExitRoot, rollupManager, flags)
	require.NoError(t, err)
}
