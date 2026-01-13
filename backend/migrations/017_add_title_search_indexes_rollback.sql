-- Rollback Migration: 017_add_title_search_indexes.sql
-- Description: title_kana, title_en の検索インデックス削除

DROP INDEX IF EXISTS idx_animes_title_kana_trgm;
DROP INDEX IF EXISTS idx_animes_title_en_trgm;
DROP INDEX IF EXISTS idx_animes_all_titles_tsvector;

-- 削除した idx_animes_title_tsvector を復元
CREATE INDEX IF NOT EXISTS idx_animes_title_tsvector 
ON animes USING GIN (to_tsvector('simple', COALESCE(title, '')));

-- 統計情報を更新
ANALYZE animes;
