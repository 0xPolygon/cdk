package config

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"

	jRPC "github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk/aggoracle"
	"github.com/0xPolygon/cdk/aggregator"
	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/claimsponsor"
	"github.com/0xPolygon/cdk/common"
	ethermanconfig "github.com/0xPolygon/cdk/etherman/config"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/lastgersync"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/0xPolygon/cdk/sequencesender"
	"github.com/0xPolygon/zkevm-ethtx-manager/ethtxmanager"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

const (
	// FlagYes is the flag for yes.
	FlagYes = "yes"
	// FlagCfg is the flag for cfg.
	FlagCfg = "cfg"
	// FlagCustomNetwork is the flag for the custom network file.
	FlagCustomNetwork = "custom-network-file"
	// FlagAmount is the flag for amount.
	FlagAmount = "amount"
	// FlagRemoteMT is the flag for remote-merkletree.
	FlagRemoteMT = "remote-merkletree"
	// FlagComponents is the flag for components.
	FlagComponents = "components"
	// FlagHTTPAPI is the flag for http.api.
	FlagHTTPAPI = "http.api"
	// FlagKeyStorePath is the path of the key store file containing the private key
	// of the account going to sing and approve the tokens.
	FlagKeyStorePath = "key-store-path"
	// FlagPassword is the password needed to decrypt the key store
	FlagPassword = "password"
	// FlagMigrations is the flag for migrations.
	FlagMigrations = "migrations"
	// FlagOutputFile is the flag for the output file
	FlagOutputFile = "output"
	// FlagMaxAmount is the flag to avoid to use the flag FlagAmount
	FlagMaxAmount = "max-amount"

	deprecatedFieldSyncDB = "Aggregator.Synchronizer.DB is deprecated use Aggregator.Synchronizer.SQLDB instead"
)

type ForbiddenField struct {
	FieldName string
	Reason    string
}

var (
	forbiddenFieldsOnConfig = []ForbiddenField{
		{
			FieldName: "aggregator.synchronizer.db.",
			Reason:    deprecatedFieldSyncDB,
		},
	}
)

/*
Config represents the configuration of the entire Hermez Node
The file is [TOML format]
You could find some examples:
  - `config/environments/local/local.node.config.toml`: running a permisionless node
  - `config/environments/mainnet/node.config.toml`
  - `config/environments/public/node.config.toml`
  - `test/config/test.node.config.toml`: configuration for a trusted node used in CI

[TOML format]: https://en.wikipedia.org/wiki/TOML
*/
type Config struct {
	// Configuration of the etherman (client for access L1)
	Etherman ethermanconfig.Config
	// Configuration for ethereum transaction manager
	EthTxManager ethtxmanager.Config
	// Configuration of the aggregator
	Aggregator aggregator.Config
	// Configure Log level for all the services, allow also to store the logs in a file
	Log log.Config
	// Configuration of the genesis of the network. This is used to known the initial state of the network
	NetworkConfig NetworkConfig
	// Configuration of the sequence sender service
	SequenceSender sequencesender.Config

	// Common Config that affects all the services
	Common common.Config
	// Configuration of the reorg detector service to be used for the L1
	ReorgDetectorL1 reorgdetector.Config
	// Configuration of the reorg detector service to be used for the L2
	ReorgDetectorL2 reorgdetector.Config
	// Configuration of the aggOracle service
	AggOracle aggoracle.Config
	// Configuration of the L1 Info Treee Sync service
	L1InfoTreeSync l1infotreesync.Config

	// RPC is the config for the RPC server
	RPC jRPC.Config

	// ClaimSponsor is the config for the claim sponsor
	ClaimSponsor claimsponsor.EVMClaimSponsorConfig

	// BridgeL1Sync is the configuration for the synchronizer of the bridge of the L1
	BridgeL1Sync bridgesync.Config

	// BridgeL2Sync is the configuration for the synchronizer of the bridge of the L2
	BridgeL2Sync bridgesync.Config

	// LastGERSync is the config for the synchronizer in charge of syncing the last GER injected on L2.
	// Needed for the bridge service (RPC)
	LastGERSync lastgersync.Config
}

// Default parses the default configuration values.
func Default() (*Config, error) {
	var cfg Config
	viper.SetConfigType("toml")

	err := viper.ReadConfig(bytes.NewBuffer([]byte(DefaultValues)))
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&cfg, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc()))
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
func Load(ctx *cli.Context) (*Config, error) {
	configFilePath := ctx.String(FlagCfg)
	return LoadFile(configFilePath)
}

// Load loads the configuration
func LoadFile(configFilePath string) (*Config, error) {
	cfg, err := Default()
	if err != nil {
		return nil, err
	}
	expectedKeys := viper.AllKeys()
	if configFilePath != "" {
		dirName, fileName := filepath.Split(configFilePath)

		fileExtension := strings.TrimPrefix(filepath.Ext(fileName), ".")
		fileNameWithoutExtension := strings.TrimSuffix(fileName, "."+fileExtension)

		viper.AddConfigPath(dirName)
		viper.SetConfigName(fileNameWithoutExtension)
		viper.SetConfigType(fileExtension)
	}

	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("CDK")

	err = viper.ReadInConfig()
	if err != nil {
		var configNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configNotFoundError) {
			log.Error("config file not found")
		} else {
			log.Errorf("error reading config file: ", err)
			return nil, err
		}
	}

	decodeHooks := []viper.DecoderConfigOption{
		// this allows arrays to be decoded from env var separated by ",", example: MY_VAR="value1,value2,value3"
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				mapstructure.TextUnmarshallerHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
			),
		),
	}

	err = viper.Unmarshal(&cfg, decodeHooks...)
	if err != nil {
		return nil, err
	}
	if expectedKeys != nil {
		configKeys := viper.AllKeys()
		unexpectedFields := getUnexpectedFields(configKeys, expectedKeys)
		for _, field := range unexpectedFields {
			forbbidenInfo := getForbiddenField(field)
			if forbbidenInfo != nil {
				log.Warnf("forbidden field %s in config file: %s", field, forbbidenInfo.Reason)
			} else {
				log.Debugf("field %s in config file doesnt have a default value", field)
			}
		}
	}
	return cfg, nil
}

func getForbiddenField(fieldName string) *ForbiddenField {
	for _, forbiddenField := range forbiddenFieldsOnConfig {
		if forbiddenField.FieldName == fieldName || strings.HasPrefix(fieldName, forbiddenField.FieldName) {
			return &forbiddenField
		}
	}
	return nil
}

func getUnexpectedFields(keysOnFile, expectedConfigKeys []string) []string {
	wrongFields := make([]string, 0)
	for _, key := range keysOnFile {
		if !contains(expectedConfigKeys, key) {
			wrongFields = append(wrongFields, key)
		}
	}
	return wrongFields
}

func contains(keys []string, key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}
