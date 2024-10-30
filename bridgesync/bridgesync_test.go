package bridgesync_test

import (
	"context"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/bridgesync"
	mocksbridgesync "github.com/0xPolygon/cdk/bridgesync/mocks"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for the interfaces
type MockEthClienter struct {
	mock.Mock
}

type MockBridgeContractor struct {
	mock.Mock
}

func TestNewLx(t *testing.T) {
	ctx := context.Background()
	dbPath := "test_db_path"
	bridge := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	syncBlockChunkSize := uint64(100)
	blockFinalityType := etherman.SafeBlock
	initialBlock := uint64(0)
	waitForNewBlocksPeriod := time.Second * 10
	retryAfterErrorPeriod := time.Second * 5
	maxRetryAttemptsAfterError := 3
	originNetwork := uint32(1)

	mockEthClient := mocksbridgesync.NewEthClienter(t)
	mockReorgDetector := mocksbridgesync.NewReorgDetector(t)

	mockReorgDetector.EXPECT().Subscribe(mock.Anything).Return(nil, nil)

	bridgeSync, err := bridgesync.NewL1(
		ctx,
		dbPath,
		bridge,
		syncBlockChunkSize,
		blockFinalityType,
		mockReorgDetector,
		mockEthClient,
		initialBlock,
		waitForNewBlocksPeriod,
		retryAfterErrorPeriod,
		maxRetryAttemptsAfterError,
		originNetwork,
	)

	assert.NoError(t, err)
	assert.NotNil(t, bridgeSync)
	assert.Equal(t, originNetwork, bridgeSync.OriginNetwork())
	assert.Equal(t, blockFinalityType, bridgeSync.BlockFinality())

	bridgeSyncL2, err := bridgesync.NewL2(
		ctx,
		dbPath,
		bridge,
		syncBlockChunkSize,
		blockFinalityType,
		mockReorgDetector,
		mockEthClient,
		initialBlock,
		waitForNewBlocksPeriod,
		retryAfterErrorPeriod,
		maxRetryAttemptsAfterError,
		originNetwork,
	)

	assert.NoError(t, err)
	assert.NotNil(t, bridgeSync)
	assert.Equal(t, originNetwork, bridgeSyncL2.OriginNetwork())
	assert.Equal(t, blockFinalityType, bridgeSyncL2.BlockFinality())
}
