package flags

import (
	"github.com/labbs/git-server-s3/internal/config"

	altsrc "github.com/urfave/cli-altsrc/v3"
	altsrcyaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

func StorageFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "storage.type",
			Aliases:     []string{"st"},
			Destination: &config.Storage.Type,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("STORAGE_TYPE"),
				altsrcyaml.YAML("storage.type", altsrc.NewStringPtrSourcer(&config.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "storage.s3.bucket",
			Aliases:     []string{"ssb"},
			Destination: &config.Storage.S3.Bucket,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("STORAGE_S3_BUCKET"),
				altsrcyaml.YAML("storage.s3.bucket", altsrc.NewStringPtrSourcer(&config.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "storage.s3.endpoint",
			Aliases:     []string{"sse"},
			Destination: &config.Storage.S3.Endpoint,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("STORAGE_S3_ENDPOINT"),
				altsrcyaml.YAML("storage.s3.endpoint", altsrc.NewStringPtrSourcer(&config.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "storage.s3.access-key",
			Aliases:     []string{"ssa"},
			Destination: &config.Storage.S3.AccessKey,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("STORAGE_S3_ACCESS_KEY"),
				altsrcyaml.YAML("storage.s3.access-key", altsrc.NewStringPtrSourcer(&config.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "storage.s3.secret-key",
			Aliases:     []string{"sss"},
			Destination: &config.Storage.S3.SecretKey,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("STORAGE_S3_SECRET_KEY"),
				altsrcyaml.YAML("storage.s3.secret-key", altsrc.NewStringPtrSourcer(&config.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "storage.s3.region",
			Aliases:     []string{"ssr"},
			Destination: &config.Storage.S3.Region,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("STORAGE_S3_REGION"),
				altsrcyaml.YAML("storage.s3.region", altsrc.NewStringPtrSourcer(&config.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "storage.local.path",
			Aliases:     []string{"slp"},
			Destination: &config.Storage.Local.Path,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("STORAGE_LOCAL_PATH"),
				altsrcyaml.YAML("storage.local.path", altsrc.NewStringPtrSourcer(&config.ConfigFile)),
			),
		},
	}
}
