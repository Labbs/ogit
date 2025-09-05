package router

import "github.com/labbs/git-server-s3/internal/api/controller"

func NewGitRouter(c *Config) {
	gc := controller.GitController{
		Logger:  c.Logger,
		Storage: c.Storage,
	}

	c.Fiber.Get("/:repo/info/refs", gc.InfoRefs)
	c.Fiber.Post("/:repo/git-upload-pack", gc.HandleUploadPack)
	c.Fiber.Post("/:repo/git-receive-pack", gc.HandleReceivePack)
}
