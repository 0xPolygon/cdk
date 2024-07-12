package localbridgesync

import (
	"errors"
	"time"

	"github.com/0xPolygon/cdk/log"
)

var (
	retryAfterErrorPeriod      = time.Second * 10
	maxRetryAttemptsAfterError = 5
)

type LocalBridgeSync struct {
	*processor
}

func New() (*LocalBridgeSync, error) {
	// init driver, processor and downloader
	return &LocalBridgeSync{}, errors.New("not implemented")
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
