-- Drop watch_statuses table (episode watch tracking)
DROP TABLE IF EXISTS watch_statuses;

-- Drop episodes table
DROP TABLE IF EXISTS episodes;

-- Remove episode-related columns from animes table
ALTER TABLE animes DROP COLUMN IF EXISTS episodes_count;
ALTER TABLE animes DROP COLUMN IF EXISTS episodes_synced_at;
