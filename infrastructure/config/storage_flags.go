package config

import (
	altsrc "github.com/urfave/cli-altsrc/v3"
	altsrcyaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

func StorageFlags(cfg *Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name: "storage.type",
			Usage: "The storage type (e.g., local, s3, gcs). " +
				"Local storage stores files on the local filesystem, " +
				"S3 uses Amazon S3-compatible storage services, " +
				"and GCS uses Google Cloud Storage.",
			Aliases:     []string{"storage"},
			Value:       "local",
			Destination: &cfg.Storage.Type,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("STORAGE_TYPE"),
				altsrcyaml.YAML("storage.type", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "storage.local.path",
			Usage:       "The local storage path (used when storage.type is 'local')",
			Aliases:     []string{"storage.local", "storage.path"},
			Value:       "./repositories",
			Destination: &cfg.Storage.Local.Path,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("STORAGE_LOCAL_PATH"),
				altsrcyaml.YAML("storage.local.path", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "storage.s3.bucket",
			Usage:       "The S3 bucket name (used when storage.type is 's3')",
			Aliases:     []string{"storage.s3"},
			Value:       "",
			Destination: &cfg.Storage.S3.Bucket,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("STORAGE_S3_BUCKET"),
				altsrcyaml.YAML("storage.s3.bucket", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "storage.s3.endpoint",
			Usage:       "The S3 endpoint URL (used when storage.type is 's3')",
			Aliases:     []string{"storage.s3.endpoint"},
			Value:       "",
			Destination: &cfg.Storage.S3.Endpoint,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("STORAGE_S3_ENDPOINT"),
				altsrcyaml.YAML("storage.s3.endpoint", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "storage.s3.accesskey",
			Usage:       "The S3 access key (used when storage.type is 's3')",
			Aliases:     []string{"storage.s3.accesskey"},
			Value:       "",
			Destination: &cfg.Storage.S3.AccessKey,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("STORAGE_S3_ACCESSKEY"),
				altsrcyaml.YAML("storage.s3.accesskey", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "storage.s3.secretkey",
			Usage:       "The S3 secret key (used when storage.type is 's3')",
			Aliases:     []string{"storage.s3.secretkey"},
			Value:       "",
			Destination: &cfg.Storage.S3.SecretKey,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("STORAGE_S3_SECRETKEY"),
				altsrcyaml.YAML("storage.s3.secretkey", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "storage.s3.region",
			Usage:       "The S3 region (used when storage.type is 's3')",
			Aliases:     []string{"storage.s3.region"},
			Value:       "us-east-1",
			Destination: &cfg.Storage.S3.Region,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("STORAGE_S3_REGION"),
				altsrcyaml.YAML("storage.s3.region", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
	}
}
