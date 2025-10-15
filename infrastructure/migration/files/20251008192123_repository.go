package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upRepository, downRepository)
}

func upRepository(ctx context.Context, tx *sql.Tx) error {
	var query string
	dialect, _ := ctx.Value("dbDialect").(string)
	switch dialect {
	case "sqlite":
		query = `
		CREATE TABLE IF NOT EXISTS repository (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			slug TEXT NOT NULL UNIQUE,
			default_branch TEXT NOT NULL DEFAULT 'main',
			is_archived BOOLEAN DEFAULT FALSE,
			archived_at TIMESTAMP,
			is_private BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_repository_slug ON repository(slug);
		CREATE INDEX IF NOT EXISTS idx_repository_is_private ON repository(is_private);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_repository_name ON repository(name);
		CREATE INDEX IF NOT EXISTS idx_repository_is_archived ON repository(is_archived);
		`
	case "postgres":
		query = `
		CREATE TABLE IF NOT EXISTS repository (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			description TEXT,
			slug TEXT NOT NULL UNIQUE,
			default_branch TEXT NOT NULL DEFAULT 'main',
			is_archived BOOLEAN DEFAULT FALSE,
			archived_at TIMESTAMPTZ,
			is_private BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_repository_slug ON repository(slug);
		CREATE INDEX IF NOT EXISTS idx_repository_is_private ON repository(is_private);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_repository_name ON repository(name);
		CREATE INDEX IF NOT EXISTS idx_repository_is_archived ON repository(is_archived);

		-- Trigger to automatically update updated_at
		CREATE OR REPLACE FUNCTION update_repository_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		DROP TRIGGER IF EXISTS update_repository_updated_at_trigger ON repository;
		CREATE TRIGGER update_repository_updated_at_trigger
			BEFORE UPDATE ON repository
			FOR EACH ROW
			EXECUTE FUNCTION update_repository_updated_at();
		`
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	_, err := tx.ExecContext(ctx, query)
	return err
}

func downRepository(ctx context.Context, tx *sql.Tx) error {
	dialect, _ := ctx.Value("dbDialect").(string)

	var query string
	switch dialect {
	case "sqlite":
		query = `DROP TABLE IF EXISTS repository;`
	case "postgres":
		query = `
		DROP TABLE IF EXISTS repository;
		DROP FUNCTION IF EXISTS update_repository_updated_at();
		`
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	_, err := tx.ExecContext(ctx, query)
	return err
}
