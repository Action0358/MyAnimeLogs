-- Restore episode special flags that were removed by rollback migration
-- is_special: 特別エピソード（OVA等）フラグ
-- filler: フィラーエピソードフラグ
-- recap: 総集編エピソードフラグ

ALTER TABLE episodes ADD COLUMN IF NOT EXISTS is_special BOOLEAN DEFAULT FALSE;
ALTER TABLE episodes ADD COLUMN IF NOT EXISTS filler BOOLEAN DEFAULT FALSE;
ALTER TABLE episodes ADD COLUMN IF NOT EXISTS recap BOOLEAN DEFAULT FALSE;

-- Restore unique constraint (drop old one first if exists, then add new one)
ALTER TABLE episodes DROP CONSTRAINT IF EXISTS unique_anime_episode;
ALTER TABLE episodes DROP CONSTRAINT IF EXISTS unique_anime_episode_type;
ALTER TABLE episodes ADD CONSTRAINT unique_anime_episode_type UNIQUE(anime_id, episode_number, is_special);

-- Restore indexes
CREATE INDEX IF NOT EXISTS idx_episodes_is_special ON episodes(anime_id, is_special);
CREATE INDEX IF NOT EXISTS idx_episodes_filler ON episodes(anime_id, filler) WHERE filler = TRUE;
CREATE INDEX IF NOT EXISTS idx_episodes_recap ON episodes(anime_id, recap) WHERE recap = TRUE;
