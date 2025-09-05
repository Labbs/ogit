package flags

import (
	"github.com/labbs/git-server-s3/internal/config"

	"github.com/urfave/cli/v3"
)

func GenericFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Value:       "config.yaml",
			Usage:       "Path to the configuration file",
			Destination: &config.ConfigFile,
		},
	}
}
