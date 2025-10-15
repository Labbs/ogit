package config

import (
	altsrc "github.com/urfave/cli-altsrc/v3"
	altsrcyaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

func SessionFlags(cfg *Config) []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:        "session.expiration_minutes",
			Value:       60 * 24 * 30, // 1 month
			Destination: &cfg.Session.ExpirationMinutes,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("SESSION_EXPIRATION_MINUTES"),
				altsrcyaml.YAML("session.expiration_minutes", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "session.secret_key",
			Value:       "supersecretkey", // In production, use a secure key and do not hardcode it
			Destination: &cfg.Session.SecretKey,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("SESSION_SECRET_KEY"),
				altsrcyaml.YAML("session.secret_key", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "session.issuer",
			Value:       "ogit", // Issuer name for the session tokens
			Destination: &cfg.Session.Issuer,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("SESSION_ISSUER"),
				altsrcyaml.YAML("session.issuer", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
	}
}
