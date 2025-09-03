package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/labbs/git-server-s3/pkg/common"
	"github.com/labbs/git-server-s3/pkg/storage"
	"github.com/rs/zerolog"
)

type RepoController struct {
	Logger  zerolog.Logger
	Storage storage.GitRepositoryStorage
}

func (c *RepoController) CreateRepo(ctx *fiber.Ctx) error {
	logger := c.Logger.With().Str("event", "CreateRepo").Logger()

	var req struct {
		Name string `json:"name"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	normName := common.NormalizeRepoPath(req.Name)

	err := c.Storage.CreateRepository(normName)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create repository")
		return ctx.Status(fiber.StatusInternalServerError).SendString("failed to create repository")
	}

	logger.Info().Str("repo", normName).Msg("Repository created successfully")
	return ctx.Status(fiber.StatusCreated).SendString("repository created")
}

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
