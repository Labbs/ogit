package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSshkey, downSshkey)
}

func upSshkey(ctx context.Context, tx *sql.Tx) error {
	var query string
	dialect, _ := ctx.Value("dbDialect").(string)
	switch dialect {
	case "sqlite":
		query = `
		CREATE TABLE IF NOT EXISTS ssh_key (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			public_key TEXT NOT NULL UNIQUE,
			private_key TEXT NOT NULL,
			user_id TEXT NOT NULL,
			owner_type TEXT NOT NULL CHECK(owner_type IN ('user', 'repository')),
			type TEXT NOT NULL CHECK(type IN ('rsa', 'ed25519', 'ecdsa')),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_ssh_key_name ON ssh_key(name);
		CREATE INDEX IF NOT EXISTS idx_ssh_key_user_id ON ssh_key(user_id);
		CREATE INDEX IF NOT EXISTS idx_ssh_key_owner_type ON ssh_key(owner_type);
		CREATE INDEX IF NOT EXISTS idx_ssh_key_type ON ssh_key(type);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_ssh_key_public_key ON ssh_key(public_key);
		`
	case "postgres":
		query = `
		-- Create enum types for PostgreSQL
		DO $$ BEGIN
			CREATE TYPE ssh_key_owner_type_enum AS ENUM ('user', 'repository');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;

		DO $$ BEGIN
			CREATE TYPE ssh_key_type_enum AS ENUM ('rsa', 'ed25519', 'ecdsa');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;

		CREATE TABLE IF NOT EXISTS ssh_key (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL UNIQUE,
			public_key TEXT NOT NULL UNIQUE,
			private_key TEXT NOT NULL,
			user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
			owner_type ssh_key_owner_type_enum NOT NULL,
			type ssh_key_type_enum NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_ssh_key_name ON ssh_key(name);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_ssh_key_public_key ON ssh_key(public_key);
		CREATE INDEX IF NOT EXISTS idx_ssh_key_user_id ON ssh_key(user_id);
		CREATE INDEX IF NOT EXISTS idx_ssh_key_owner_type ON ssh_key(owner_type);
		CREATE INDEX IF NOT EXISTS idx_ssh_key_type ON ssh_key(type);

		-- Trigger to automatically update updated_at
		CREATE OR REPLACE FUNCTION update_ssh_key_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		DROP TRIGGER IF EXISTS update_ssh_key_updated_at_trigger ON ssh_key;
		CREATE TRIGGER update_ssh_key_updated_at_trigger
			BEFORE UPDATE ON ssh_key
			FOR EACH ROW
			EXECUTE FUNCTION update_ssh_key_updated_at();
		`
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	_, err := tx.ExecContext(ctx, query)
	return err
}

func downSshkey(ctx context.Context, tx *sql.Tx) error {
	dialect, _ := ctx.Value("dbDialect").(string)

	var query string
	switch dialect {
	case "sqlite":
		query = `DROP TABLE IF EXISTS ssh_key;`
	case "postgres":
		query = `
		DROP TABLE IF EXISTS ssh_key;
		DROP TYPE IF EXISTS ssh_key_owner_type_enum;
		DROP TYPE IF EXISTS ssh_key_type_enum;
		DROP FUNCTION IF EXISTS update_ssh_key_updated_at();
		`
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	_, err := tx.ExecContext(ctx, query)
	return err
}
