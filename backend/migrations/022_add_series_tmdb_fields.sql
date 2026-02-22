-- +migrate Up
-- シリーズ作品のTMDB ID共有機能
-- series_title: タイトル正規化後のシリーズ名（内部識別用）
-- series_tmdb_id: シリーズ代表のTMDB ID（あらすじ・配信サービス取得用）

ALTER TABLE animes ADD COLUMN IF NOT EXISTS series_title VARCHAR(255);
ALTER TABLE animes ADD COLUMN IF NOT EXISTS series_tmdb_id INTEGER;

-- インデックス作成
CREATE INDEX IF NOT EXISTS idx_animes_series_title ON animes(series_title);
CREATE INDEX IF NOT EXISTS idx_animes_series_tmdb_id ON animes(series_tmdb_id);

-- 既存データのseries_titleを更新（タイトルから正規化）
-- 注: 正規化ロジックはGoコードで実行するため、ここでは空のまま

-- +migrate Down
DROP INDEX IF EXISTS idx_animes_series_tmdb_id;
DROP INDEX IF EXISTS idx_animes_series_title;
ALTER TABLE animes DROP COLUMN IF EXISTS series_tmdb_id;
ALTER TABLE animes DROP COLUMN IF EXISTS series_title;
