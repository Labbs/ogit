package server

import (
	"context"
	"strconv"

	"github.com/labbs/ogit/application"
	"github.com/labbs/ogit/infrastructure"
	"github.com/labbs/ogit/infrastructure/config"
	"github.com/labbs/ogit/infrastructure/cronscheduler"
	"github.com/labbs/ogit/infrastructure/database"
	"github.com/labbs/ogit/infrastructure/http"
	"github.com/labbs/ogit/infrastructure/jobs"
	"github.com/labbs/ogit/infrastructure/logger"
	"github.com/labbs/ogit/infrastructure/persistence"
	"github.com/labbs/ogit/infrastructure/storage"
	routes "github.com/labbs/ogit/interfaces/http"

	"github.com/urfave/cli/v3"
)

// NewInstance creates a new CLI command for starting the server.
// It's called by the main application to add the "server" command to the CLI.
func NewInstance(version string) *cli.Command {
	cfg := &config.Config{}
	cfg.Version = version
	serverFlags := getFlags(cfg)

	return &cli.Command{
		Name:  "server",
		Usage: "Start the HTTP server",
		Flags: serverFlags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runServer(*cfg)
		},
	}
}

// getFlags returns the list of CLI flags required for the server command.
func getFlags(cfg *config.Config) (list []cli.Flag) {
	list = append(list, config.GenericFlags(cfg)...)
	list = append(list, config.ServerFlags(cfg)...)
	list = append(list, config.LoggerFlags(cfg)...)
	list = append(list, config.DatabaseFlags(cfg)...)
	list = append(list, config.SessionFlags(cfg)...)
	list = append(list, config.RegistrationFlags(cfg)...)
	list = append(list, config.StorageFlags(cfg)...)
	return
}

// runServer initializes the necessary dependencies and starts the HTTP server.
func runServer(cfg config.Config) error {
	var err error

	// Initialize dependencies
	deps := infrastructure.Deps{
		Config: cfg,
	}

	// Initialize logger
	deps.Logger = logger.NewLogger(cfg.Logger.Level, cfg.Logger.Pretty, cfg.Version)
	logger := deps.Logger.With().Str("component", "interfaces.cli.http.runserver").Logger()

	// Initialize other cron scheduler (go-cron)
	deps.CronScheduler, err = cronscheduler.Configure(deps.Logger)
	if err != nil {
		logger.Fatal().Err(err).Str("event", "http.runserver.cronscheduler.configure").Msg("Failed to configure cron scheduler")
		return err
	}

	// Initialize database connection (gorm)
	deps.Database, err = database.Configure(deps.Config, deps.Logger)
	if err != nil {
		logger.Fatal().Err(err).Str("event", "http.runserver.database.configure").Msg("Failed to configure database connection")
		return err
	}

	// Initialize storage
	deps.Storage, err = storage.Configure(deps.Config, deps.Logger)
	if err != nil {
		logger.Fatal().Err(err).Str("event", "http.runserver.storage.configure").Msg("Failed to configure storage")
		return err
	}

	// Initialize application services
	userPers := persistence.NewUserPers(deps.Database.Db)
	groupPers := persistence.NewGroupPers(deps.Database.Db)
	sessionPers := persistence.NewSessionPers(deps.Database.Db)
	repoPers := persistence.NewRepositoryPers(deps.Database.Db)
	tokenPers := persistence.NewTokenPers(deps.Database.Db)
	repositoryMemberPers := persistence.NewRepositoryMemberPers(deps.Database.Db)

	deps.UserApp = application.NewUserApp(deps.Config, deps.Logger, userPers, groupPers)
	deps.SessionApp = application.NewSessionApp(deps.Config, deps.Logger, sessionPers, deps.UserApp)
	deps.RepoApp = application.NewRepositoryApp(deps.Config, deps.Logger, repoPers, repositoryMemberPers, nil, deps.Storage)
	deps.TokenApp = application.NewTokenApp(deps.Config, deps.Logger, tokenPers)
	deps.AuthApp = application.NewAuthApp(deps.Config, deps.Logger, *deps.UserApp, *deps.SessionApp, *deps.TokenApp, *deps.RepoApp)

	// Initialize HTTP server (fiber + fiberoapi)
	deps.Http, err = http.Configure(deps.Config, deps.Logger, *deps.SessionApp, true)
	if err != nil {
		logger.Fatal().Err(err).Str("event", "http.runserver.http.configure").Msg("Failed to configure HTTP server")
		return err
	}

	// Setup cron jobs
	configJobs := jobs.Config{
		Logger:        deps.Logger,
		CronScheduler: deps.CronScheduler,
		SessionApp:    *deps.SessionApp,
	}

	err = configJobs.SetupJobs()
	if err != nil {
		logger.Fatal().Err(err).Str("event", "http.runserver.jobs.setup").Msg("Failed to setup cron jobs")
		return err
	}

	// Setup routes
	routes.SetupRoutes(deps)

	// Start HTTP server
	logger.Info().Str("event", "http.runserver.http.listen").Msgf("Starting HTTP server on port %d", cfg.Server.Port)
	err = deps.Http.Fiber.Listen(":" + strconv.Itoa(cfg.Server.Port))
	if err != nil {
		logger.Fatal().Err(err).Str("event", "http.runserver.http.listen").Msg("Failed to start HTTP server")
		return err
	}

	return nil
}
