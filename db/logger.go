package db

import (
	"context"

	"github.com/jackc/pgx"
)

type logger struct{}

func (l logger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	// TODO: adapt to sqlite

	// m := fmt.Sprintf("%s %v", msg, data)

	// switch level {
	// case pgx.LogLevelInfo:
	// 	log.Info(m)
	// case pgx.LogLevelWarn:
	// 	log.Warn(m)
	// case pgx.LogLevelError:
	// 	log.Error(m)
	// default:
	// 	m = fmt.Sprintf("%s %s %v", level.String(), msg, data)
	// 	log.Debug(m)
	// }
}
