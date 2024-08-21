package l1bridge2infoindexsync

import "github.com/0xPolygon/cdk/config/types"

type Config struct {
	// DBPath path of the DB
	DBPath string `mapstructure:"DBPath"`
	// RetryAfterErrorPeriod is the time that will be waited when an unexpected error happens before retry
	RetryAfterErrorPeriod types.Duration `mapstructure:"RetryAfterErrorPeriod"`
	// MaxRetryAttemptsAfterError is the maximum number of consecutive attempts that will happen before panicing.
	// Any number smaller than zero will be considered as unlimited retries
	MaxRetryAttemptsAfterError int `mapstructure:"MaxRetryAttemptsAfterError"`
	// WaitForSyncersPeriod time that will be waited when the synchronizer has reached the latest state
	WaitForSyncersPeriod types.Duration `mapstructure:"WaitForSyncersPeriod"`
}
