package aggsender

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/aggsender/mocks"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExploratoryBlockNotifierPolling(t *testing.T) {
	t.Skip()
	urlRPCL1 := os.Getenv("L1URL")
	fmt.Println("URL=", urlRPCL1)
	ethClient, err := ethclient.Dial(urlRPCL1)
	require.NoError(t, err)

	sut, errSut := NewBlockNotifierPolling(ethClient,
		ConfigBlockNotifierPolling{
			BlockFinalityType: etherman.LatestBlock,
		}, log.WithFields("test", "test"), nil)
	require.NoError(t, errSut)
	sut.StartAsync(context.Background())
	ch := sut.Subscribe("test")
	for {
		select {
		case block := <-ch:
			fmt.Println(block)
		}
	}
}
func TestDelayNoPreviousBLock(t *testing.T) {
	testData := newBlockNotifierPollingTestData(t, nil)
	status := blockNotifierPollingInternalStatus{
		lastBlockSeen: 100,
	}
	delay := testData.sut.nextBlockRequestDelay(&status, nil)
	require.Equal(t, minBlockInterval, delay)
}

func TestDelayBLock(t *testing.T) {
	testData := newBlockNotifierPollingTestData(t, nil)
	pt := time.Second * 10
	status := blockNotifierPollingInternalStatus{
		lastBlockSeen:     100,
		previousBlockTime: &pt,
	}
	delay := testData.sut.nextBlockRequestDelay(&status, nil)
	require.Equal(t, minBlockInterval, delay)
}

func TestNewBlockNotifierPolling(t *testing.T) {
	testData := newBlockNotifierPollingTestData(t, nil)
	require.NotNil(t, testData.sut)
	_, err := NewBlockNotifierPolling(testData.ethClientMock, ConfigBlockNotifierPolling{
		BlockFinalityType: etherman.BlockNumberFinality("invalid"),
	}, log.WithFields("test", "test"), nil)
	require.Error(t, err)
}

func TestBlockNotifierPollingString(t *testing.T) {
	testData := newBlockNotifierPollingTestData(t, nil)
	require.NotEmpty(t, testData.sut.String())
	testData.sut.lastStatus = &blockNotifierPollingInternalStatus{
		lastBlockSeen: 100,
	}
	require.NotEmpty(t, testData.sut.String())
}

func TestBlockNotifierPollingStart(t *testing.T) {
	testData := newBlockNotifierPollingTestData(t, nil)
	ch := testData.sut.Subscribe("test")
	hdr1 := &types.Header{
		Number: big.NewInt(100),
	}
	testData.ethClientMock.EXPECT().HeaderByNumber(mock.Anything, mock.Anything).Return(hdr1, nil).Once()
	hdr2 := &types.Header{
		Number: big.NewInt(101),
	}
	testData.ethClientMock.EXPECT().HeaderByNumber(mock.Anything, mock.Anything).Return(hdr2, nil).Once()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go testData.sut.Start(ctx)
	block := <-ch
	require.NotNil(t, block)
	require.Equal(t, uint64(101), block.BlockNumber)
}

type blockNotifierPollingTestData struct {
	sut           *BlockNotifierPolling
	ethClientMock *mocks.EthClient
	ctx           context.Context
}

func newBlockNotifierPollingTestData(t *testing.T, config *ConfigBlockNotifierPolling) blockNotifierPollingTestData {
	t.Helper()
	if config == nil {
		config = &ConfigBlockNotifierPolling{
			BlockFinalityType:     etherman.LatestBlock,
			CheckNewBlockInterval: time.Second,
		}
	}
	EthClientMock := mocks.NewEthClient(t)
	logger := log.WithFields("test", "BlockNotifierPolling")
	sut, err := NewBlockNotifierPolling(EthClientMock, *config, logger, nil)
	require.NoError(t, err)
	return blockNotifierPollingTestData{
		sut:           sut,
		ethClientMock: EthClientMock,
		ctx:           context.TODO(),
	}
}