package database

import (
	"github.com/labbs/ogit/infrastructure/config"
	zerologadapter "github.com/labbs/ogit/infrastructure/logger/zerolog"

	z "github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	Db *gorm.DB
}

// Configure sets up the database connection based on the provided configuration and logger.
// It supports sqlite, postgres, and mysql databases.
// Will return an error if the connection cannot be established (fatal)
func Configure(_cfg config.Config, logger z.Logger) (Config, error) {
	logger = logger.With().Str("component", "infrastructure.database").Logger()
	gormLogger := zerologadapter.NewGormLogger(logger)

	var db *gorm.DB
	var err error

	// Check if the database is managed
	switch _cfg.Database.Dialect {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(_cfg.Database.DSN), &gorm.Config{Logger: gormLogger})
	case "postgres":
		db, err = gorm.Open(postgres.Open(_cfg.Database.DSN), &gorm.Config{Logger: gormLogger})
	default:
		logger.Fatal().Str("event", "database.configure.invalid_dialect").Msg("Invalid database type")
	}
	if err != nil {
		return Config{}, err
	}

	return Config{Db: db}, nil
}
