package repository

import (
	fiberoapi "github.com/labbs/fiber-oapi"
	"github.com/labbs/ogit/application"
	"github.com/labbs/ogit/infrastructure/config"
	"github.com/rs/zerolog"
)

type Controller struct {
	Config        config.Config
	Logger        zerolog.Logger
	FiberOapi     *fiberoapi.OApiGroup
	RepositoryApp *application.RepositoryApp
}

func SetupRepositoryRouter(controller Controller) {
	fiberoapi.Get(controller.FiberOapi, "/:identifier", controller.GetRepository, fiberoapi.OpenAPIOptions{
		Summary:     "Get repository",
		Description: "Retrieve the details of a specific repository",
		OperationID: "repository.get",
		Tags:        []string{"Repository"},
	})

	fiberoapi.Post(controller.FiberOapi, "", controller.CreateRepository, fiberoapi.OpenAPIOptions{
		Summary:     "Create repository",
		Description: "Create a new repository",
		OperationID: "repository.create",
		Tags:        []string{"Repository"},
	})
}
