package bridgesync

import (
	"context"
	"time"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
	tree "github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	bridgeSyncL1       = "bridgesyncl1"
	bridgeSyncL2       = "bridgesyncl2"
	downloadBufferSize = 1000
)

type ReorgDetector interface {
	sync.ReorgDetector
}

// BridgeSync manages the state of the exit tree for the bridge contract by processing Ethereum blockchain events.
type BridgeSync struct {
	processor *processor
	driver    *sync.EVMDriver

	originNetwork uint32
	blockFinality etherman.BlockNumberFinality
}

// NewL1 creates a bridge syncer that synchronizes the mainnet exit tree
func NewL1(
	ctx context.Context,
	dbPath string,
	bridge common.Address,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
	rd ReorgDetector,
	ethClient EthClienter,
	initialBlock uint64,
	waitForNewBlocksPeriod time.Duration,
	retryAfterErrorPeriod time.Duration,
	maxRetryAttemptsAfterError int,
	originNetwork uint32,
	syncFullClaims bool,
) (*BridgeSync, error) {
	return newBridgeSync(
		ctx,
		dbPath,
		bridge,
		syncBlockChunkSize,
		blockFinalityType,
		rd,
		ethClient,
		initialBlock,
		bridgeSyncL1,
		waitForNewBlocksPeriod,
		retryAfterErrorPeriod,
		maxRetryAttemptsAfterError,
		originNetwork,
		syncFullClaims,
	)
}

// NewL2 creates a bridge syncer that synchronizes the local exit tree
func NewL2(
	ctx context.Context,
	dbPath string,
	bridge common.Address,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
	rd ReorgDetector,
	ethClient EthClienter,
	initialBlock uint64,
	waitForNewBlocksPeriod time.Duration,
	retryAfterErrorPeriod time.Duration,
	maxRetryAttemptsAfterError int,
	originNetwork uint32,
	syncFullClaims bool,
) (*BridgeSync, error) {
	return newBridgeSync(
		ctx,
		dbPath,
		bridge,
		syncBlockChunkSize,
		blockFinalityType,
		rd,
		ethClient,
		initialBlock,
		bridgeSyncL2,
		waitForNewBlocksPeriod,
		retryAfterErrorPeriod,
		maxRetryAttemptsAfterError,
		originNetwork,
		syncFullClaims,
	)
}

func newBridgeSync(
	ctx context.Context,
	dbPath string,
	bridge common.Address,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
	rd ReorgDetector,
	ethClient EthClienter,
	initialBlock uint64,
	l1OrL2ID string,
	waitForNewBlocksPeriod time.Duration,
	retryAfterErrorPeriod time.Duration,
	maxRetryAttemptsAfterError int,
	originNetwork uint32,
	syncFullClaims bool,
) (*BridgeSync, error) {
	logger := log.WithFields("bridge-syncer", l1OrL2ID)
	processor, err := newProcessor(dbPath, logger)
	if err != nil {
		return nil, err
	}

	lastProcessedBlock, err := processor.GetLastProcessedBlock(ctx)
	if err != nil {
		return nil, err
	}

	if lastProcessedBlock < initialBlock {
		err = processor.ProcessBlock(ctx, sync.Block{
			Num: initialBlock,
		})
		if err != nil {
			return nil, err
		}
	}
	rh := &sync.RetryHandler{
		MaxRetryAttemptsAfterError: maxRetryAttemptsAfterError,
		RetryAfterErrorPeriod:      retryAfterErrorPeriod,
	}

	appender, err := buildAppender(ethClient, bridge, syncFullClaims)
	if err != nil {
		return nil, err
	}
	downloader, err := sync.NewEVMDownloader(
		l1OrL2ID,
		ethClient,
		syncBlockChunkSize,
		blockFinalityType,
		waitForNewBlocksPeriod,
		appender,
		[]common.Address{bridge},
		rh,
	)
	if err != nil {
		return nil, err
	}

	driver, err := sync.NewEVMDriver(rd, processor, downloader, l1OrL2ID, downloadBufferSize, rh)
	if err != nil {
		return nil, err
	}
	logger.Infof("BridgeSyncer [%s] created: dbPath: %s initialBlock: %d bridgeAddr: %s, syncFullClaims: %d,"+
		" maxRetryAttemptsAfterError: %d RetryAfterErrorPeriod: %s"+
		"syncBlockChunkSize: %d, blockFinalityType: %s waitForNewBlocksPeriod: %s",
		l1OrL2ID,
		dbPath, initialBlock, bridge.String(), syncFullClaims,
		maxRetryAttemptsAfterError, retryAfterErrorPeriod.String(),
		syncBlockChunkSize, blockFinalityType, waitForNewBlocksPeriod.String())

	return &BridgeSync{
		processor:     processor,
		driver:        driver,
		originNetwork: originNetwork,
		blockFinality: blockFinalityType,
	}, nil
}

// Start starts the synchronization process
func (s *BridgeSync) Start(ctx context.Context) {
	s.driver.Sync(ctx)
}

func (s *BridgeSync) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	if s.processor.isHalted() {
		return 0, sync.ErrInconsistentState
	}
	return s.processor.GetLastProcessedBlock(ctx)
}

func (s *BridgeSync) GetBridgeRootByHash(ctx context.Context, root common.Hash) (*tree.Root, error) {
	if s.processor.isHalted() {
		return nil, sync.ErrInconsistentState
	}
	return s.processor.exitTree.GetRootByHash(ctx, root)
}

func (s *BridgeSync) GetClaims(ctx context.Context, fromBlock, toBlock uint64) ([]Claim, error) {
	if s.processor.isHalted() {
		return nil, sync.ErrInconsistentState
	}
	return s.processor.GetClaims(ctx, fromBlock, toBlock)
}

func (s *BridgeSync) GetBridges(ctx context.Context, fromBlock, toBlock uint64) ([]Bridge, error) {
	if s.processor.isHalted() {
		return nil, sync.ErrInconsistentState
	}
	return s.processor.GetBridges(ctx, fromBlock, toBlock)
}

func (s *BridgeSync) GetBridgesPublished(ctx context.Context, fromBlock, toBlock uint64) ([]Bridge, error) {
	if s.processor.isHalted() {
		return nil, sync.ErrInconsistentState
	}
	return s.processor.GetBridgesPublished(ctx, fromBlock, toBlock)
}

func (s *BridgeSync) GetProof(ctx context.Context, depositCount uint32, localExitRoot common.Hash) (tree.Proof, error) {
	if s.processor.isHalted() {
		return tree.Proof{}, sync.ErrInconsistentState
	}
	return s.processor.exitTree.GetProof(ctx, depositCount, localExitRoot)
}

func (s *BridgeSync) GetBlockByLER(ctx context.Context, ler common.Hash) (uint64, error) {
	if s.processor.isHalted() {
		return 0, sync.ErrInconsistentState
	}
	root, err := s.processor.exitTree.GetRootByHash(ctx, ler)
	if err != nil {
		return 0, err
	}
	return root.BlockNum, nil
}

func (s *BridgeSync) GetRootByLER(ctx context.Context, ler common.Hash) (*tree.Root, error) {
	if s.processor.isHalted() {
		return nil, sync.ErrInconsistentState
	}
	root, err := s.processor.exitTree.GetRootByHash(ctx, ler)
	if err != nil {
		return root, err
	}
	return root, nil
}

// GetExitRootByIndex returns the root of the exit tree at the moment the leaf with the given index was added
func (s *BridgeSync) GetExitRootByIndex(ctx context.Context, index uint32) (tree.Root, error) {
	if s.processor.isHalted() {
		return tree.Root{}, sync.ErrInconsistentState
	}
	return s.processor.exitTree.GetRootByIndex(ctx, index)
}

// OriginNetwork returns the network ID of the origin chain
func (s *BridgeSync) OriginNetwork() uint32 {
	return s.originNetwork
}

// BlockFinality returns the block finality type
func (s *BridgeSync) BlockFinality() etherman.BlockNumberFinality {
	return s.blockFinality
}
