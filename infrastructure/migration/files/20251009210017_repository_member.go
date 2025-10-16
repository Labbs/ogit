package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upRepositoryMember, downRepositoryMember)
}

func upRepositoryMember(ctx context.Context, tx *sql.Tx) error {
	var query string
	dialect, _ := ctx.Value("dbDialect").(string)
	switch dialect {
	case "sqlite":
		query = `
		CREATE TABLE IF NOT EXISTS repository_member (
			id TEXT PRIMARY KEY,
			repository_id TEXT NOT NULL,

			member_type TEXT NOT NULL CHECK(member_type IN ('user', 'group')),
			user_id TEXT,
			group_id TEXT,

			access_type TEXT NOT NULL CHECK(access_type IN ('none', 'reporter', 'developer', 'maintainer', 'owner')),

			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,

			-- Constraint to ensure only one of user_id or group_id is set based on member_type
			CHECK (
				(member_type = 'user' AND user_id IS NOT NULL AND group_id IS NULL) OR
				(member_type = 'group' AND group_id IS NOT NULL AND user_id IS NULL)
			)
		);
		-- Index for faster lookups
		CREATE INDEX IF NOT EXISTS idx_repository_member_repo ON repository_member(repository_id);
		CREATE INDEX IF NOT EXISTS idx_repository_member_user ON repository_member(user_id) WHERE user_id IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_repository_member_group ON repository_member(group_id) WHERE group_id IS NOT NULL;

		-- Composite indexes for access check queries
		CREATE INDEX IF NOT EXISTS idx_repo_user_access ON repository_member(repository_id, user_id) WHERE member_type = 'user';
		CREATE INDEX IF NOT EXISTS idx_repo_group_access ON repository_member(repository_id, group_id) WHERE member_type = 'group';

		-- Index to find all repos for a user or group
		CREATE INDEX IF NOT EXISTS idx_user_repositories ON repository_member(user_id, repository_id) WHERE member_type = 'user';
		CREATE INDEX IF NOT EXISTS idx_group_repositories ON repository_member(group_id, repository_id) WHERE member_type = 'group';
		
		-- Unicity constraints to prevent duplicate memberships
		CREATE UNIQUE INDEX IF NOT EXISTS unique_repo_user ON repository_member(repository_id, user_id) WHERE member_type = 'user';
		CREATE UNIQUE INDEX IF NOT EXISTS unique_repo_group ON repository_member(repository_id, group_id) WHERE member_type = 'group';
		`
	case "postgres":
		query = `
		-- Create enum types for PostgreSQL
		DO $$ BEGIN
			CREATE TYPE member_type_enum AS ENUM ('user', 'group');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;

		DO $$ BEGIN
			CREATE TYPE access_type_enum AS ENUM ('none', 'reporter', 'developer', 'maintainer', 'owner');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;

		CREATE TABLE IF NOT EXISTS repository_member (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			repository_id UUID NOT NULL REFERENCES repository(id) ON DELETE CASCADE,

			member_type member_type_enum NOT NULL,
			user_id UUID REFERENCES "user"(id) ON DELETE CASCADE,
			group_id UUID REFERENCES "group"(id) ON DELETE CASCADE,

			access_type access_type_enum NOT NULL,

			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

			-- Constraint to ensure only one of user_id or group_id is set based on member_type
			CONSTRAINT chk_member_type_consistency CHECK (
				(member_type = 'user' AND user_id IS NOT NULL AND group_id IS NULL) OR
				(member_type = 'group' AND group_id IS NOT NULL AND user_id IS NULL)
			)
		);

		-- Index for faster lookups
		CREATE INDEX IF NOT EXISTS idx_repository_member_repo ON repository_member(repository_id);
		CREATE INDEX IF NOT EXISTS idx_repository_member_user ON repository_member(user_id) WHERE user_id IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_repository_member_group ON repository_member(group_id) WHERE group_id IS NOT NULL;

		-- Composite indexes for access check queries
		CREATE INDEX IF NOT EXISTS idx_repo_user_access ON repository_member(repository_id, user_id) WHERE member_type = 'user';
		CREATE INDEX IF NOT EXISTS idx_repo_group_access ON repository_member(repository_id, group_id) WHERE member_type = 'group';

		-- Index to find all repos for a user or group
		CREATE INDEX IF NOT EXISTS idx_user_repositories ON repository_member(user_id, repository_id) WHERE member_type = 'user';
		CREATE INDEX IF NOT EXISTS idx_group_repositories ON repository_member(group_id, repository_id) WHERE member_type = 'group';
		
		-- Unicity constraints to prevent duplicate memberships
		CREATE UNIQUE INDEX IF NOT EXISTS unique_repo_user ON repository_member(repository_id, user_id) WHERE member_type = 'user';
		CREATE UNIQUE INDEX IF NOT EXISTS unique_repo_group ON repository_member(repository_id, group_id) WHERE member_type = 'group';

		-- Trigger to automatically update updated_at
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		DROP TRIGGER IF EXISTS update_repository_member_updated_at ON repository_member;
		CREATE TRIGGER update_repository_member_updated_at
			BEFORE UPDATE ON repository_member
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
		`
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	_, err := tx.ExecContext(ctx, query)
	return err
}

func downRepositoryMember(ctx context.Context, tx *sql.Tx) error {
	dialect, _ := ctx.Value("dbDialect").(string)

	var query string
	switch dialect {
	case "sqlite":
		query = `DROP TABLE IF EXISTS repository_member;`
	case "postgres":
		query = `
		DROP TABLE IF EXISTS repository_member;
		DROP TYPE IF EXISTS member_type_enum;
		DROP TYPE IF EXISTS access_type_enum;
		DROP FUNCTION IF EXISTS update_updated_at_column();
		`
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	_, err := tx.ExecContext(ctx, query)
	return err
}
