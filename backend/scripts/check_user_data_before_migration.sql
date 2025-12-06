-- Pre-migration data check script
-- Run this before executing 012_update_user_text_fields_to_varchar.sql
-- to understand the impact of the migration

-- Check for users with description longer than 500 characters
SELECT 
    id,
    name,
    email,
    LENGTH(description) as description_length,
    LEFT(description, 50) || '...' as description_preview
FROM users
WHERE LENGTH(description) > 500
ORDER BY LENGTH(description) DESC;

-- Check for users with favorite_anime longer than 200 characters
SELECT 
    id,
    name,
    email,
    LENGTH(favorite_anime) as favorite_anime_length,
    LEFT(favorite_anime, 50) || '...' as favorite_anime_preview
FROM users
WHERE LENGTH(favorite_anime) > 200
ORDER BY LENGTH(favorite_anime) DESC;

-- Summary statistics
SELECT 
    'description' as field,
    COUNT(*) as total_users,
    COUNT(CASE WHEN LENGTH(description) > 500 THEN 1 END) as exceeds_limit,
    MAX(LENGTH(description)) as max_length,
    AVG(LENGTH(description))::INTEGER as avg_length
FROM users
WHERE description IS NOT NULL AND description != ''

UNION ALL

SELECT 
    'favorite_anime' as field,
    COUNT(*) as total_users,
    COUNT(CASE WHEN LENGTH(favorite_anime) > 200 THEN 1 END) as exceeds_limit,
    MAX(LENGTH(favorite_anime)) as max_length,
    AVG(LENGTH(favorite_anime))::INTEGER as avg_length
FROM users
WHERE favorite_anime IS NOT NULL AND favorite_anime != '';
