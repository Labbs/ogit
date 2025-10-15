package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upGroup, downGroup)
}

func upGroup(ctx context.Context, tx *sql.Tx) error {
	var query string
	dialect, _ := ctx.Value("dbDialect").(string)
	switch dialect {
	case "sqlite":
		query = `
		CREATE TABLE IF NOT EXISTS "group" (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			role TEXT NOT NULL CHECK(role IN ('admin', 'user', 'guest')),
			owner_id TEXT NOT NULL,
			members TEXT, -- SQLite uses TEXT instead of JSONB
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (owner_id) REFERENCES user(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_group_owner_id ON "group"(owner_id);
		CREATE INDEX IF NOT EXISTS idx_group_role ON "group"(role);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_group_name ON "group"(name);
		`
	case "postgres":
		query = `
		-- Create enum types for PostgreSQL
		DO $$ BEGIN
			CREATE TYPE group_role_enum AS ENUM ('admin', 'user', 'guest');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;

		CREATE TABLE IF NOT EXISTS "group" (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			role group_role_enum NOT NULL,
			owner_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
			members JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_group_owner_id ON "group"(owner_id);
		CREATE INDEX IF NOT EXISTS idx_group_role ON "group"(role);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_group_name ON "group"(name);

		-- Trigger to automatically update updated_at
		CREATE OR REPLACE FUNCTION update_group_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		DROP TRIGGER IF EXISTS update_group_updated_at_trigger ON "group";
		CREATE TRIGGER update_group_updated_at_trigger
			BEFORE UPDATE ON "group"
			FOR EACH ROW
			EXECUTE FUNCTION update_group_updated_at();
		`
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	_, err := tx.ExecContext(ctx, query)
	return err
}

func downGroup(ctx context.Context, tx *sql.Tx) error {
	dialect, _ := ctx.Value("dbDialect").(string)

	var query string
	switch dialect {
	case "sqlite":
		query = `DROP TABLE IF EXISTS "group";`
	case "postgres":
		query = `
		DROP TABLE IF EXISTS "group";
		DROP TYPE IF EXISTS group_role_enum;
		DROP FUNCTION IF EXISTS update_group_updated_at();
		`
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	_, err := tx.ExecContext(ctx, query)
	return err
}
