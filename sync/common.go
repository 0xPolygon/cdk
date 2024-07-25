package sync

import (
	"log"
	"time"
)

var (
	RetryAfterErrorPeriod      = time.Second * 10
	MaxRetryAttemptsAfterError = 5
)

func RetryHandler(funcName string, attempts int) {
	if attempts >= MaxRetryAttemptsAfterError {
		log.Fatalf(
			"%s failed too many times (%d)",
			funcName, MaxRetryAttemptsAfterError,
		)
	}
	time.Sleep(RetryAfterErrorPeriod)
}
