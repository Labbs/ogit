package git

import (
	"github.com/gofiber/fiber/v2"
	"github.com/labbs/ogit/infrastructure/config"
	"github.com/labbs/ogit/infrastructure/storage"
	"github.com/rs/zerolog"
)

type Controller struct {
	Config        config.Config
	Logger        zerolog.Logger
	Fiber         *fiber.App
	StorageConfig storage.Config
}

func SetupGitRouter(controller Controller) {
	controller.Fiber.Get("/:repo/info/refs", controller.InfoRefs)
	controller.Fiber.Post("/:repo/git-upload-pack", controller.HandleUploadPack)
	controller.Fiber.Post("/:repo/git-receive-pack", controller.HandleReceivePack)
}
