package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func NewLogger(level string, pretty bool, version string) zerolog.Logger {
	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}

	logger := zerolog.New(os.Stderr).With().
		Caller().
		Timestamp().
		Str("host", host).
		Str("version", version).
		Logger()

	if pretty {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	switch level {
	case "debug":
		logger = logger.Level(zerolog.DebugLevel)
	case "warn":
		logger = logger.Level(zerolog.WarnLevel)
	case "error":
		logger = logger.Level(zerolog.ErrorLevel)
	case "fatal":
		logger = logger.Level(zerolog.FatalLevel)
	case "panic":
		logger = logger.Level(zerolog.PanicLevel)
	default:
		logger = logger.Level(zerolog.InfoLevel)
	}

	return logger
}
