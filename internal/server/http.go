package server

import (
	"strconv"

	"github.com/labbs/git-server-s3/internal/api/router"
	"github.com/labbs/git-server-s3/pkg/logger/zerolog"
	"github.com/labbs/git-server-s3/pkg/storage"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	z "github.com/rs/zerolog"
)

type HttpConfig struct {
	Port     int
	HttpLogs bool
	Fiber    *fiber.App
	Logger   z.Logger
	Storage  storage.GitRepositoryStorage
}

func (c *HttpConfig) Configure() {
	fiberConfig := fiber.Config{
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		DisableStartupMessage: true,
	}

	r := fiber.New(fiberConfig)

	if c.HttpLogs {
		r.Use(zerolog.HTTPLogger(c.Logger))
	}

	r.Use(recover.New())
	r.Use(cors.New())
	r.Use(compress.New())
	r.Use(requestid.New())

	r.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"status":  "ok",
			"service": "git-server-s3",
		})
	})

	c.Fiber = r
}

func (c *HttpConfig) NewServer() error {
	c.Configure()

	apirc := router.Config{
		Logger:  c.Logger,
		Fiber:   c.Fiber,
		Storage: c.Storage,
	}

	apirc.Configure()

	c.Logger.Info().Msgf("Starting server on port %d", c.Port)

	err := c.Fiber.Listen(":" + strconv.Itoa(c.Port))
	if err != nil {
		c.Logger.Fatal().Err(err).Msg("Failed to start server")
		return err
	}
	return nil
}
