package l1infotreesync

import (
	"context"
	"time"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/common"
)

var (
	retryAfterErrorPeriod      = time.Second * 10
	maxRetryAttemptsAfterError = 5
)

type L1InfoTreeSync struct {
	*processor
	*driver
}

func New(
	ctx context.Context,
	dbPath string,
	globalExitRoot common.Address,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
	rd ReorgDetector,
	l1Client EthClienter,
	treeHeight uint8,
) (*L1InfoTreeSync, error) {
	p, err := newProcessor(ctx, dbPath, treeHeight)
	if err != nil {
		return nil, err
	}
	dwn, err := newDownloader(globalExitRoot, l1Client, syncBlockChunkSize, blockFinalityType)
	if err != nil {
		return nil, err
	}
	dri, err := newDriver(rd, p, dwn)
	if err != nil {
		return nil, err
	}
	return &L1InfoTreeSync{p, dri}, nil
}

func retryHandler(funcName string, attempts int) {
	if attempts >= maxRetryAttemptsAfterError {
		log.Fatalf(
			"%s failed too many times (%d)",
			funcName, maxRetryAttemptsAfterError,
		)
	}
	time.Sleep(retryAfterErrorPeriod)
}
