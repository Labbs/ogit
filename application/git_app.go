package application

import (
	"github.com/labbs/ogit/infrastructure/config"
	"github.com/labbs/ogit/infrastructure/storage"
	"github.com/rs/zerolog"
)

type GitApp struct {
	Config        config.Config
	Logger        zerolog.Logger
	StorageConfig storage.Config
}

func NewGitApp(config config.Config, logger zerolog.Logger, storageConfig storage.Config) *GitApp {
	return &GitApp{
		Config:        config,
		Logger:        logger,
		StorageConfig: storageConfig,
	}
}
