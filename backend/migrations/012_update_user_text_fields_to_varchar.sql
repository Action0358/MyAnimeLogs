-- Update users table text fields to VARCHAR with appropriate length limits
-- This ensures data integrity at the database level

-- Update description from TEXT to VARCHAR(500)
ALTER TABLE users ALTER COLUMN description TYPE VARCHAR(500);

-- Update favorite_anime from TEXT to VARCHAR(200)
ALTER TABLE users ALTER COLUMN favorite_anime TYPE VARCHAR(200);

-- Add comments to document the constraints
COMMENT ON COLUMN users.description IS 'User profile description (max 500 characters)';
COMMENT ON COLUMN users.favorite_anime IS 'User favorite anime description (max 200 characters)';
