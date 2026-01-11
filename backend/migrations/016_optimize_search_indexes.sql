-- Migration: 016_optimize_search_indexes.sql
-- Description: 検索パフォーマンス最適化のためのインデックス追加
-- Phase 2.2: クエリ最適化

-- pg_trgm拡張を有効化（ILIKE検索の高速化）
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 1. アニメタイトルの全文検索用GINインデックス
-- to_tsvector('simple', title)を事前計算してインデックス化
CREATE INDEX IF NOT EXISTS idx_animes_title_tsvector 
ON animes USING GIN (to_tsvector('simple', COALESCE(title, '')));

-- 2. アニメタイトルのトライグラムインデックス（ILIKE検索用）
CREATE INDEX IF NOT EXISTS idx_animes_title_trgm 
ON animes USING GIN (title gin_trgm_ops);

-- 3. アニメ説明文のトライグラムインデックス（ILIKE検索用）
CREATE INDEX IF NOT EXISTS idx_animes_description_trgm 
ON animes USING GIN (description gin_trgm_ops);

-- 4. シーズン検索用のインデックス（今期アニメ取得の高速化）
CREATE INDEX IF NOT EXISTS idx_animes_season_watchers 
ON animes (season, watchers_count DESC NULLS LAST);

-- 5. 人気アニメ取得用のインデックス
CREATE INDEX IF NOT EXISTS idx_animes_watchers_count 
ON animes (watchers_count DESC) 
WHERE watchers_count > 0;

-- 6. ユーザーお気に入り検索用の複合インデックス
CREATE INDEX IF NOT EXISTS idx_user_favorites_user_anime 
ON user_favorites (user_id, anime_id);

-- 7. 視聴ステータス検索用の複合インデックス
CREATE INDEX IF NOT EXISTS idx_watch_statuses_user_episode 
ON watch_statuses (user_id, episode_id) 
WHERE viewed = true;

-- 8. エピソード検索用のインデックス（視聴進捗計算の高速化）
CREATE INDEX IF NOT EXISTS idx_episodes_anime_id 
ON episodes (anime_id);

-- 9. ジャンル検索用のインデックス
CREATE INDEX IF NOT EXISTS idx_anime_genres_genre 
ON anime_genres (genre);

-- 10. 評価統計用のインデックス
CREATE INDEX IF NOT EXISTS idx_user_ratings_anime_id 
ON user_ratings (anime_id);

-- 11. カバー画像キャッシュ用のインデックス（MAL IDでの検索高速化）
CREATE INDEX IF NOT EXISTS idx_animes_mal_anime_id_cover 
ON animes (mal_anime_id) 
WHERE cover_image_large IS NOT NULL;

-- 統計情報を更新
ANALYZE animes;
ANALYZE user_favorites;
ANALYZE watch_statuses;
ANALYZE episodes;
ANALYZE anime_genres;
ANALYZE user_ratings;
