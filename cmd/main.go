package main

import (
	"os"

	zkevm "github.com/0xPolygon/cdk"
	"github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/config"
	"github.com/0xPolygon/cdk/log"
	"github.com/urfave/cli/v2"
)

const appName = "cdk"

const (
	// NETWORK_CONFIGFILE name to identify the network_custom (genesis) config-file
	NETWORK_CONFIGFILE = "custom_network" //nolint:stylecheck
)

var (
	configFileFlag = cli.StringSliceFlag{
		Name:     config.FlagCfg,
		Aliases:  []string{"c"},
		Usage:    "Configuration file(s)",
		Required: true,
	}
	customNetworkFlag = cli.StringFlag{
		Name:     config.FlagCustomNetwork,
		Aliases:  []string{"net-file"},
		Usage:    "Load the network configuration file if --network=custom",
		Required: false,
	}
	yesFlag = cli.BoolFlag{
		Name:     config.FlagYes,
		Aliases:  []string{"y"},
		Usage:    "Automatically accepts any confirmation to execute the command",
		Required: false,
	}
	componentsFlag = cli.StringSliceFlag{
		Name:     config.FlagComponents,
		Aliases:  []string{"co"},
		Usage:    "List of components to run",
		Required: false,
		Value: cli.NewStringSlice(common.SEQUENCE_SENDER, common.AGGREGATOR,
			common.AGGORACLE, common.RPC, common.AGGSENDER),
	}
	saveConfigFlag = cli.StringFlag{
		Name:     config.FlagSaveConfigPath,
		Aliases:  []string{"s"},
		Usage:    "Save final configuration into to the indicated path (name: cdk-node-config.toml)",
		Required: false,
	}
	disableDefaultConfigVars = cli.BoolFlag{
		Name:     config.FlagDisableDefaultConfigVars,
		Aliases:  []string{"d"},
		Usage:    "Disable default configuration variables, all of them must be defined on config files",
		Required: false,
	}

	allowDeprecatedFields = cli.BoolFlag{
		Name:     config.FlagAllowDeprecatedFields,
		Usage:    "Allow that config-files contains deprecated fields",
		Required: false,
	}
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Version = zkevm.Version
	flags := []cli.Flag{
		&configFileFlag,
		&yesFlag,
		&componentsFlag,
		&saveConfigFlag,
		&disableDefaultConfigVars,
		&allowDeprecatedFields,
	}
	app.Commands = []*cli.Command{
		{
			Name:    "version",
			Aliases: []string{},
			Usage:   "Application version and build",
			Action:  versionCmd,
		},
		{
			Name:    "run",
			Aliases: []string{},
			Usage:   "Run the cdk client",
			Action:  start,
			Flags:   append(flags, &customNetworkFlag),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
