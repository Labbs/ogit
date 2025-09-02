package flags

import (
	"github.com/labbs/git-server-s3/internal/config"

	altsrc "github.com/urfave/cli-altsrc/v3"
	altsrcyaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

func ServerFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:        "http.port",
			Aliases:     []string{"p"},
			Value:       8080,
			Destination: &config.Server.Port,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("HTTP_PORT"),
				altsrcyaml.YAML("http.port", altsrc.NewStringPtrSourcer(&config.ConfigFile)),
			),
		},
		&cli.BoolFlag{
			Name:        "http.logs",
			Aliases:     []string{"l"},
			Value:       false,
			Destination: &config.Server.HttpLogs,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("HTTP_LOGS"),
				altsrcyaml.YAML("http.logs", altsrc.NewStringPtrSourcer(&config.ConfigFile)),
			),
		},
	}
}
