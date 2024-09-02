package main

import (
	"os"

	zkevm "github.com/0xPolygon/cdk"
	"github.com/urfave/cli/v2"
)

func versionCmd(*cli.Context) error {
	zkevm.PrintVersion(os.Stdout)

	return nil
}
