-- Add special episode flags to episodes table
-- This migration adds columns to distinguish regular episodes from specials (OVA, recap, filler)

-- Add new columns
ALTER TABLE episodes 
ADD COLUMN IF NOT EXISTS is_special BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS filler BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS recap BOOLEAN DEFAULT FALSE;

-- Drop existing unique constraint
ALTER TABLE episodes DROP CONSTRAINT IF EXISTS unique_anime_episode;

-- Add new unique constraint that includes is_special
-- This allows same episode_number for regular and special episodes
ALTER TABLE episodes ADD CONSTRAINT unique_anime_episode_type UNIQUE(anime_id, episode_number, is_special);

-- Create index for filtering by episode type
CREATE INDEX IF NOT EXISTS idx_episodes_is_special ON episodes(anime_id, is_special);
CREATE INDEX IF NOT EXISTS idx_episodes_filler ON episodes(anime_id, filler) WHERE filler = TRUE;
CREATE INDEX IF NOT EXISTS idx_episodes_recap ON episodes(anime_id, recap) WHERE recap = TRUE;
