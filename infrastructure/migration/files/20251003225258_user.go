package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upUser, downUser)
}

func upUser(ctx context.Context, tx *sql.Tx) error {
	var query string
	dialect, _ := ctx.Value("dbDialect").(string)
	switch dialect {
	case "sqlite":
		query = `
		CREATE TABLE IF NOT EXISTS user (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			avatar_url TEXT,
			preferences TEXT, -- SQLite uses TEXT instead of JSONB
			active BOOLEAN DEFAULT TRUE,
			role TEXT CHECK(role IN ('admin', 'user', 'guest')) DEFAULT 'user',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON user(username);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON user(email);
		CREATE INDEX IF NOT EXISTS idx_user_active ON user(active);
		CREATE INDEX IF NOT EXISTS idx_user_role ON user(role);
		`
	case "postgres":
		query = `
		-- Create enum types for PostgreSQL
		DO $$ BEGIN
			CREATE TYPE user_role_enum AS ENUM ('admin', 'user', 'guest');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;

		CREATE TABLE IF NOT EXISTS "user" (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			avatar_url TEXT,
			preferences JSONB,
			active BOOLEAN DEFAULT TRUE,
			role user_role_enum DEFAULT 'user',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON "user"(username);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON "user"(email);
		CREATE INDEX IF NOT EXISTS idx_users_active ON "user"(active);
		CREATE INDEX IF NOT EXISTS idx_user_role ON "user"(role);

		-- Trigger to automatically update updated_at
		CREATE OR REPLACE FUNCTION update_user_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		DROP TRIGGER IF EXISTS update_user_updated_at_trigger ON "user";
		CREATE TRIGGER update_user_updated_at_trigger
			BEFORE UPDATE ON "user"
			FOR EACH ROW
			EXECUTE FUNCTION update_user_updated_at();
		`
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	_, err := tx.ExecContext(ctx, query)
	return err
}

func downUser(ctx context.Context, tx *sql.Tx) error {
	dialect, _ := ctx.Value("dbDialect").(string)

	var query string
	switch dialect {
	case "sqlite":
		query = `DROP TABLE IF EXISTS user;`
	case "postgres":
		query = `
		DROP TABLE IF EXISTS "user";
		DROP TYPE IF EXISTS user_role_enum;
		DROP FUNCTION IF EXISTS update_user_updated_at();
		`
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	_, err := tx.ExecContext(ctx, query)
	return err
}
