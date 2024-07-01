package txbuilder

import "context"

type Selector interface {
	Get(ctx context.Context) (string, error)
}

type SelectorPerForkID struct {
}

func NewSelectorPerForkID() Selector {
	return &SelectorPerForkID{}
}

func (s *SelectorPerForkID) Get(ctx context.Context) (string, error) {
	//TODO: Implement this
	return "banana", nil
}
