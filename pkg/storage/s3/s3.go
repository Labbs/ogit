package s3

import (
	"context"

	"github.com/labbs/git-server-s3/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog"
)

type S3Config struct {
	Logger zerolog.Logger
	Client *awss3.Client
}

func (c *S3Config) Configure() error {
	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(),
		awsCfg.WithRegion(config.Storage.S3.Region),
		awsCfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.Storage.S3.AccessKey,
			config.Storage.S3.SecretKey,
			"",
		)),
		// TODO: Check what is alternative for WithEndpointResolverWithOptions, this function is deprecated
		awsCfg.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           config.Storage.S3.Endpoint,
					SigningRegion: region,
				}, nil
			})),
	)
	if err != nil {
		c.Logger.Fatal().Err(err).Str("event", "s3.configure.client").Msg("Failed to configure S3 client")
	}

	c.Client = awss3.NewFromConfig(cfg)
	return nil
}
