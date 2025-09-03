package local

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	billyos "github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/labbs/git-server-s3/internal/config"
	"github.com/rs/zerolog"
)

type LocalStorage struct {
	Logger   zerolog.Logger
	basePath string
}

func NewLocalStorage(logger zerolog.Logger) *LocalStorage {
	return &LocalStorage{
		Logger: logger,
	}
}

func (ls *LocalStorage) Configure() error {
	ls.Logger.Info().Msg("Configuring local storage")

	if config.Storage.Local.Path == "" {
		return errors.New("local storage path is not configured")
	}

	// Store the base path
	ls.basePath = config.Storage.Local.Path

	// Check if local storage path exists and create if necessary
	info, err := os.Stat(ls.basePath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(ls.basePath, os.ModePerm); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if !info.IsDir() {
		return errors.New("local storage path is not a directory")
	}

	return nil
}

func (ls *LocalStorage) GetStorer(repoPath string) (storer.Storer, error) {
	fullPath := ls.getFullPath(repoPath)

	if !ls.RepositoryExists(repoPath) {
		return nil, errors.New("repository does not exist")
	}

	fs := billyos.New(fullPath)
	return filesystem.NewStorage(fs, cache.NewObjectLRUDefault()), nil
}

func (ls *LocalStorage) CreateRepository(repoPath string) error {
	fullPath := ls.getFullPath(repoPath)

	if ls.RepositoryExists(repoPath) {
		return errors.New("repository already exists")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	// Initialize bare repository
	_, err := git.PlainInit(fullPath, true)
	return err
}

func (ls *LocalStorage) RepositoryExists(repoPath string) bool {
	fullPath := ls.getFullPath(repoPath)
	info, err := os.Stat(fullPath)
	return err == nil && info.IsDir()
}

func (ls *LocalStorage) DeleteRepository(repoPath string) error {
	fullPath := ls.getFullPath(repoPath)

	if !ls.RepositoryExists(repoPath) {
		return errors.New("repository does not exist")
	}

	return os.RemoveAll(fullPath)
}

func (ls *LocalStorage) ListRepositories() ([]string, error) {
	var repos []string

	err := filepath.Walk(ls.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if this is a git repository (ends with .git and is a directory)
		if info.IsDir() && strings.HasSuffix(info.Name(), ".git") {
			// Get relative path from base
			relPath, err := filepath.Rel(ls.basePath, path)
			if err != nil {
				return err
			}
			repos = append(repos, relPath)
		}

		return nil
	})

	return repos, err
}

func (ls *LocalStorage) getFullPath(repoPath string) string {
	// Clean the repo path and ensure it ends with .git
	cleanPath := filepath.Clean(repoPath)
	if !strings.HasSuffix(cleanPath, ".git") {
		cleanPath += ".git"
	}

	// Join with base path
	return filepath.Join(ls.basePath, cleanPath)
}
