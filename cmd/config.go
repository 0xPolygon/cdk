package main

import (
	"os"
	"strings"

	"github.com/0xPolygon/cdk/config"
	"github.com/urfave/cli/v2"
)

func configCmd(*cli.Context) error {
	// var files []config.FileData
	// fileData := make([]config.FileData, 0)
	// fileData = append(fileData, config.FileData{Name: "default_mandatory_vars", Content: config.DefaultMandatoryVars})
	// fileData = append(fileData, config.FileData{Name: "default_vars", Content: config.DefaultVars})
	// fileData = append(fileData, config.FileData{Name: "default_values", Content: config.DefaultValues})
	// fileData = append(fileData, files...)
	// String buffer to concatenate all the default config vars
	defaultConfig := strings.Builder{}
	defaultConfig.WriteString(config.DefaultMandatoryVars)
	defaultConfig.WriteString(config.DefaultVars)
	defaultConfig.WriteString(config.DefaultValues)

	// merger := config.NewConfigRender(fileData, config.EnvVarPrefix)

	// renderedCfg, err := merger.Render()
	// if err != nil {
	// 	return err
	// }

	_, err := os.Stdout.WriteString(defaultConfig.String())
	if err != nil {
		return err
	}

	return nil
}
