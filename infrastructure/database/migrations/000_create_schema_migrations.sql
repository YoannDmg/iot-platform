-- Migration: Create schema_migrations table
-- Description: Tracks which migrations have been applied
-- This migration MUST be idempotent as it bootstraps the migration system

CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE schema_migrations IS 'Tracks database schema migrations that have been applied';
COMMENT ON COLUMN schema_migrations.version IS 'Migration version identifier (e.g., 001_create_devices_table)';
COMMENT ON COLUMN schema_migrations.applied_at IS 'Timestamp when the migration was applied';
