package aggsender

import (
	"github.com/0xPolygon/cdk/config/types"
)

// Config is the configuration for the AggSender
type Config struct {
	// DBPath is the path of the sqlite db on which the AggSender will store the data
	DBPath string `mapstructure:"DBPath"`
	// AggLayerURL is the URL of the AggLayer
	AggLayerURL string `mapstructure:"AggLayerURL"`
	// BlockGetInterval is the interval at which the AggSender will get the blocks from L1
	BlockGetInterval types.Duration `mapstructure:"BlockGetInterval"`
	// CheckSettledInterval is the interval at which the AggSender will check if the blocks are settled
	CheckSettledInterval types.Duration `mapstructure:"CheckSettledInterval"`
	// AggsenderPrivateKey is the private key which is used to sign certificates
	AggsenderPrivateKey types.KeystoreFileConfig `mapstructure:"AggsenderPrivateKey"`
	// URLRPCL2 is the URL of the L2 RPC node
	URLRPCL2 string `mapstructure:"URLRPCL2"`
	// EpochSize is the size of the epoch on L1 (configured on agglayer) in blocks
	EpochSize uint64 `mapstructure:"EpochSize"`
	// BlocksBeforeEpochEnding indicates how many blocks before the epoch ending
	// the AggSender should send the certificate
	BlocksBeforeEpochEnding uint64 `mapstructure:"BlocksBeforeEpochEnding"`
}
