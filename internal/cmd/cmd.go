package cmd

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/labbs/git-server-s3/internal/config"
	flags "github.com/labbs/git-server-s3/internal/flags"
	"github.com/labbs/git-server-s3/internal/server"
	"github.com/labbs/git-server-s3/pkg/logger"
	"github.com/labbs/git-server-s3/pkg/storage"

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
	list = append(list, flags.StorageFlags()...)
	return
}

// runServer starts the server following the configuration.
func runServer(ctx context.Context, c *cli.Command) error {
	l := logger.NewLogger(config.Logger.Level, config.Logger.Pretty, c.Root().Version)

	str, err := storage.NewGitRepositoryStorage(l)
	if err != nil {
		return err
	}

	// Configure the storage backend
	if err := str.Configure(); err != nil {
		l.Fatal().Err(err).Msg("Failed to configure storage")
		return err
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// WaitGroup to wait for all servers to shutdown
	var wg sync.WaitGroup

	// Configure HTTP server
	var httpConfig server.HttpConfig
	httpConfig.Port = config.Server.Port
	httpConfig.HttpLogs = config.Server.HttpLogs
	httpConfig.Logger = l
	httpConfig.Storage = str

	// Start HTTP server in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.Info().Int("port", config.Server.Port).Msg("Starting HTTP server")
		if err := httpConfig.NewServer(); err != nil {
			l.Error().Err(err).Msg("HTTP server failed")
		}
	}()

	var sshConfig *server.GitSSHConfig

	// Start SSH server if enabled
	if config.SSH.Enabled {
		sshConfig = &server.GitSSHConfig{
			Port:        config.SSH.Port,
			HostKeyPath: config.SSH.HostKeyPath,
			Logger:      l,
			Storage:     str,
		}

		if err := sshConfig.Configure(); err != nil {
			l.Fatal().Err(err).Msg("Failed to configure SSH server")
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Info().Int("port", config.SSH.Port).Msg("Starting Git SSH server")
			if err := sshConfig.NewServer(); err != nil {
				l.Error().Err(err).Msg("Git SSH server failed")
			}
		}()
	}

	// Wait for interrupt signal
	<-sigChan
	l.Info().Msg("Shutdown signal received, stopping servers...")

	// Shutdown servers gracefully
	go func() {
		if err := httpConfig.Shutdown(); err != nil {
			l.Error().Err(err).Msg("Error shutting down HTTP server")
		}
		if sshConfig != nil {
			if err := sshConfig.Shutdown(); err != nil {
				l.Error().Err(err).Msg("Error shutting down SSH server")
			}
		}
	}()

	// Give servers time to shutdown gracefully
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		l.Info().Msg("All servers stopped gracefully")
	case <-time.After(30 * time.Second):
		l.Warn().Msg("Shutdown timeout reached, forcing exit")
	}

	return nil
}
