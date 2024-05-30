package main

import (
	"log"
	"os"

	"github.com/0xPolygon/cdk/cmd"
	"github.com/0xPolygon/cdk/config"
	"github.com/0xPolygon/cdk/version"
	"github.com/urfave/cli/v2"
)

const appName = "cdk"

var (
	configFileFlag = cli.StringFlag{
		Name:     config.FlagCfg,
		Aliases:  []string{"c"},
		Usage:    "Configuration `FILE`",
		Required: true,
	}
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Version = version.Version

	app.Commands = []*cli.Command{
		{
			Name:    "version",
			Aliases: []string{},
			Usage:   "Application version and build",
			Action:  cmd.VersionCmd,
		},
		{
			Name:    "run",
			Aliases: []string{},
			Usage:   "Run cdk client",
			Action:  cmd.RunCmd,
			Flags:   []cli.Flag{&configFileFlag},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
