package logger

import (
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

func getLogLevel() zerolog.Level {
	var logLevelString, logLevelSet = os.LookupEnv("LOG_LEVEL")
	if logLevel, err := strconv.ParseInt(logLevelString, 10, 8); err != nil || !logLevelSet {
		return zerolog.DebugLevel
	} else {
		return zerolog.Level(int8(logLevel))
	}
}

var baseLogger = zerolog.New(zerolog.NewConsoleWriter()).
	Level(getLogLevel()).
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
