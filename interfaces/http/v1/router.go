package v1

import (
	"github.com/labbs/ogit/infrastructure"
	"github.com/labbs/ogit/interfaces/http/v1/auth"
	"github.com/labbs/ogit/interfaces/http/v1/repository"
	"github.com/labbs/ogit/interfaces/http/v1/user"
)

func SetupRouterV1(deps infrastructure.Deps) {
	deps.Logger.Info().Str("component", "http.router.v1").Msg("Setting up API v1 routes")
	grp := deps.Http.FiberOapi.Group("/api/v1")

	authCtrl := auth.Controller{
		Config:    deps.Config,
		Logger:    deps.Logger,
		FiberOapi: grp.Group("/auth"),
		AuthApp:   deps.AuthApp,
	}
	auth.SetupAuthRouter(authCtrl)

	userCtrl := user.Controller{
		Config:    deps.Config,
		Logger:    deps.Logger,
		FiberOapi: grp.Group("/user"),
		UserApp:   deps.UserApp,
	}
	user.SetupUserRouter(userCtrl)

	repoCtrl := repository.Controller{
		Config:        deps.Config,
		Logger:        deps.Logger,
		FiberOapi:     grp.Group("/repository"),
		RepositoryApp: deps.RepoApp,
	}
	repository.SetupRepositoryRouter(repoCtrl)
}
