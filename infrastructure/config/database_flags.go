package config

import (
	altsrc "github.com/urfave/cli-altsrc/v3"
	altsrcyaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

func DatabaseFlags(cfg *Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "database.dialect",
			Usage:       "The database dialect (e.g., sqlite, postgres, mysql)",
			Aliases:     []string{"db.dialect"},
			Value:       "sqlite",
			Destination: &cfg.Database.Dialect,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("DATABASE_DIALECT"),
				altsrcyaml.YAML("database.dialect", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
		&cli.StringFlag{
			Name:        "database.dsn",
			Usage:       "The database DSN (Data Source Name)",
			Aliases:     []string{"db.dsn"},
			Value:       "./database.sqlite",
			Destination: &cfg.Database.DSN,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("DATABASE_DSN"),
				altsrcyaml.YAML("database.dsn", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
	}
}
