// Package controller provides HTTP handlers for repository management operations.
package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/labbs/git-server-s3/pkg/common"
	"github.com/labbs/git-server-s3/pkg/storage"
	"github.com/rs/zerolog"
)

// RepoController handles HTTP requests for repository management operations.
// It provides endpoints for creating and listing Git repositories through
// the configured storage backend.
type RepoController struct {
	Logger  zerolog.Logger               // Logger for request logging and error reporting
	Storage storage.GitRepositoryStorage // Storage backend for repository operations
}

// CreateRepo handles POST requests to create a new Git repository.
// It expects a JSON payload with a "name" field and creates a bare repository
// in the configured storage backend.
//
// Request body: {"name": "repository-name"}
// Response: 201 Created with "repository created" message on success
func (c *RepoController) CreateRepo(ctx *fiber.Ctx) error {
	logger := c.Logger.With().Str("event", "CreateRepo").Logger()

	var req struct {
		Name string `json:"name"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	// Normalize the repository name to ensure proper .git suffix and path format
	normName := common.NormalizeRepoPath(req.Name)

	err := c.Storage.CreateRepository(normName)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create repository")
		return ctx.Status(fiber.StatusInternalServerError).SendString("failed to create repository")
	}

	logger.Info().Str("repo", normName).Msg("Repository created successfully")
	return ctx.Status(fiber.StatusCreated).SendString("repository created")
}

// ListRepos handles GET requests to retrieve a list of all repositories.
// Returns a JSON array containing the names of all repositories in the storage backend.
//
// Response: 200 OK with JSON array of repository names
func (c *RepoController) ListRepos(ctx *fiber.Ctx) error {
	logger := c.Logger.With().Str("event", "ListRepos").Logger()

	repos, err := c.Storage.ListRepositories()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to list repositories")
		return ctx.Status(fiber.StatusInternalServerError).SendString("failed to list repositories")
	}

	logger.Info().Int("count", len(repos)).Msg("Repositories listed successfully")
	return ctx.Status(fiber.StatusOK).JSON(repos)
}
