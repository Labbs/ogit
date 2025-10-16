package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upToken, downToken)
}

func upToken(ctx context.Context, tx *sql.Tx) error {
	var query string
	dialect, _ := ctx.Value("dbDialect").(string)
	switch dialect {
	case "sqlite":
		query = `
		CREATE TABLE IF NOT EXISTS token (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			token TEXT NOT NULL UNIQUE,

			type TEXT NOT NULL CHECK(type IN ('user', 'repository')),
			scopes TEXT NOT NULL, -- SQLite uses TEXT instead of JSONB

			user_id TEXT,
			repository_id TEXT,

			active BOOLEAN NOT NULL DEFAULT 1,

			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP,
			
			FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
			FOREIGN KEY (repository_id) REFERENCES repository(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_token_user_id ON token(user_id);
		CREATE INDEX IF NOT EXISTS idx_token_repository_id ON token(repository_id);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_token_token ON token(token);
		CREATE INDEX IF NOT EXISTS idx_token_name ON token(name);
		CREATE INDEX IF NOT EXISTS idx_token_expires_at ON token(expires_at);
		CREATE INDEX IF NOT EXISTS idx_token_type ON token(type);
		CREATE INDEX IF NOT EXISTS idx_token_active ON token(active);
		`
	case "postgres":
		query = `
		-- Create enum types for PostgreSQL
		DO $$ BEGIN
			CREATE TYPE token_type_enum AS ENUM ('user', 'repository');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;

		CREATE TABLE IF NOT EXISTS token (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			description TEXT,
			token TEXT NOT NULL UNIQUE,

			type token_type_enum NOT NULL,
			scopes JSONB NOT NULL,

			user_id UUID REFERENCES "user"(id) ON DELETE CASCADE,
			repository_id UUID REFERENCES repository(id) ON DELETE CASCADE,

			active BOOLEAN NOT NULL DEFAULT TRUE,

			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMPTZ
		);
		CREATE INDEX IF NOT EXISTS idx_token_user_id ON token(user_id);
		CREATE INDEX IF NOT EXISTS idx_token_repository_id ON token(repository_id);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_token_token ON token(token);
		CREATE INDEX IF NOT EXISTS idx_token_name ON token(name);
		CREATE INDEX IF NOT EXISTS idx_token_expires_at ON token(expires_at);
		CREATE INDEX IF NOT EXISTS idx_token_type ON token(type);
		CREATE INDEX IF NOT EXISTS idx_token_active ON token(active);

		-- Trigger to automatically update updated_at (adding updated_at column first)
		ALTER TABLE token ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

		CREATE OR REPLACE FUNCTION update_token_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		DROP TRIGGER IF EXISTS update_token_updated_at_trigger ON token;
		CREATE TRIGGER update_token_updated_at_trigger
			BEFORE UPDATE ON token
			FOR EACH ROW
			EXECUTE FUNCTION update_token_updated_at();
		`
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	_, err := tx.ExecContext(ctx, query)
	return err
}

func downToken(ctx context.Context, tx *sql.Tx) error {
	dialect, _ := ctx.Value("dbDialect").(string)

	var query string
	switch dialect {
	case "sqlite":
		query = `DROP TABLE IF EXISTS token;`
	case "postgres":
		query = `
		DROP TABLE IF EXISTS token;
		DROP TYPE IF EXISTS token_type_enum;
		DROP FUNCTION IF EXISTS update_token_updated_at();
		`
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	_, err := tx.ExecContext(ctx, query)
	return err
}
