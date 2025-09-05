package local

import (
	"errors"
	"os"

	"github.com/labbs/git-server-s3/internal/config"
	"github.com/rs/zerolog"
)

type LocalConfig struct {
	Logger zerolog.Logger
}

func (c *LocalConfig) Configure() error {
	c.Logger.Info().Msg("Configuring local storage")
	if config.Storage.Local.Path == "" {
		return errors.New("local storage path is not configured")
	}

	// check if local storage path is a directory
	info, err := os.Stat(config.Storage.Local.Path)
	if os.IsNotExist(err) {
		os.MkdirAll(config.Storage.Local.Path, os.ModePerm)
	} else if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("local storage path is not a directory")
	}

	return nil
}
