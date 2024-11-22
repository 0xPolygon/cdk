package aggsender

import (
	"fmt"

	"github.com/0xPolygon/cdk/config/types"
)

// Config is the configuration for the AggSender
type Config struct {
	// AggLayerURL is the URL of the AggLayer
	AggLayerURL string `mapstructure:"AggLayerURL"`
	// AggsenderPrivateKey is the private key which is used to sign certificates
	AggsenderPrivateKey types.KeystoreFileConfig `mapstructure:"AggsenderPrivateKey"`
	// URLRPCL2 is the URL of the L2 RPC node
	URLRPCL2 string `mapstructure:"URLRPCL2"`
	// BlockFinality indicates which finality follows AggLayer
	BlockFinality string `jsonschema:"enum=LatestBlock, enum=SafeBlock, enum=PendingBlock, enum=FinalizedBlock, enum=EarliestBlock" mapstructure:"BlockFinality"` //nolint:lll
	// EpochNotificationPercentage indicates the percentage of the epoch
	// the AggSender should send the certificate
	// 0 -> Begin
	// 50 -> Middle
	EpochNotificationPercentage uint `mapstructure:"EpochNotificationPercentage"`
	// SaveCertificatesToFilesPath if != "" tells  the AggSender to save the certificates to a file in this path
	SaveCertificatesToFilesPath string `mapstructure:"SaveCertificatesToFilesPath"`
	// DelayBeetweenRetries is the delay between retries to get the AggLayer's certificate on aggsender startup
	DelayBeetweenRetries types.Duration `mapstructure:"DelayBeetweenRetries"`
}

// String returns a string representation of the Config
func (c Config) String() string {
	return "AggLayerURL: " + c.AggLayerURL + "\n" +
		"AggsenderPrivateKeyPath: " + c.AggsenderPrivateKey.Path + "\n" +
		"URLRPCL2: " + c.URLRPCL2 + "\n" +
		"BlockFinality: " + c.BlockFinality + "\n" +
		"EpochNotificationPercentage: " + fmt.Sprintf("%d", c.EpochNotificationPercentage) + "\n" +
		"SaveCertificatesToFilesPath: " + c.SaveCertificatesToFilesPath + "\n" +
		"DelayBeetweenRetries: " + c.DelayBeetweenRetries.String() + "\n"
}
