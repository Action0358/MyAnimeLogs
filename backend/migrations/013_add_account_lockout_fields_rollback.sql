-- Rollback: Remove account lockout fields from users table

-- Drop index
DROP INDEX IF EXISTS idx_users_locked_until;

-- Remove columns
ALTER TABLE users
DROP COLUMN IF EXISTS failed_login_attempts,
DROP COLUMN IF EXISTS last_failed_login_at,
DROP COLUMN IF EXISTS locked_until;
