package common

import (
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
	"github.com/gofiber/fiber/v2"
	"github.com/labbs/git-server-s3/pkg/storage"
)

func GetTransportServer(repoPath string, str storage.GitRepositoryStorage) (transport.Transport, *transport.Endpoint, error) {
	normalizedPath := NormalizeRepoPath(repoPath)

	if !str.RepositoryExists(normalizedPath) {
		return nil, nil, fiber.NewError(fiber.StatusNotFound, "repository not found")
	}

	// Create a loader for this specific repository
	loader := storage.NewGitServerLoader(str, normalizedPath)

	// Create the transport server
	srv := server.NewServer(loader)
	ep := &transport.Endpoint{Path: "/" + filepath.Base(normalizedPath)}

	return srv, ep, nil
}
