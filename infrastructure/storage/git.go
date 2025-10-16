package storage

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/labbs/ogit/infrastructure/config"
	"github.com/labbs/ogit/infrastructure/storage/local"
	"github.com/labbs/ogit/infrastructure/storage/s3"
	"github.com/rs/zerolog"
)

// GitRepositoryStorage defines the interface for Git repository storage backends
type Config interface {
	// GetStorer returns a go-git storer for the given repository path
	GetStorer(repoPath string) (storer.Storer, error)

	// CreateRepository creates a new bare repository at the specified path
	CreateRepository(repoPath string) error

	// RepositoryExists checks if a repository exists at the given path
	RepositoryExists(repoPath string) bool

	// DeleteRepository removes a repository at the given path
	DeleteRepository(repoPath string) error

	// ListRepositories returns a list of all repository paths
	ListRepositories() ([]string, error)

	// Configure initializes the storage backend
	Configure() error
}

// GitServerLoader implements go-git's server.Loader interface
// using our storage abstraction
type GitServerLoader struct {
	storage  Config
	repoPath string
}

// NewGitServerLoader creates a new loader for a specific repository
func NewGitServerLoader(storage Config, repoPath string) *GitServerLoader {
	return &GitServerLoader{
		storage:  storage,
		repoPath: repoPath,
	}
}

// Load implements server.Loader interface
func (l *GitServerLoader) Load(ep *transport.Endpoint) (storer.Storer, error) {
	return l.storage.GetStorer(l.repoPath)
}

// Config creates a new Config instance based on configuration
func Configure(_cfg config.Config, logger zerolog.Logger) (Config, error) {
	logger = logger.With().Str("component", "infrastructure.storage").Logger()
	logger.Info().Str("storage_type", _cfg.Storage.Type).Msg("configuring storage backend")
	var storage Config
	switch _cfg.Storage.Type {
	case "local":
		storage = local.NewLocalStorage(logger, _cfg)
	case "s3":
		storage = s3.NewS3Storage(logger, _cfg)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", _cfg.Storage.Type)
	}

	err := storage.Configure()
	if err != nil {
		logger.Error().Err(err).Str("storage_type", _cfg.Storage.Type).Msg("failed to configure storage")
		return nil, err
	}

	return storage, nil
}
