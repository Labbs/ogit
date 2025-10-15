package config

import (
	"github.com/urfave/cli/v3"
)

func GenericFlags(cfg *Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Value:       "config.yaml",
			Usage:       "Path to the configuration file",
			Destination: &cfg.ConfigFile,
		},
	}
}
