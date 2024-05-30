package cmd

import (
	"os"

	"github.com/0xPolygon/cdk/version"
	"github.com/urfave/cli/v2"
)

func VersionCmd(*cli.Context) error {
	version.PrintVersion(os.Stdout)
	return nil
}
