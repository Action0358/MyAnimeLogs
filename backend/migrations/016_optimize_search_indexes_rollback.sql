-- Rollback Migration: 016_optimize_search_indexes_rollback.sql
-- Description: 検索パフォーマンス最適化インデックスの削除

-- インデックスを削除
DROP INDEX IF EXISTS idx_animes_title_tsvector;
DROP INDEX IF EXISTS idx_animes_title_trgm;
DROP INDEX IF EXISTS idx_animes_description_trgm;
DROP INDEX IF EXISTS idx_animes_season_watchers;
DROP INDEX IF EXISTS idx_animes_watchers_count;
DROP INDEX IF EXISTS idx_user_favorites_user_anime;
DROP INDEX IF EXISTS idx_watch_statuses_user_episode;
DROP INDEX IF EXISTS idx_episodes_anime_id;
DROP INDEX IF EXISTS idx_anime_genres_genre;
DROP INDEX IF EXISTS idx_user_ratings_anime_id;
DROP INDEX IF EXISTS idx_animes_mal_anime_id_cover;

-- 注意: pg_trgm拡張は他の機能で使用されている可能性があるため、
-- 削除しない（必要に応じて手動で削除）
-- DROP EXTENSION IF EXISTS pg_trgm;
