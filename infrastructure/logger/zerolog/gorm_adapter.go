package zerolog

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type GormLogger struct {
	logger        zerolog.Logger
	LogLevel      gormlogger.LogLevel
	SlowThreshold time.Duration
}

func NewGormLogger(logger zerolog.Logger) *GormLogger {
	return &GormLogger{
		logger:        logger.With().Str("component", "gorm").Logger(),
		LogLevel:      gormlogger.Info,
		SlowThreshold: 200 * time.Millisecond,
	}
}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormlogger.Info {
		l.logger.Info().Msgf(msg, data...)
	}
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormlogger.Warn {
		l.logger.Warn().Msgf(msg, data...)
	}
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormlogger.Error {
		l.logger.Error().Msgf(msg, data...)
	}
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	logEvent := l.logger.With().
		Str("type", "sql").
		Float64("elapsed_ms", float64(elapsed.Nanoseconds())/1e6).
		Str("sql", sql)

	if rows >= 0 {
		logEvent = logEvent.Int64("rows", rows)
	}

	logger := logEvent.Logger()

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error().Err(err).Send()
		return
	}

	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		logger.Warn().Msg("SLOW SQL")
		return
	}

	logger.Debug().Send()
}
