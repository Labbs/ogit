package server

import (
	"fmt"

	"github.com/labbs/git-server-s3/pkg/storage"
	"github.com/rs/zerolog"
)

// GitSSHConfig holds configuration for the custom Git SSH server.
// This replaces the old SSHConfig that used gliderlabs/ssh.
type GitSSHConfig struct {
	Port        int                          // SSH server port
	HostKeyPath string                       // Path to SSH host key file
	Logger      zerolog.Logger               // Logger for SSH operations
	Storage     storage.GitRepositoryStorage // Storage backend for repositories
	server      *GitSSHServer                // The underlying Git SSH server instance
}

// Configure sets up the custom Git SSH server.
func (c *GitSSHConfig) Configure() error {
	c.server = &GitSSHServer{
		Port:        c.formatPort(),
		Logger:      c.Logger,
		Storage:     c.Storage,
		HostKeyPath: c.HostKeyPath,
	}

	return c.server.Configure()
}

// NewServer starts the custom Git SSH server.
func (c *GitSSHConfig) NewServer() error {
	if c.server == nil {
		if err := c.Configure(); err != nil {
			return err
		}
	}
	return c.server.Start()
}

// Shutdown gracefully stops the Git SSH server.
func (c *GitSSHConfig) Shutdown() error {
	if c.server != nil {
		return c.server.Stop()
	}
	return nil
}

// formatPort formats the port number as a string with colon prefix.
func (c *GitSSHConfig) formatPort() string {
	if c.Port == 0 {
		c.Port = 2222
	}
	return fmt.Sprintf(":%d", c.Port)
}
