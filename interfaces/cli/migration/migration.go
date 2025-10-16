package migration

import (
	"context"

	"github.com/labbs/ogit/infrastructure"
	"github.com/labbs/ogit/infrastructure/config"
	"github.com/labbs/ogit/infrastructure/database"
	"github.com/labbs/ogit/infrastructure/logger"
	"github.com/labbs/ogit/infrastructure/migration"

	"github.com/urfave/cli/v3"
)

// NewInstance creates a new CLI command for running database migrations.
// It's called by the main application to add the "migration" command to the CLI.
func NewInstance(version string) *cli.Command {
	cfg := &config.Config{}
	cfg.Version = version
	migrationFlags := getFlags(cfg)

	return &cli.Command{
		Name:  "migration",
		Usage: "Start the migration tool",
		Flags: migrationFlags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runMigration(*cfg)
		},
	}
}

// getFlags returns the list of CLI flags required for the migration command.
func getFlags(cfg *config.Config) (list []cli.Flag) {
	list = append(list, config.GenericFlags(cfg)...)
	list = append(list, config.LoggerFlags(cfg)...)
	list = append(list, config.DatabaseFlags(cfg)...)
	return
}

// runMigration initializes the necessary dependencies and runs the database migrations.
func runMigration(cfg config.Config) error {
	var err error

	// Initialize dependencies
	deps := infrastructure.Deps{
		Config: cfg,
	}

	// Initialize logger
	deps.Logger = logger.NewLogger(cfg.Logger.Level, cfg.Logger.Pretty, cfg.Version)
	logger := deps.Logger.With().Str("component", "migration.runserver").Logger()

	// Initialize database connection
	deps.Database, err = database.Configure(deps.Config, deps.Logger)
	if err != nil {
		logger.Fatal().Err(err).Str("event", "migration.runserver.database.configure").Msg("Failed to configure database connection")
		return err
	}

	if err := migration.RunMigration(deps.Logger, deps.Database.Db); err != nil {
		logger.Error().Err(err).Str("event", "migration.runmigration").Msg("Failed to run migrations")
		return err
	}

	return nil
}
