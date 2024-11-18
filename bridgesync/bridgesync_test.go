package bridgesync

import (
	"context"
	"errors"
	"path"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/bridgesync"
	mocksbridgesync "github.com/0xPolygon/cdk/bridgesync/mocks"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	dbPath := path.Join(t.TempDir(), "TestNewLx.sqlite")
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
		false,
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
		false,
	)

	assert.NoError(t, err)
	assert.NotNil(t, bridgeSync)
	assert.Equal(t, originNetwork, bridgeSyncL2.OriginNetwork())
	assert.Equal(t, blockFinalityType, bridgeSyncL2.BlockFinality())
}

func TestGetLastProcessedBlock(t *testing.T) {
	s := BridgeSync{processor: &processor{halted: true}}
	_, err := s.GetLastProcessedBlock(context.Background())
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetBridgeRootByHash(t *testing.T) {
	s := BridgeSync{processor: &processor{halted: true}}
	_, err := s.GetBridgeRootByHash(context.Background(), common.Hash{})
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetBridges(t *testing.T) {
	s := BridgeSync{processor: &processor{halted: true}}
	_, err := s.GetBridges(context.Background(), 0, 0)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetProof(t *testing.T) {
	s := BridgeSync{processor: &processor{halted: true}}
	_, err := s.GetProof(context.Background(), 0, common.Hash{})
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetBlockByLER(t *testing.T) {
	s := BridgeSync{processor: &processor{halted: true}}
	_, err := s.GetBlockByLER(context.Background(), common.Hash{})
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetRootByLER(t *testing.T) {
	s := BridgeSync{processor: &processor{halted: true}}
	_, err := s.GetRootByLER(context.Background(), common.Hash{})
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetExitRootByIndex(t *testing.T) {
	s := BridgeSync{processor: &processor{halted: true}}
	_, err := s.GetExitRootByIndex(context.Background(), 0)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}
