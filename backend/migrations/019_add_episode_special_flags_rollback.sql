-- Rollback: Remove special episode flags from episodes table

-- Drop indexes
DROP INDEX IF EXISTS idx_episodes_is_special;
DROP INDEX IF EXISTS idx_episodes_filler;
DROP INDEX IF EXISTS idx_episodes_recap;

-- Drop new unique constraint
ALTER TABLE episodes DROP CONSTRAINT IF EXISTS unique_anime_episode_type;

-- Restore original unique constraint
ALTER TABLE episodes ADD CONSTRAINT unique_anime_episode UNIQUE(anime_id, episode_number);

-- Remove columns
ALTER TABLE episodes 
DROP COLUMN IF EXISTS is_special,
DROP COLUMN IF EXISTS filler,
DROP COLUMN IF EXISTS recap;
