package localbridgesync

import "errors"

type LocalBridgeSync struct {
	*processor
}

func New() (*LocalBridgeSync, error) {
	return &LocalBridgeSync{}, errors.New("not implemented")
}
