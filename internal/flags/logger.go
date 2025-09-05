package flags

import (
	"github.com/labbs/git-server-s3/internal/config"

	altsrc "github.com/urfave/cli-altsrc/v3"
	altsrcyaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

func LoggerFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "logger.level",
			Aliases:     []string{"l"},
			Value:       "info",
			Destination: &config.Logger.Level,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("LOGGER_LEVEL"),
				altsrcyaml.YAML("logger.level", altsrc.NewStringPtrSourcer(&config.ConfigFile)),
			),
		},
		&cli.BoolFlag{
			Name:        "logger.pretty",
			Value:       false,
			Destination: &config.Logger.Pretty,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("LOGGER_PRETTY"),
				altsrcyaml.YAML("logger.pretty", altsrc.NewStringPtrSourcer(&config.ConfigFile)),
			),
		},
	}
}
