package localbridgesync

import (
	"time"

	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/common"
)

var (
	retryAfterErrorPeriod      = time.Second * 10
	maxRetryAttemptsAfterError = 5
)

type LocalBridgeSync struct {
	*processor
	*driver
}

func New(
	dbPath string,
	bridge common.Address,
	rd ReorgDetectorInterface,
	l2Client EthClienter,
) (*LocalBridgeSync, error) {
	p, err := newProcessor(dbPath)
	if err != nil {
		return nil, err
	}
	dwn, err := newDownloader(bridge, l2Client)
	if err != nil {
		return nil, err
	}
	dri, err := newDriver(rd, p, dwn)
	if err != nil {
		return nil, err
	}
	return &LocalBridgeSync{p, dri}, nil
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
