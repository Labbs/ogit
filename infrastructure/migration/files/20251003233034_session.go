package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSession, downSession)
}

func upSession(ctx context.Context, tx *sql.Tx) error {
	var query string
	dialect, _ := ctx.Value("dbDialect").(string)
	switch dialect {
	case "sqlite":
		query = `
		CREATE TABLE IF NOT EXISTS session (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			user_agent TEXT,
			ip_address TEXT,
			expires_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_session_user_id ON session(user_id);
		CREATE INDEX IF NOT EXISTS idx_session_expires_at ON session(expires_at);
		`
	case "postgres":
		query = `
		CREATE TABLE IF NOT EXISTS session (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
			user_agent TEXT,
			ip_address TEXT,
			expires_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_session_user_id ON session(user_id);
		CREATE INDEX IF NOT EXISTS idx_session_expires_at ON session(expires_at);

		-- Trigger to automatically update updated_at
		CREATE OR REPLACE FUNCTION update_session_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		DROP TRIGGER IF EXISTS update_session_updated_at_trigger ON session;
		CREATE TRIGGER update_session_updated_at_trigger
			BEFORE UPDATE ON session
			FOR EACH ROW
			EXECUTE FUNCTION update_session_updated_at();
		`
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	_, err := tx.ExecContext(ctx, query)
	return err
}

func downSession(ctx context.Context, tx *sql.Tx) error {
	dialect, _ := ctx.Value("dbDialect").(string)

	var query string
	switch dialect {
	case "sqlite":
		query = `DROP TABLE IF EXISTS session;`
	case "postgres":
		query = `
		DROP TABLE IF EXISTS session;
		DROP FUNCTION IF EXISTS update_session_updated_at();
		`
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	_, err := tx.ExecContext(ctx, query)
	return err
}
