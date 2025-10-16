package migration

import (
	"context"
	"embed"

	"github.com/labbs/ogit/infrastructure/logger/zerolog"

	_ "github.com/labbs/ogit/infrastructure/migration/files"

	"github.com/pressly/goose/v3"
	z "github.com/rs/zerolog"
	"gorm.io/gorm"
)

//go:embed files/*
var migrationFiles embed.FS

func RunMigration(l z.Logger, db *gorm.DB) error {
	logger := l.With().Str("component", "infrastructure.migration").Logger()
	goose.SetBaseFS(migrationFiles)
	goose.SetLogger(&zerolog.ZerologGooseAdapter{Logger: logger})

	// Set the dialect following the gorm dialect
	dbDialect := db.Dialector.Name()

	if err := goose.SetDialect(dbDialect); err != nil {
		logger.Error().Err(err).Str("event", "migration.failed_to_set_dialect").Msg("Failed to set dialect")
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error().Err(err).Str("event", "migration.failed_to_get_sql_db").Msg("Failed to get sql db")
		return err
	}

	ctx := context.WithValue(context.Background(), "dbDialect", dbDialect)

	if err := goose.UpContext(ctx, sqlDB, "files"); err != nil {
		if err.Error() != "no change" {
			logger.Error().Err(err).Str("event", "migration.failed_to_run_migrations").Msg("Failed to run migrations")
			return err
		}
	}

	return nil
}
