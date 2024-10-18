package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	jRPC "github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk/aggoracle"
	"github.com/0xPolygon/cdk/aggregator"
	"github.com/0xPolygon/cdk/aggsender"
	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/claimsponsor"
	"github.com/0xPolygon/cdk/common"
	ethermanconfig "github.com/0xPolygon/cdk/etherman/config"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/lastgersync"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/0xPolygon/cdk/sequencesender"
	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml/v2"
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
	// FlagDisableDefaultConfigVars is the flag to force all variables to be set on config-files
	FlagDisableDefaultConfigVars = "disable-default-config-vars"
	// FlagAllowDeprecatedFields is the flag to allow deprecated fields
	FlagAllowDeprecatedFields = "allow-deprecated-fields"

	deprecatedFieldSyncDB = "Aggregator.Synchronizer.DB is deprecated. Use Aggregator.Synchronizer.SQLDB instead."

	deprecatedFieldPersistenceFilename = "EthTxManager.PersistenceFilename is deprecated." +
		" Use EthTxManager.StoragePath instead."

	EnvVarPrefix       = "CDK"
	ConfigType         = "toml"
	SaveConfigFileName = "cdk_config.toml"

	DefaultCreationFilePermissions = os.FileMode(0600)
)

type DeprecatedFieldsError struct {
	// key is the rule and the value is the field's name that matches the rule
	Fields map[DeprecatedField][]string
}

func NewErrDeprecatedFields() *DeprecatedFieldsError {
	return &DeprecatedFieldsError{
		Fields: make(map[DeprecatedField][]string),
	}
}

func (e *DeprecatedFieldsError) AddDeprecatedField(fieldName string, rule DeprecatedField) {
	p := e.Fields[rule]
	e.Fields[rule] = append(p, fieldName)
}

func (e *DeprecatedFieldsError) Error() string {
	res := "found deprecated fields:"
	for rule, fieldsMatches := range e.Fields {
		res += fmt.Sprintf("\n\t- %s: %s", rule.Reason, strings.Join(fieldsMatches, ", "))
	}
	return res
}

type DeprecatedField struct {
	// If the field name ends with a dot means that match a section
	FieldNamePattern string
	Reason           string
}

var (
	deprecatedFieldsOnConfig = []DeprecatedField{
		{
			FieldNamePattern: "sequencesender.ethtxmanager.persistencefilename",
			Reason:           deprecatedFieldPersistenceFilename,
		},
		{
			FieldNamePattern: "aggregator.synchronizer.db.",
			Reason:           deprecatedFieldSyncDB,
		},

		{
			FieldNamePattern: "aggregator.ethtxmanager.persistencefilename",
			Reason:           deprecatedFieldPersistenceFilename,
		},
	}
)

/*
Config represents the configuration of the entire CDK Node
The file is [TOML format]

[TOML format]: https://en.wikipedia.org/wiki/TOML
*/
type Config struct {
	// Configuration of the etherman (client for access L1)
	Etherman ethermanconfig.Config
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

	// AggSender is the configuration of the agg sender service
	AggSender aggsender.Config
}

// Load loads the configuration
func Load(ctx *cli.Context) (*Config, error) {
	configFilePath := ctx.StringSlice(FlagCfg)
	filesData, err := readFiles(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading files:  Err:%w", err)
	}
	saveConfigPath := ctx.String(FlagSaveConfigPath)
	defaultConfigVars := !ctx.Bool(FlagDisableDefaultConfigVars)
	allowDeprecatedFields := ctx.Bool(FlagAllowDeprecatedFields)
	return LoadFile(filesData, saveConfigPath, defaultConfigVars, allowDeprecatedFields)
}

func readFiles(files []string) ([]FileData, error) {
	result := make([]FileData, 0, len(files))
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
	err := loadString(cfg, configFileData, configType, true, EnvVarPrefix)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

func SaveConfigToFile(cfg *Config, saveConfigPath string) error {
	marshaled, err := toml.Marshal(cfg)
	if err != nil {
		log.Errorf("Can't marshal config to toml. Err: %w", err)
		return err
	}
	return SaveDataToFile(saveConfigPath, "final config file", marshaled)
}

func SaveDataToFile(fullPath, reason string, data []byte) error {
	log.Infof("Writing %s to: %s", reason, fullPath)
	err := os.WriteFile(fullPath, data, DefaultCreationFilePermissions)
	if err != nil {
		err = fmt.Errorf("error writing %s to file %s. Err: %w", reason, fullPath, err)
		log.Error(err)
		return err
	}
	return nil
}

// Load loads the configuration
func LoadFile(files []FileData, saveConfigPath string,
	setDefaultVars bool, allowDeprecatedFields bool) (*Config, error) {
	log.Infof("Loading configuration: saveConfigPath: %s, setDefaultVars: %t, allowDeprecatedFields: %t",
		saveConfigPath, setDefaultVars, allowDeprecatedFields)
	fileData := make([]FileData, 0)
	if setDefaultVars {
		log.Info("Setting default vars")
		fileData = append(fileData, FileData{Name: "default_mandatory_vars", Content: DefaultMandatoryVars})
	}
	fileData = append(fileData, FileData{Name: "default_vars", Content: DefaultVars})
	fileData = append(fileData, FileData{Name: "default_values", Content: DefaultValues})
	fileData = append(fileData, files...)

	merger := NewConfigRender(fileData, EnvVarPrefix)

	renderedCfg, err := merger.Render()
	if err != nil {
		return nil, err
	}
	if saveConfigPath != "" {
		fullPath := saveConfigPath + "/" + SaveConfigFileName + ".merged"
		err = SaveDataToFile(fullPath, "merged config file", []byte(renderedCfg))
		if err != nil {
			return nil, err
		}
	}
	cfg, err := LoadFileFromString(renderedCfg, ConfigType)
	// If allowDeprecatedFields is true, we ignore the deprecated fields
	if err != nil && allowDeprecatedFields {
		var customErr *DeprecatedFieldsError
		if errors.As(err, &customErr) {
			log.Warnf("detected deprecated fields: %s", err.Error())
			err = nil
		}
	}

	if err != nil {
		return nil, err
	}
	if saveConfigPath != "" {
		fullPath := saveConfigPath + "/" + SaveConfigFileName
		err = SaveConfigToFile(cfg, fullPath)
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// Load loads the configuration
func loadString(cfg *Config, configData string, configType string,
	allowEnvVars bool, envPrefix string) error {
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
	configKeys := viper.AllKeys()
	err = checkDeprecatedFields(configKeys)
	if err != nil {
		return err
	}

	return nil
}

func checkDeprecatedFields(keysOnConfig []string) error {
	err := NewErrDeprecatedFields()
	for _, key := range keysOnConfig {
		forbbidenInfo := getDeprecatedField(key)
		if forbbidenInfo != nil {
			err.AddDeprecatedField(key, *forbbidenInfo)
		}
	}
	if len(err.Fields) > 0 {
		return err
	}
	return nil
}

func getDeprecatedField(fieldName string) *DeprecatedField {
	for _, deprecatedField := range deprecatedFieldsOnConfig {
		if deprecatedField.FieldNamePattern == fieldName {
			return &deprecatedField
		}
		// If the field name ends with a dot, it means FieldNamePattern*
		if deprecatedField.FieldNamePattern[len(deprecatedField.FieldNamePattern)-1] == '.' &&
			strings.HasPrefix(fieldName, deprecatedField.FieldNamePattern) {
			return &deprecatedField
		}
	}
	return nil
}
