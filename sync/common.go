package sync

import (
	"log"
	"sync"
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

func UnhaltIfAffectedRows(halted *bool, haltedReason *string, mu *sync.RWMutex, rowsAffected uint64) {
	if rowsAffected > 0 {
		mu.Lock()
		defer mu.Unlock()
		*halted = false
		*haltedReason = ""
	}
}
