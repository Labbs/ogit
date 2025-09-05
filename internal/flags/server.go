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
		&cli.BoolFlag{
			Name:        "ssh.enabled",
			Value:       false,
			Destination: &config.SSH.Enabled,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("SSH_ENABLED"),
				altsrcyaml.YAML("ssh.enabled", altsrc.NewStringPtrSourcer(&config.ConfigFile)),
			),
		},
		&cli.IntFlag{
			Name:        "ssh.port",
			Value:       2222,
			Destination: &config.SSH.Port,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("SSH_PORT"),
				altsrcyaml.YAML("ssh.port", altsrc.NewStringPtrSourcer(&config.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "ssh.hostkey",
			Value:       "./ssh_host_key",
			Destination: &config.SSH.HostKeyPath,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("SSH_HOST_KEY_PATH"),
				altsrcyaml.YAML("ssh.hostkey", altsrc.NewStringPtrSourcer(&config.ConfigFile)),
			),
		},
		&cli.BoolFlag{
			Name:        "debug.endpoints",
			Value:       false,
			Destination: &config.Debug.Endpoints,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("DEBUG_ENDPOINTS"),
				altsrcyaml.YAML("debug.endpoints", altsrc.NewStringPtrSourcer(&config.ConfigFile)),
			),
		},
	}
}
