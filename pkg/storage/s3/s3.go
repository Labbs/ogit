package s3

import (
	"context"
	"os"

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
	// Set AWS environment variables to disable automatic checksums for S3-compatible services
	os.Setenv("AWS_REQUEST_CHECKSUM_CALCULATION", "WHEN_REQUIRED")
	os.Setenv("AWS_RESPONSE_CHECKSUM_VALIDATION", "WHEN_REQUIRED")

	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(),
		awsCfg.WithRegion(config.Storage.S3.Region),
		awsCfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.Storage.S3.AccessKey,
			config.Storage.S3.SecretKey,
			"",
		)),
	)
	if err != nil {
		c.Logger.Fatal().Err(err).Str("event", "s3.configure.client").Msg("Failed to configure S3 client")
	}

	// Configure client with custom endpoint and disable checksums for S3-compatible services
	c.Client = awss3.NewFromConfig(cfg, func(o *awss3.Options) {
		o.BaseEndpoint = aws.String(config.Storage.S3.Endpoint)
		o.UsePathStyle = true // Important pour Outscale et autres services S3-compatibles
		// Disable checksums for S3-compatible services that don't support them
		o.DisableMultiRegionAccessPoints = true
		// Disable request and response checksums
		o.ClientLogMode = 0 // Reduce logging if needed
	})
	return nil
}
