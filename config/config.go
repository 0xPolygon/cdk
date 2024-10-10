package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
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
	// FlagSaveConfigPath is the flag to save the final configuration file
	FlagSaveConfigPath = "save-config-path"

	deprecatedFieldSyncDB = "Aggregator.Synchronizer.DB is deprecated. Use Aggregator.Synchronizer.SQLDB instead."

	deprecatedFieldPersistenceFilename = "EthTxManager.PersistenceFilename is deprecated." +
		" Use EthTxManager.StoragePath instead."

	EnvVarPrefix       = "CDK"
	ConfigType         = "toml"
	SaveConfigFileName = "cdk_config.toml"

	DefaultCreationFilePermissions = os.FileMode(0600)
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
		{
			FieldName: "sequencesender.ethtxmanager.persistencefilename",
			Reason:    deprecatedFieldPersistenceFilename,
		},
		{
			FieldName: "aggregator.ethtxmanager.persistencefilename",
			Reason:    deprecatedFieldPersistenceFilename,
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

// Load loads the configuration
func Load(ctx *cli.Context) (*Config, error) {
	configFilePath := ctx.StringSlice(FlagCfg)
	filesData, err := readFiles(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading files:  Err:%w", err)
	}
	saveConfigPath := ctx.String(FlagSaveConfigPath)
	return LoadFile(filesData, saveConfigPath)
}

func readFiles(files []string) ([]FileData, error) {
	result := make([]FileData, 0)
	for _, file := range files {
		fileContent, err := readFileToString(file)
		if err != nil {
			return nil, fmt.Errorf("error reading file content: %s. Err:%w", file, err)
		}
		fileExtension := getFileExtension(file)
		if fileExtension != ConfigType {
			fileContent, err = convertFileToToml(fileContent, fileExtension)
			if err != nil {
				return nil, fmt.Errorf("error converting file: %s from %s to TOML. Err:%w", file, fileExtension, err)
			}
		}
		result = append(result, FileData{Name: file, Content: fileContent})
	}
	return result, nil
}

func getFileExtension(fileName string) string {
	return fileName[strings.LastIndex(fileName, ".")+1:]
}

// Load loads the configuration
func LoadFileFromString(configFileData string, configType string) (*Config, error) {
	cfg := &Config{}
	expectedKeys := viper.AllKeys()
	err := loadString(cfg, configFileData, configType, true, EnvVarPrefix, &expectedKeys)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func SaveConfigToString(cfg Config) (string, error) {
	b, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Load loads the configuration
func LoadFile(files []FileData, saveConfigPath string) (*Config, error) {
	fileData := make([]FileData, 0)
	fileData = append(fileData, FileData{Name: "default_vars", Content: DefaultVars})
	fileData = append(fileData, FileData{Name: "default_values", Content: DefaultValues})
	fileData = append(fileData, files...)

	merger := NewConfigRender(fileData, EnvVarPrefix)

	renderedCfg, err := merger.Render()
	if err != nil {
		return nil, err
	}
	if saveConfigPath != "" {
		fullPath := saveConfigPath + "/" + SaveConfigFileName
		err = os.WriteFile(fullPath, []byte(renderedCfg), DefaultCreationFilePermissions)
		if err != nil {
			err = fmt.Errorf("error writing config file: %s. Err: %w", fullPath, err)
			log.Error(err)
			return nil, err
		}
	}
	cfg, err := LoadFileFromString(renderedCfg, ConfigType)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Load loads the configuration
func loadString(cfg *Config, configData string, configType string,
	allowEnvVars bool, envPrefix string, expectedKeys *[]string) error {
	viper.SetConfigType(configType)
	if allowEnvVars {
		replacer := strings.NewReplacer(".", "_")
		viper.SetEnvKeyReplacer(replacer)
		viper.SetEnvPrefix(envPrefix)
		viper.AutomaticEnv()
	}
	err := viper.ReadConfig(bytes.NewBuffer([]byte(configData)))
	if err != nil {
		return err
	}
	decodeHooks := []viper.DecoderConfigOption{
		// this allows arrays to be decoded from env var separated by ",", example: MY_VAR="value1,value2,value3"
		viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
			mapstructure.TextUnmarshallerHookFunc(), mapstructure.StringToSliceHookFunc(","))),
	}

	err = viper.Unmarshal(&cfg, decodeHooks...)
	if err != nil {
		return err
	}
	// TODO: this is a non-sense. I need a place to get the expected values

	if expectedKeys != nil {
		configKeys := viper.AllKeys()
		unexpectedFields := getUnexpectedFields(configKeys, *expectedKeys)
		for _, field := range unexpectedFields {
			forbbidenInfo := getForbiddenField(field)
			if forbbidenInfo != nil {
				log.Warnf("forbidden field %s in config file: %s", field, forbbidenInfo.Reason)
			} else {
				log.Debugf("field %s in config file doesnt have a default value", field)
			}
		}
	}
	// TODO: Write config file must a parameter of ARGS
	writeConfigFileName := "/tmp/cdk_config.toml"
	log.Infof("Writing used config to %s", writeConfigFileName)
	err = os.WriteFile(writeConfigFileName, []byte(configData), DefaultCreationFilePermissions)
	if err != nil {
		log.Warnf("Error writing config file: %s", err)
	}
	return nil
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
