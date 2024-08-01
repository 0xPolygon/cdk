package sync

import (
	"log"
	"time"
)

type RetryHandler struct {
	RetryAfterErrorPeriod      time.Duration
	MaxRetryAttemptsAfterError int
}

func (h *RetryHandler) Handle(funcName string, attempts int) {
	if h.MaxRetryAttemptsAfterError > -1 && attempts >= h.MaxRetryAttemptsAfterError {
		log.Fatalf(
			"%s failed too many times (%d)",
			funcName, h.MaxRetryAttemptsAfterError,
		)
	}
	time.Sleep(h.RetryAfterErrorPeriod)
}
