package zerolog

import (
	"strings"

	"github.com/rs/zerolog"
)

// ZerologAdapter is a wrapper around zerolog.Logger to implement the goose.Logger interface
type ZerologGooseAdapter struct {
	Logger zerolog.Logger
}

// Print implement the goose.Logger interface method
func (z *ZerologGooseAdapter) Print(args ...any) {
	z.Logger.Info().Msgf("%v", args...)
}

// Printf implement the goose.Logger interface method
func (z *ZerologGooseAdapter) Printf(format string, args ...any) {
	f := strings.Replace(format, "\n", "", -1)
	z.Logger.Info().Msgf(f, args...)
}

// Println implement the goose.Logger interface method
func (z *ZerologGooseAdapter) Println(args ...any) {
	z.Logger.Info().Msgf("%v", args...)
}

// Fatal implement the goose.Logger interface method
func (z *ZerologGooseAdapter) Fatal(args ...any) {
	z.Logger.Fatal().Msgf("%v", args...)
}

// Fatalf implement the goose.Logger interface method
func (z *ZerologGooseAdapter) Fatalf(format string, args ...any) {
	f := strings.Replace(format, "\n", "", -1)
	z.Logger.Fatal().Msgf(f, args...)
}
