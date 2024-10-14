package config

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestLExploratorySetConfigFalg(t *testing.T) {
	value := []string{"config.json", "another_config.json"}
	ctx := newCliContextConfigFlag(t, value...)
	configFilePath := ctx.StringSlice(FlagCfg)
	require.Equal(t, value, configFilePath)
}

func TestLoadDeafaultConfig(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ut_config")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.Write([]byte(DefaultVars))
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
	fmt.Printf("file: %s\n", tmpFile.Name())
	ctx := newCliContextConfigFlag(t, tmpFile.Name())
	cfg, err := Load(ctx)
	require.Error(t, err)
	require.Nil(t, cfg)
}

func TestTLoadFileFromStringDeprecatedField(t *testing.T) {
	configFileData := configWithDeprecatedFields
	cfg, err := LoadFileFromString(configFileData, "toml")
	require.Error(t, err)
	require.Nil(t, cfg)
}

func TestCheckDeprecatedFields(t *testing.T) {
	err := checkDeprecatedFields([]string{deprecatedFieldsOnConfig[0].FieldNamePattern})
	require.Error(t, err)
	require.Contains(t, err.Error(), deprecatedFieldsOnConfig[0].FieldNamePattern)
	require.Contains(t, err.Error(), deprecatedFieldsOnConfig[0].Reason)
}

func TestLoadConfigWithInvalidFilename(t *testing.T) {
	flagSet := flag.FlagSet{}
	flagSet.String(FlagCfg, "invalid_file", "")
	ctx := cli.NewContext(nil, &flagSet, nil)
	cfg, err := Load(ctx)
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestLoadConfigWithForbiddenFields(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{
			name: "[Aggregator.Synchronizer] DB",
			input: `[aggregator.synchronizer.db]
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
			flagSet := flag.FlagSet{}
			flagSet.String(FlagCfg, tmpFile.Name(), "")
			ctx := cli.NewContext(nil, &flagSet, nil)
			cfg, err := Load(ctx)
			require.NoError(t, err)
			require.NotNil(t, cfg)
		})
	}
}

func newCliContextConfigFlag(t *testing.T, values ...string) *cli.Context {
	t.Helper()
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	var configFilePaths cli.StringSlice
	flagSet.Var(&configFilePaths, FlagCfg, "")
	for _, value := range values {
		err := flagSet.Parse([]string{"--" + FlagCfg, value})
		require.NoError(t, err)
	}
	return cli.NewContext(nil, flagSet, nil)
}
