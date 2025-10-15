package cronscheduler

import (
	"time"

	"github.com/labbs/ogit/infrastructure/logger/zerolog"

	"github.com/go-co-op/gocron/v2"
	z "github.com/rs/zerolog"
)

type Config struct {
	CronScheduler gocron.Scheduler
}

// Configure sets up the cron scheduler with the provided logger.
// Will return an error if the scheduler cannot be created (fatal)
func Configure(logger z.Logger) (Config, error) {
	logger = logger.With().Str("component", "infrastructure.cronscheduler").Logger()
	var cfg Config
	s, err := gocron.NewScheduler(gocron.WithLogger(zerolog.GocronAdapter{Logger: logger}), gocron.WithLocation(time.UTC))
	if err != nil {
		logger.Fatal().Err(err).Str("event", "cronscheduler.configure.new").Msg("Failed to create cron scheduler")
		return cfg, err
	}
	cfg.CronScheduler = s
	s.Start()
	return cfg, nil
}
