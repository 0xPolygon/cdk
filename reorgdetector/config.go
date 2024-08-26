package reorgdetector

import "time"

const (
	defaultCheckReorgsInterval = 2 * time.Second
)

// Config is the configuration for the reorg detector
type Config struct {
	// DBPath is the path to the database
	DBPath string `mapstructure:"DBPath"`

	// CheckReorgsInterval is the interval to check for reorgs in tracked blocks
	CheckReorgsInterval time.Duration `mapstructure:"CheckReorgsInterval"`
}

// GetCheckReorgsInterval returns the interval to check for reorgs in tracked blocks
func (c *Config) GetCheckReorgsInterval() time.Duration {
	if c.CheckReorgsInterval == 0 {
		return defaultCheckReorgsInterval
	}

	return c.CheckReorgsInterval
}
