-- Add missing columns that were removed by rollback migrations
-- description: あらすじ（TMDB/AniList APIから取得）
-- tmdb_id: TMDB作品ID
-- series_tmdb_id: シリーズTMDB ID

ALTER TABLE animes ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE animes ADD COLUMN IF NOT EXISTS tmdb_id INTEGER;
ALTER TABLE animes ADD COLUMN IF NOT EXISTS series_tmdb_id INTEGER;

CREATE INDEX IF NOT EXISTS idx_animes_tmdb_id ON animes(tmdb_id);
