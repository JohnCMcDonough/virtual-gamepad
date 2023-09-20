package logger

import (
	"github.com/rs/zerolog"
)

var baseLogger = zerolog.New(zerolog.NewConsoleWriter()).
	Level(zerolog.DebugLevel).
	With().
	Timestamp().
	Caller().
	Logger()

func GetBaseLogger() zerolog.Logger {
	return baseLogger
}

func AddContext(logger zerolog.Logger, context map[string]string) zerolog.Logger {
	ctx := logger.With()
	for k, v := range context {
		ctx = ctx.Str(k, v)
	}
	return ctx.Logger()
}

func CreateLogger(context map[string]string) zerolog.Logger {
	return AddContext(baseLogger, context)
}
