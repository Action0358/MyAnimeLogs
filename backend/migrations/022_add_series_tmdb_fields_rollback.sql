-- Rollback: シリーズ作品のTMDB ID共有機能

DROP INDEX IF EXISTS idx_animes_series_tmdb_id;
DROP INDEX IF EXISTS idx_animes_series_title;

ALTER TABLE animes DROP COLUMN IF EXISTS series_tmdb_id;
ALTER TABLE animes DROP COLUMN IF EXISTS series_title;
