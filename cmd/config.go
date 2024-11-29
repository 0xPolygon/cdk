package main

import (
	"os"
	"strings"

	"github.com/0xPolygon/cdk/config"
	"github.com/urfave/cli/v2"
)

func configCmd(*cli.Context) error {
	// String buffer to concatenate all the default config vars
	defaultConfig := strings.Builder{}
	defaultConfig.WriteString(config.DefaultMandatoryVars)
	defaultConfig.WriteString(config.DefaultVars)
	defaultConfig.WriteString(config.DefaultValues)

	_, err := os.Stdout.WriteString(defaultConfig.String())
	if err != nil {
		return err
	}

	return nil
}
