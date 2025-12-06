-- Rollback migration: Revert VARCHAR columns back to TEXT
-- Use this if you need to rollback the changes from 012_update_user_text_fields_to_varchar.sql

-- Revert description from VARCHAR(500) to TEXT
ALTER TABLE users ALTER COLUMN description TYPE TEXT;

-- Revert favorite_anime from VARCHAR(200) to TEXT
ALTER TABLE users ALTER COLUMN favorite_anime TYPE TEXT;

-- Remove comments
COMMENT ON COLUMN users.description IS NULL;
COMMENT ON COLUMN users.favorite_anime IS NULL;

-- Log the rollback
DO $$
BEGIN
    RAISE NOTICE 'Successfully rolled back users table columns to TEXT type';
END $$;
