package localbridgesync

import "errors"

type LocalBridgeSync struct {
	*processor
}

func New() (*LocalBridgeSync, error) {
	// init driver, processor and downloader
	return &LocalBridgeSync{}, errors.New("not implemented")
}
