package cmd

import (
	"context"

	"github.com/labbs/git-server-s3/internal/config"
	flags "github.com/labbs/git-server-s3/internal/flags"
	"github.com/labbs/git-server-s3/internal/server"
	"github.com/labbs/git-server-s3/pkg/logger"

	"github.com/urfave/cli/v3"
)

// NewInstance creates a new 'server' command instance for urfave cli
func NewInstance(version string) *cli.Command {
	config.Version = version
	serverFlags := getFlags()

	return &cli.Command{
		Name:   "server",
		Usage:  "Start the stack-deployer application",
		Flags:  serverFlags,
		Action: runServer,
	}
}

// getFlags returns the list of flags for the server command.
func getFlags() (list []cli.Flag) {
	list = append(list, flags.GenericFlags()...)
	list = append(list, flags.ServerFlags()...)
	list = append(list, flags.LoggerFlags()...)
	return
}

// runServer starts the server following the configuration.
func runServer(ctx context.Context, c *cli.Command) error {
	l := logger.NewLogger(config.Logger.Level, config.Logger.Pretty, c.Root().Version)

	var httpConfig server.HttpConfig
	httpConfig.Port = config.Server.Port
	httpConfig.HttpLogs = config.Server.HttpLogs
	httpConfig.Logger = l

	httpConfig.NewServer()

	return nil
}
