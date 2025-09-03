package storage

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/labbs/git-server-s3/internal/config"
	"github.com/labbs/git-server-s3/pkg/storage/local"
	"github.com/labbs/git-server-s3/pkg/storage/s3"
	"github.com/rs/zerolog"
)

// GitRepositoryStorage defines the interface for Git repository storage backends
type GitRepositoryStorage interface {
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
	storage  GitRepositoryStorage
	repoPath string
}

// NewGitServerLoader creates a new loader for a specific repository
func NewGitServerLoader(storage GitRepositoryStorage, repoPath string) *GitServerLoader {
	return &GitServerLoader{
		storage:  storage,
		repoPath: repoPath,
	}
}

// Load implements server.Loader interface
func (l *GitServerLoader) Load(ep *transport.Endpoint) (storer.Storer, error) {
	return l.storage.GetStorer(l.repoPath)
}

// NewGitRepositoryStorage creates a new GitRepositoryStorage instance based on configuration
func NewGitRepositoryStorage(logger zerolog.Logger) (GitRepositoryStorage, error) {
	switch config.Storage.Type {
	case "local":
		storage := local.NewLocalStorage(logger)
		return storage, nil
	case "s3":
		storage := s3.NewS3Storage(logger)
		return storage, nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Storage.Type)
	}
}
