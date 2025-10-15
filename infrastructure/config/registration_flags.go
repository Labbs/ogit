package config

import (
	altsrc "github.com/urfave/cli-altsrc/v3"
	altsrcyaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

func RegistrationFlags(cfg *Config) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "registration.enabled",
			Value:       true,
			Destination: &cfg.Registration.Enabled,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("REGISTRATION_ENABLED"),
				altsrcyaml.YAML("registration.enabled", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
		&cli.BoolFlag{
			Name:        "registration.require_email_verification",
			Value:       true,
			Destination: &cfg.Registration.RequireEmailVerification,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("REGISTRATION_REQUIRE_EMAIL_VERIFICATION"),
				altsrcyaml.YAML("registration.require_email_verification", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
		&cli.StringSliceFlag{
			Name:        "registration.domain_whitelist",
			Value:       []string{},
			Destination: &cfg.Registration.DomainWhitelist,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("REGISTRATION_DOMAIN_WHITELIST"),
				altsrcyaml.YAML("registration.domain_whitelist", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
			Usage: "List of allowed email domains for registration (comma separated)",
		},
		&cli.IntFlag{
			Name:        "registration.password_min_length",
			Value:       12,
			Destination: &cfg.Registration.PasswordMinLength,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("REGISTRATION_PASSWORD_MIN_LENGTH"),
				altsrcyaml.YAML("registration.password_min_length", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
		},
		&cli.BoolFlag{
			Name:        "registration.password_complexity",
			Value:       true,
			Destination: &cfg.Registration.PasswordComplexity,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("REGISTRATION_PASSWORD_COMPLEXITY"),
				altsrcyaml.YAML("registration.password_complexity", altsrc.NewStringPtrSourcer(&cfg.ConfigFile)),
			),
			Usage: "Require complex passwords (uppercase, lowercase, numbers, symbols)",
		},
	}
}
