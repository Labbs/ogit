package zerolog

import z "github.com/rs/zerolog"

type GocronAdapter struct {
	Logger z.Logger
}

func (l GocronAdapter) Println(msg string, v ...any) {
	l.Logger.Info().Msgf(msg, v...)
}

func (l GocronAdapter) Debug(msg string, v ...any) {
	l.Logger.Debug().Msgf(msg, v...)
}

func (l GocronAdapter) Info(msg string, v ...any) {
	l.Logger.Info().Msgf(msg, v...)
}

func (l GocronAdapter) Warn(msg string, v ...any) {
	l.Logger.Warn().Msgf(msg, v...)
}

func (l GocronAdapter) Error(msg string, v ...any) {
	l.Logger.Error().Msgf(msg, v...)
}
