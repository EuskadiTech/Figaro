-- Migration: Remove default center fields from users table (rollback)
-- Version: 006

-- Remove index first
DROP INDEX IF EXISTS idx_users_default_center_id;

-- Remove columns
ALTER TABLE users DROP COLUMN force_default_center;
ALTER TABLE users DROP COLUMN default_center_id;