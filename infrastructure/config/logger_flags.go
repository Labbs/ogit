package config

import (
	altsrc "github.com/urfave/cli-altsrc/v3"
	altsrcyaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

func LoggerFlags(cfg *Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "logger.level",
			Aliases:     []string{"l"},
			Value:       "info",
			Destination: &cfg.Logger.Level,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("LOGGER_LEVEL"),
				altsrcyaml.YAML("logger.level", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
		&cli.BoolFlag{
			Name:        "logger.pretty",
			Value:       false,
			Destination: &cfg.Logger.Pretty,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("LOGGER_PRETTY"),
				altsrcyaml.YAML("logger.pretty", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
	}
}
