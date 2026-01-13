-- Migration: 017_add_title_search_indexes.sql
-- Description: title_kana, title_en の検索インデックス追加
-- Issue: 検索機能がひらがな・カタカナ・ローマ字で正しくヒットしない問題の修正

-- title_kana のトライグラムインデックス（ILIKE検索用）
CREATE INDEX IF NOT EXISTS idx_animes_title_kana_trgm 
ON animes USING GIN (title_kana gin_trgm_ops);

-- title_en のトライグラムインデックス（ILIKE検索用）
CREATE INDEX IF NOT EXISTS idx_animes_title_en_trgm 
ON animes USING GIN (title_en gin_trgm_ops);

-- 複合全文検索インデックス（FTS用）
-- title, title_kana, title_en を結合した全文検索インデックス
CREATE INDEX IF NOT EXISTS idx_animes_all_titles_tsvector 
ON animes USING GIN (to_tsvector('simple', COALESCE(title, '') || ' ' || COALESCE(title_kana, '') || ' ' || COALESCE(title_en, '')));

-- 統計情報を更新
ANALYZE animes;
