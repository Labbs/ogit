package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/labbs/git-server-s3/pkg/storage"
	"github.com/rs/zerolog"
)

type Config struct {
	Logger  zerolog.Logger
	Fiber   *fiber.App
	Storage storage.GitRepositoryStorage
}

func (c *Config) Configure() {
	c.Logger.Info().Msg("Configuring API routes")

	NewGitRouter(c)
	NewRepoRouter(c)
}
