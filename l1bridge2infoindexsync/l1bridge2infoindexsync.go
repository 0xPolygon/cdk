package l1bridge2infoindexsync

import (
	"context"
	"time"

	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum"
)

type L1Bridge2InfoIndexSync struct {
	processor *processor
	driver    *driver
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

// GetLastProcessedBlock retrieves the last processed block number by the processor.
func (s *L1Bridge2InfoIndexSync) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	lpb, _, err := s.processor.GetLastProcessedBlockAndL1InfoTreeIndex(ctx)

	return lpb, err
}

// GetL1InfoTreeIndexByDepositCount retrieves the L1 Info Tree index for a given deposit count.
func (s *L1Bridge2InfoIndexSync) GetL1InfoTreeIndexByDepositCount(
	ctx context.Context, depositCount uint32,
) (uint32, error) {
	return s.processor.getL1InfoTreeIndexByBridgeIndex(ctx, depositCount)
}
