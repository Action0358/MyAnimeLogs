-- Migration: Update lockout timestamp columns to TIMESTAMP WITH TIME ZONE
-- Purpose: Ensure timezone-aware timestamps for correct lockout expiration handling
-- This prevents subtle bugs where account lockouts might expire prematurely or last longer than intended

-- Update last_failed_login_at to TIMESTAMP WITH TIME ZONE
ALTER TABLE users
ALTER COLUMN last_failed_login_at TYPE TIMESTAMP WITH TIME ZONE;

-- Update locked_until to TIMESTAMP WITH TIME ZONE
ALTER TABLE users
ALTER COLUMN locked_until TYPE TIMESTAMP WITH TIME ZONE;

-- Update comments to reflect the new type
COMMENT ON COLUMN users.last_failed_login_at IS 'Timestamp (with timezone) of the last failed login attempt';
COMMENT ON COLUMN users.locked_until IS 'Timestamp (with timezone) until which the account is locked';
