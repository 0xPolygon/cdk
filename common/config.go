package common

import "github.com/0xPolygon/cdk/translator"

type Config struct {
	// IsValidiumMode has the value true if the sequence sender is running in validium mode.
	IsValidiumMode bool `mapstructure:"IsValidiumMode"`
	// Contract Versions: elderberry, banana
	ContractVersions string            `mapstructure:"ContractVersions"`
	Translator       translator.Config `mapstructure:"Translator"`
}
