package infrastructure

import (
	"github.com/labbs/ogit/application"
	"github.com/labbs/ogit/infrastructure/config"
	"github.com/labbs/ogit/infrastructure/cronscheduler"
	"github.com/labbs/ogit/infrastructure/database"
	"github.com/labbs/ogit/infrastructure/http"
	"github.com/labbs/ogit/infrastructure/storage"
	"github.com/rs/zerolog"
)

type Deps struct {
	Config        config.Config
	Logger        zerolog.Logger
	Http          http.Config
	CronScheduler cronscheduler.Config
	Database      database.Config
	Storage       storage.Config

	UserApp    *application.UserApp
	SessionApp *application.SessionApp
	AuthApp    *application.AuthApp
	RepoApp    *application.RepositoryApp
	GitApp     *application.GitApp
	TokenApp   *application.TokenApp
}
