package sourcemanager

import (
	"github.com/0xPolygon/cdk/source"
)

type SourceManager struct {
	DataStream source.Source
	Avail      source.Source
}
