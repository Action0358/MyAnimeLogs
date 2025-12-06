-- Update users table text fields to VARCHAR with appropriate length limits
-- This ensures data integrity at the database level

-- Step 1: Check for data that exceeds the new limits and log warnings
DO $$
DECLARE
    long_description_count INTEGER;
    long_favorite_anime_count INTEGER;
BEGIN
    -- Efficiently count both fields in a single table scan using FILTER
    SELECT 
        COUNT(*) FILTER (WHERE LENGTH(description) > 500),
        COUNT(*) FILTER (WHERE LENGTH(favorite_anime) > 200)
    INTO 
        long_description_count,
        long_favorite_anime_count
    FROM users;
    
    -- Log warnings if data will be truncated
    IF long_description_count > 0 THEN
        RAISE WARNING 'Found % user(s) with description longer than 500 characters. Data will be truncated.', long_description_count;
    END IF;
    
    IF long_favorite_anime_count > 0 THEN
        RAISE WARNING 'Found % user(s) with favorite_anime longer than 200 characters. Data will be truncated.', long_favorite_anime_count;
    END IF;
END $$;

-- Step 2: Truncate data that exceeds the new limits
-- This prevents migration failure but may result in data loss
UPDATE users
SET description = LEFT(description, 500)
WHERE LENGTH(description) > 500;

UPDATE users
SET favorite_anime = LEFT(favorite_anime, 200)
WHERE LENGTH(favorite_anime) > 200;

-- Step 3: Update column types to VARCHAR with length constraints
ALTER TABLE users ALTER COLUMN description TYPE VARCHAR(500);
ALTER TABLE users ALTER COLUMN favorite_anime TYPE VARCHAR(200);

-- Step 4: Add comments to document the constraints
COMMENT ON COLUMN users.description IS 'User profile description (max 500 characters)';
COMMENT ON COLUMN users.favorite_anime IS 'User favorite anime description (max 200 characters)';
