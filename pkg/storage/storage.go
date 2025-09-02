package storage

import (
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labbs/git-server-s3/internal/config"
	"github.com/labbs/git-server-s3/pkg/storage/s3"
	"github.com/rs/zerolog"
)

type Storage struct {
	S3Client *awss3.Client
	Logger   zerolog.Logger
}

func (c *Storage) Configure() error {
	logger := c.Logger.With().Str("component", "storage").Logger()

	switch config.Storage.Type {
	case "s3":
		logger.Info().Msg("Configuring S3 storage")
		var s3Config s3.S3Config
		s3Config.Logger = logger
		s3Config.Configure()
		c.S3Client = s3Config.Client
	case "local":
		logger.Info().Msg("Configuring local storage")
	default:
		logger.Warn().Msg("Unknown storage type, using in-memory storage")
	}

	return nil
}
