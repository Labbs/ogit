package auth

import (
	fiberoapi "github.com/labbs/fiber-oapi"
	"github.com/labbs/ogit/application"
	"github.com/labbs/ogit/infrastructure/config"
	"github.com/rs/zerolog"
)

type Controller struct {
	Config    config.Config
	Logger    zerolog.Logger
	FiberOapi *fiberoapi.OApiGroup
	AuthApp   *application.AuthApp
}

func SetupAuthRouter(controller Controller) {
	fiberoapi.Post(controller.FiberOapi, "/login", controller.Login, fiberoapi.OpenAPIOptions{
		Summary:     "Login user",
		Description: "Authenticate user and return a token",
		OperationID: "auth.login",
		Tags:        []string{"Auth"},
		Security:    "disabled",
	})

	fiberoapi.Get(controller.FiberOapi, "/logout", controller.Logout, fiberoapi.OpenAPIOptions{
		Summary:     "Logout user",
		Description: "Invalidate user session",
		OperationID: "auth.logout",
		Tags:        []string{"Auth"},
	})

	fiberoapi.Post(controller.FiberOapi, "/register", controller.Register, fiberoapi.OpenAPIOptions{
		Summary:     "Register user",
		Description: "Register a new user",
		OperationID: "auth.register",
		Tags:        []string{"Auth"},
		Security:    "disabled",
	})
}
