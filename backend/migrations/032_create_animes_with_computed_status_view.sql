-- +migrate Up
-- アニメのステータスを動的に計算するVIEWを作成
-- released_onが未来の日付の場合は'upcoming'、過去の日付でnot_yet_airedの場合は'airing'に変更
CREATE OR REPLACE VIEW animes_with_computed_status AS
SELECT 
    id,
    title,
    image_url,
    season,
    -- ステータスを動的に計算（正規表現でYYYY-MM-DD形式を検証してからキャスト）
    CASE 
        WHEN released_on ~ '^\d{4}-\d{2}-\d{2}$' THEN
            CASE
                WHEN released_on::date > CURRENT_DATE THEN 'upcoming'
                WHEN released_on::date <= CURRENT_DATE AND status = 'not_yet_aired' THEN 'airing'
                ELSE status
            END
        ELSE status
    END as computed_status,
    status as original_status,
    mal_anime_id,
    annict_id,
    created_at,
    updated_at,
    watchers_count,
    COALESCE(cover_image_extra_large, '') as cover_image_extra_large,
    COALESCE(cover_image_large, '') as cover_image_large,
    COALESCE(cover_image_medium, '') as cover_image_medium,
    released_on,
    description,
    media,
    official_site_url,
    title_kana,
    title_en,
    episodes_count,
    series_title,
    series_tmdb_id,
    broadcast_day,
    broadcast_time,
    broadcast_timezone,
    duration_minutes,
    genres_synced_at,
    last_synced_at,
    annict_updated_at,
    metadata_synced_at,
    sync_status,
    sync_error
FROM animes;

-- VIEWにコメントを追加
COMMENT ON VIEW animes_with_computed_status IS 'アニメテーブルのVIEW。released_onを考慮してステータスを動的に計算する。正規表現でYYYY-MM-DD形式を検証してから日付キャストを行うため、不正な日付文字列によるエラーを防ぐ。';

-- +migrate Down
DROP VIEW IF EXISTS animes_with_computed_status;
