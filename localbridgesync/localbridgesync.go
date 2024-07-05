package localbridgesync

import (
	"errors"
	"time"
)

const (
	retryAfterErrorPeriod = time.Second * 10
)

type LocalBridgeSync struct {
	*processor
}

func New() (*LocalBridgeSync, error) {
	// init driver, processor and downloader
	return &LocalBridgeSync{}, errors.New("not implemented")
}
