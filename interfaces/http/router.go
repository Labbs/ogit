package http

import (
	"github.com/labbs/ogit/infrastructure"
	"github.com/labbs/ogit/interfaces/http/git"
	v1 "github.com/labbs/ogit/interfaces/http/v1"
)

func SetupRoutes(deps infrastructure.Deps) {
	logger := deps.Logger.With().Str("component", "http.router").Logger()
	logger.Info().Str("event", "setup_routes").Msg("Setting up HTTP routes")

	// Setup system routes (health, metrics, etc.)
	setupSystemRoutes(deps)

	// Setup git routes
	gitCtrl := git.Controller{
		Config:        deps.Config,
		Logger:        deps.Logger,
		Fiber:         deps.Http.Fiber,
		StorageConfig: deps.Storage,
	}
	git.SetupGitRouter(gitCtrl)

	// Setup v1 routes
	v1.SetupRouterV1(deps)
}
