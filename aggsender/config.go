package aggsender

import (
	"github.com/0xPolygon/cdk/config/types"
)

// Config is the configuration for the AggSender
type Config struct {
	// StoragePath is the path of the sqlite db on which the AggSender will store the data
	StoragePath string `mapstructure:"StoragePath"`
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
	// SaveCertificatesToFilesPath if != "" tells  the AggSender to save the certificates to a file in this path
	SaveCertificatesToFilesPath string `mapstructure:"SaveCertificatesToFilesPath"`
}

// String returns a string representation of the Config
func (c Config) String() string {
	return "AggSender Config:\n" +
		"StoragePath: " + c.StoragePath + "\n" +
		"AggLayerURL: " + c.AggLayerURL + "\n" +
		"BlockGetInterval: " + c.BlockGetInterval.String() + "\n" +
		"CheckSettledInterval: " + c.CheckSettledInterval.String() + "\n" +
		"AggsenderPrivateKeyPath: " + c.AggsenderPrivateKey.Path + "\n" +
		"AggsenderPrivateKeyPassword: " + c.AggsenderPrivateKey.Password + "\n" +
		"URLRPCL2: " + c.URLRPCL2 + "\n" +
		"SaveCertificatesToFilesPath: " + c.SaveCertificatesToFilesPath + "\n"
}
