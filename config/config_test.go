package config

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestLExploratorySetConfigFlag(t *testing.T) {
	value := []string{"config.json", "another_config.json"}
	ctx := newCliContextConfigFlag(t, value...)
	configFilePath := ctx.StringSlice(FlagCfg)
	require.Equal(t, value, configFilePath)
}

func TestLoadDefaultConfig(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ut_config")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.Write([]byte(DefaultMandatoryVars))
	require.NoError(t, err)
	ctx := newCliContextConfigFlag(t, tmpFile.Name())
	cfg, err := Load(ctx)
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

const configWithDeprecatedFields = `
[SequenceSender.EthTxManager]
nodepretatedfield = "value2"
persistencefilename = "value"
`

func TestLoadConfigWithDeprecatedFields(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ut_config")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.Write([]byte(DefaultVars + "\n" + configWithDeprecatedFields))
	require.NoError(t, err)
	ctx := newCliContextConfigFlag(t, tmpFile.Name())
	cfg, err := Load(ctx)
	require.Error(t, err)
	require.Nil(t, cfg)
}

func TestLoadConfigWithSaveConfigFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ut_config")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.Write([]byte(DefaultVars + "\n"))
	require.NoError(t, err)
	fmt.Printf("file: %s\n", tmpFile.Name())
	ctx := newCliContextConfigFlag(t, tmpFile.Name())
	dir, err := os.MkdirTemp("", "ut_test_save_config")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	err = ctx.Set(FlagSaveConfigPath, dir)
	require.NoError(t, err)
	cfg, err := Load(ctx)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	_, err = os.Stat(dir + "/" + SaveConfigFileName)
	require.NoError(t, err)
}

func TestTLoadFileFromStringDeprecatedField(t *testing.T) {
	configFileData := configWithDeprecatedFields
	_, err := LoadFileFromString(configFileData, "toml")
	require.Error(t, err)
}
func TestTLoadDeprecatedField(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ut_config")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.Write([]byte(DefaultVars + "\n" + configWithDeprecatedFields))
	require.NoError(t, err)
	ctx := newCliContextConfigFlag(t, tmpFile.Name())
	_, err = Load(ctx)
	require.Error(t, err)
}

func TestTLoadDeprecatedFieldWithAllowFlag(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ut_config")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.Write([]byte(DefaultVars + "\n" + configWithDeprecatedFields))
	require.NoError(t, err)
	ctx := newCliContextConfigFlag(t, tmpFile.Name())
	err = ctx.Set(FlagAllowDeprecatedFields, "true")
	require.NoError(t, err)
	_, err = Load(ctx)
	require.NoError(t, err)
}

func TestCheckDeprecatedFields(t *testing.T) {
	err := checkDeprecatedFields([]string{deprecatedFieldsOnConfig[0].FieldNamePattern})
	require.Error(t, err)
	require.Contains(t, err.Error(), deprecatedFieldsOnConfig[0].FieldNamePattern)
	require.Contains(t, err.Error(), deprecatedFieldsOnConfig[0].Reason)
}

func TestCheckDeprecatedFieldsPattern(t *testing.T) {
	err := checkDeprecatedFields([]string{"aggregator.synchronizer.db.name"})
	require.Error(t, err)
	require.Contains(t, err.Error(), deprecatedFieldSyncDB)
}

func TestLoadConfigWithInvalidFilename(t *testing.T) {
	ctx := newCliContextConfigFlag(t, "invalid_file")
	cfg, err := Load(ctx)
	require.Error(t, err)
	require.Nil(t, cfg)
}

func TestLoadConfigWithForbiddenFields(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{
			name: "[Aggregator.Synchronizer] DB",
			input: `[Aggregator.Synchronizer.DB]
						name = "value"`,
		},
		{
			name: "[SequenceSender.EthTxManager] PersistenceFilename",
			input: `[SequenceSender.EthTxManager]
						PersistenceFilename = "foo.json"`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "ut_config")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())
			_, err = tmpFile.Write([]byte(c.input))
			require.NoError(t, err)
			ctx := newCliContextConfigFlag(t, tmpFile.Name())
			cfg, err := Load(ctx)
			require.Error(t, err)
			require.Nil(t, cfg)
		})
	}
}

func newCliContextConfigFlag(t *testing.T, values ...string) *cli.Context {
	t.Helper()
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	var configFilePaths cli.StringSlice
	flagSet.Var(&configFilePaths, FlagCfg, "")
	flagSet.Bool(FlagAllowDeprecatedFields, false, "")
	flagSet.String(FlagSaveConfigPath, "", "")
	for _, value := range values {
		err := flagSet.Parse([]string{"--" + FlagCfg, value})
		require.NoError(t, err)
	}
	return cli.NewContext(nil, flagSet, nil)
}
