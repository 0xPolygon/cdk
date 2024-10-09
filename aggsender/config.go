package aggsender

import (
	"github.com/0xPolygon/cdk/config/types"
)

// Config is the configuration for the AggSender
type Config struct {
	DBPath                  string                   `mapstructure:"DBPath"`
	AggLayerURL             string                   `mapstructure:"AggLayerURL"`
	CertificateSendInterval types.Duration           `mapstructure:"CertificateSendInterval"`
	SequencerPrivateKey     types.KeystoreFileConfig `mapstructure:"SequencerPrivateKey"`
	URLRPCL2                string                   `mapstructure:"URLRPCL2"`
}
