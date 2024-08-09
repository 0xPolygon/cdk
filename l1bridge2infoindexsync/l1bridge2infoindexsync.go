package l1bridge2infoindexsync

import (
	"context"
	"time"

	"github.com/0xPolygon/cdk/bridgesync"
	configTypes "github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum"
)

type L1Bridge2InfoIndexSync struct {
	processor *processor
	driver    *driver
}

type Config struct {
	// DBPath path of the DB
	DBPath string `mapstructure:"DBPath"`
	// RetryAfterErrorPeriod is the time that will be waited when an unexpected error happens before retry
	RetryAfterErrorPeriod configTypes.Duration `mapstructure:"RetryAfterErrorPeriod"`
	// MaxRetryAttemptsAfterError is the maximum number of consecutive attempts that will happen before panicing.
	// Any number smaller than zero will be considered as unlimited retries
	MaxRetryAttemptsAfterError int `mapstructure:"MaxRetryAttemptsAfterError"`
	// WaitForSyncersPeriod time that will be waited when the synchronizer has reached the latest state
	WaitForSyncersPeriod configTypes.Duration `mapstructure:"WaitForSyncersPeriod"`
}

func New(
	dbPath string,
	l1Bridge *bridgesync.BridgeSync,
	l1Info *l1infotreesync.L1InfoTreeSync,
	l1Client ethereum.ChainReader,
	retryAfterErrorPeriod time.Duration,
	maxRetryAttemptsAfterError int,
	waitForSyncersPeriod time.Duration,
) (*L1Bridge2InfoIndexSync, error) {
	dwn := newDownloader(l1Bridge, l1Info, l1Client)

	prc, err := newProcessor(dbPath)
	if err != nil {
		return nil, err
	}

	rh := &sync.RetryHandler{
		RetryAfterErrorPeriod:      retryAfterErrorPeriod,
		MaxRetryAttemptsAfterError: maxRetryAttemptsAfterError,
	}
	drv := newDriver(dwn, prc, rh, waitForSyncersPeriod)

	return &L1Bridge2InfoIndexSync{
		driver:    drv,
		processor: prc,
	}, nil
}

func (s *L1Bridge2InfoIndexSync) Start(ctx context.Context) {
	s.driver.sync(ctx)
}

func (s *L1Bridge2InfoIndexSync) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	lpb, _, err := s.processor.GetLastProcessedBlockAndL1InfoTreeIndex(ctx)
	return lpb, err
}

func (s *L1Bridge2InfoIndexSync) GetL1InfoTreeIndexByDepositCount(ctx context.Context, depositCount uint32) (uint32, error) {
	return s.processor.getL1InfoTreeIndexByBrdigeIndex(ctx, depositCount)
}
