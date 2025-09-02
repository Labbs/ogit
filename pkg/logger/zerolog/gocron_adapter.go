package zerolog

import z "github.com/rs/zerolog"

type GocronAdapter struct {
	Logger z.Logger
}

func (g GocronAdapter) Println(msg string, v ...any) {
	g.Logger.Info().Msgf(msg, v...)
}

func (g GocronAdapter) Debug(msg string, v ...any) {
	g.Logger.Debug().Msgf(msg, v...)
}

func (g GocronAdapter) Error(msg string, v ...any) {
	g.Logger.Error().Msgf(msg, v...)
}

func (g GocronAdapter) Info(msg string, v ...any) {
	g.Logger.Info().Msgf(msg, v...)
}

func (g GocronAdapter) Warn(msg string, v ...any) {
	g.Logger.Warn().Msgf(msg, v...)
}
