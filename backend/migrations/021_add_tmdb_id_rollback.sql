-- Rollback: Remove TMDB ID column from animes table
DROP INDEX IF EXISTS idx_animes_tmdb_id;
ALTER TABLE animes DROP COLUMN IF EXISTS tmdb_id;
