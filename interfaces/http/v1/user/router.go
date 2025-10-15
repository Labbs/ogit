package user

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
	UserApp   *application.UserApp
}

func SetupUserRouter(controller Controller) {
	fiberoapi.Get(controller.FiberOapi, "/profile", controller.GetProfile, fiberoapi.OpenAPIOptions{
		Summary:     "Get user profile",
		Description: "Retrieve the profile of the authenticated user",
		OperationID: "user.getProfile",
		Tags:        []string{"User"},
	})
}
