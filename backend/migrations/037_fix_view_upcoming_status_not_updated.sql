-- +migrate Up
-- Fix: status='upcoming'の作品がreleased_on過ぎても一覧ページで「放送予定」のまま更新されない問題
-- 原因: CASE条件で status = 'not_yet_aired' のみチェックしており、status = 'upcoming' が漏れていた
-- 修正: status IN ('not_yet_aired', 'upcoming') に変更
DROP VIEW IF EXISTS animes_with_computed_status;

CREATE VIEW animes_with_computed_status AS
SELECT 
    id,
    title,
    image_url,
    season,
    CASE 
        -- 1. released_onがYYYY-MM-DD形式の場合: 日付ベースで判定
        WHEN released_on ~ '^\d{4}-\d{2}-\d{2}$' THEN
            CASE
                WHEN released_on::date > CURRENT_DATE THEN 'upcoming'
                WHEN released_on::date <= CURRENT_DATE AND status IN ('not_yet_aired', 'upcoming') THEN 'airing'
                ELSE COALESCE(status, 'airing')
            END
        -- 2. released_onがNULLまたは不正形式で、seasonがある場合: シーズンベースで判定
        WHEN season ~ '^\d{4}-(winter|spring|summer|autumn)$' THEN
            CASE
                WHEN CAST(SUBSTRING(season FROM '^\d{4}') AS int) > EXTRACT(YEAR FROM CURRENT_DATE) THEN 'upcoming'
                WHEN CAST(SUBSTRING(season FROM '^\d{4}') AS int) < EXTRACT(YEAR FROM CURRENT_DATE) THEN
                    COALESCE(NULLIF(status, 'not_yet_aired'), 'finished')
                WHEN season ~ '-winter$' THEN
                    CASE
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) <= 3 THEN COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'airing')
                        ELSE COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'finished')
                    END
                WHEN season ~ '-spring$' THEN
                    CASE
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) < 4 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) <= 6 THEN COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'airing')
                        ELSE COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'finished')
                    END
                WHEN season ~ '-summer$' THEN
                    CASE
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) < 7 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) <= 9 THEN COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'airing')
                        ELSE COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'finished')
                    END
                WHEN season ~ '-autumn$' THEN
                    CASE
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) < 10 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) <= 12 THEN COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'airing')
                        ELSE COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'finished')
                    END
                ELSE COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'finished')
            END
        -- 3. released_onもseasonもない場合: statusをそのまま返す
        ELSE COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'finished')
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
    cover_image_updated_at,
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

COMMENT ON VIEW animes_with_computed_status IS 'アニメテーブルのVIEW。released_onおよびseasonを考慮してステータスを動的に計算する。status=upcoming/not_yet_airedの両方を補正対象とする。cover_image_updated_atを含みTTL判定をアプリ側で行う。';

-- +migrate Down
-- 036の状態に戻す
DROP VIEW IF EXISTS animes_with_computed_status;

CREATE VIEW animes_with_computed_status AS
SELECT 
    id,
    title,
    image_url,
    season,
    CASE 
        WHEN released_on ~ '^\d{4}-\d{2}-\d{2}$' THEN
            CASE
                WHEN released_on::date > CURRENT_DATE THEN 'upcoming'
                WHEN released_on::date <= CURRENT_DATE AND status = 'not_yet_aired' THEN 'airing'
                ELSE COALESCE(status, 'airing')
            END
        WHEN season ~ '^\d{4}-(winter|spring|summer|autumn)$' THEN
            CASE
                WHEN CAST(SUBSTRING(season FROM '^\d{4}') AS int) > EXTRACT(YEAR FROM CURRENT_DATE) THEN 'upcoming'
                WHEN CAST(SUBSTRING(season FROM '^\d{4}') AS int) < EXTRACT(YEAR FROM CURRENT_DATE) THEN
                    COALESCE(NULLIF(status, 'not_yet_aired'), 'finished')
                WHEN season ~ '-winter$' THEN
                    CASE
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) <= 3 THEN COALESCE(NULLIF(status, 'not_yet_aired'), 'airing')
                        ELSE COALESCE(NULLIF(status, 'not_yet_aired'), 'finished')
                    END
                WHEN season ~ '-spring$' THEN
                    CASE
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) < 4 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) <= 6 THEN COALESCE(NULLIF(status, 'not_yet_aired'), 'airing')
                        ELSE COALESCE(NULLIF(status, 'not_yet_aired'), 'finished')
                    END
                WHEN season ~ '-summer$' THEN
                    CASE
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) < 7 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) <= 9 THEN COALESCE(NULLIF(status, 'not_yet_aired'), 'airing')
                        ELSE COALESCE(NULLIF(status, 'not_yet_aired'), 'finished')
                    END
                WHEN season ~ '-autumn$' THEN
                    CASE
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) < 10 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) <= 12 THEN COALESCE(NULLIF(status, 'not_yet_aired'), 'airing')
                        ELSE COALESCE(NULLIF(status, 'not_yet_aired'), 'finished')
                    END
                ELSE COALESCE(NULLIF(status, 'not_yet_aired'), 'finished')
            END
        ELSE COALESCE(NULLIF(status, 'not_yet_aired'), 'finished')
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
    cover_image_updated_at,
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

COMMENT ON VIEW animes_with_computed_status IS 'アニメテーブルのVIEW。released_onおよびseasonを考慮してステータスを動的に計算する。cover_image_updated_atを含みTTL判定をアプリ側で行う。';
