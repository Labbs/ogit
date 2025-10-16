package jobs

import "github.com/go-co-op/gocron/v2"

func (c *Config) CleanUsersSessions() error {
	logger := c.Logger.With().Str("component", "infrastructure.jobs.clean_users_sessions").Logger()

	_, err := c.CronScheduler.CronScheduler.NewJob(
		gocron.CronJob("*/1 * * * * ", false), // Every 1 minute
		gocron.NewTask(func() { _ = c.SessionApp.DeleteExpired() }),
		gocron.WithName("CleanUsersSessions"),
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to schedule CleanUsersSessions job")
	}

	return err
}
