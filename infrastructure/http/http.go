package http

import (
	"github.com/labbs/ogit/application"
	"github.com/labbs/ogit/infrastructure/config"
	"github.com/labbs/ogit/infrastructure/logger/zerolog"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	fiberoapi "github.com/labbs/fiber-oapi"
	z "github.com/rs/zerolog"
)

type Config struct {
	Fiber     *fiber.App
	FiberOapi *fiberoapi.OApiApp
}

// Configure sets up the HTTP server (fiber) with the provided configuration and logger.
// The FiberOapi instance is also configured for OpenAPI support and exposed documentation.
// Will return an error if the server cannot be created (fatal)
func Configure(_cfg config.Config, logger z.Logger, session application.SessionApp, enableIU bool) (Config, error) {
	var c Config
	fiberConfig := fiber.Config{
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		DisableStartupMessage: true,
	}

	r := fiber.New(fiberConfig)

	if _cfg.Server.HttpLogs {
		r.Use(zerolog.HTTPLogger(logger))
	}

	r.Use(recover.New())
	r.Use(cors.New())
	r.Use(compress.New())
	r.Use(requestid.New())

	oapiConfig := fiberoapi.Config{
		EnableValidation:  true,
		EnableOpenAPIDocs: true,
		OpenAPIDocsPath:   "/documentation",
		OpenAPIJSONPath:   "/api-spec.json",
		OpenAPIYamlPath:   "/api-spec.yaml",
		AuthService:       &session,
		SecuritySchemes: map[string]fiberoapi.SecurityScheme{
			"bearerAuth": {
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
				Description:  "JWT Bearer token",
			},
		},
		DefaultSecurity: []map[string][]string{
			{"bearerAuth": {}},
		},
	}

	authMiddleware := fiberoapi.BearerTokenMiddleware(&session)
	r.Use(fiberoapi.ConditionalAuthMiddleware(authMiddleware,
		"/documentation",
		"/api-spec.json",
		"/api-spec.yaml",
		"/health", // autres routes à exclure si nécessaire
	))

	c.FiberOapi = fiberoapi.New(r, oapiConfig)
	c.Fiber = r

	return c, nil
}
