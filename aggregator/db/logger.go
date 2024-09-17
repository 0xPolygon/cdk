package db

import (
	"context"
	"fmt"

	"github.com/0xPolygon/cdk/log"
	"github.com/jackc/pgx/v4"
)

type dbLoggerImpl struct{}

func (l dbLoggerImpl) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	m := fmt.Sprintf("%s %v", msg, data)

	switch level {
	case pgx.LogLevelInfo:
		log.Info(m)
	case pgx.LogLevelWarn:
		log.Warn(m)
	case pgx.LogLevelError:
		log.Error(m)
	default:
		m = fmt.Sprintf("[%s] %s %v", level.String(), msg, data)
		log.Debug(m)
	}
}
