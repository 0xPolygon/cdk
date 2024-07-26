package aggoracle

import (
	"github.com/0xPolygon/cdk/aggoracle/chaingersender"
	"github.com/0xPolygon/cdk/config/types"
)

type TargetChainType string

const (
	EVMChain TargetChainType = "EVM"
)

var (
	SupportedChainTypes = []TargetChainType{EVMChain}
)

type Config struct {
	TargetChainType TargetChainType `mapstructure:"TargetChainType"`
	URLRPCL1        string          `mapstructure:"URLRPCL1"`
	// TODO: BlockFinality doesnt work as per the jsonschema
	BlockFinality     string                   `jsonschema:"enum=latest,enum=safe, enum=pending, enum=finalized" mapstructure:"BlockFinality"`
	WaitPeriodNextGER types.Duration           `mapstructure:"WaitPeriodNextGER"`
	EVMSender         chaingersender.EVMConfig `mapstructure:"EVMSender"`
}
