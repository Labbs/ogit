package router

import "github.com/labbs/git-server-s3/internal/api/controller"

func NewRepoRouter(c *Config) {
	gc := controller.RepoController{
		Logger:  c.Logger,
		Storage: c.Storage,
	}

	c.Fiber.Post("/api/repo", gc.CreateRepo)
	c.Fiber.Get("/api/repos", gc.ListRepos)
}
