-- Migration: Add account lockout fields to users table
-- Purpose: Track login failures and implement account lockout for security

ALTER TABLE users
ADD COLUMN IF NOT EXISTS failed_login_attempts INT DEFAULT 0,
ADD COLUMN IF NOT EXISTS last_failed_login_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP;

-- Create index for performance optimization
CREATE INDEX IF NOT EXISTS idx_users_locked_until ON users(locked_until) WHERE locked_until IS NOT NULL;

-- Add comment for documentation
COMMENT ON COLUMN users.failed_login_attempts IS 'Number of consecutive failed login attempts';
COMMENT ON COLUMN users.last_failed_login_at IS 'Timestamp of the last failed login attempt';
COMMENT ON COLUMN users.locked_until IS 'Timestamp until which the account is locked';
