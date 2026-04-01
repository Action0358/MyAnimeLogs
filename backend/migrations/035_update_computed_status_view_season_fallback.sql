-- +migrate Up
-- VIEWを更新: released_onがNULLの場合にシーズン情報からステータスを動的に計算
-- これにより、released_onが未取得の今期作品でも正しい放送状態ラベルが表示される
-- not_yet_airedはairingとして扱う（NULLIF + COALESCEパターン）
-- 注意: COALESCE/NULLIFの結果がtext型になるため、CREATE OR REPLACEではvarchar→textの型変更エラーが発生する
-- そのため一度DROPしてから再作成する
DROP VIEW IF EXISTS animes_with_computed_status;

CREATE VIEW animes_with_computed_status AS
SELECT 
    id,
    title,
    image_url,
    season,
    -- ステータスを動的に計算
    CASE 
        -- 1. released_onがYYYY-MM-DD形式の場合: 日付ベースで判定
        WHEN released_on ~ '^\d{4}-\d{2}-\d{2}$' THEN
            CASE
                WHEN released_on::date > CURRENT_DATE THEN 'upcoming'
                WHEN released_on::date <= CURRENT_DATE AND status = 'not_yet_aired' THEN 'airing'
                ELSE COALESCE(status, 'airing')
            END
        -- 2. released_onがNULLまたは不正形式で、seasonがある場合: シーズンベースで判定
        WHEN season ~ '^\d{4}-(winter|spring|summer|autumn)$' THEN
            CASE
                -- シーズン年を抽出
                WHEN CAST(SUBSTRING(season FROM '^\d{4}') AS int) > EXTRACT(YEAR FROM CURRENT_DATE) THEN 'upcoming'
                WHEN CAST(SUBSTRING(season FROM '^\d{4}') AS int) < EXTRACT(YEAR FROM CURRENT_DATE) THEN
                    COALESCE(NULLIF(status, 'not_yet_aired'), 'finished')
                -- 同じ年の場合: シーズンの月範囲で判定
                -- 冬シーズン（1-3月）: 同じ年なら必ずシーズン期間中（年を跨ぐ場合は前段の年比較で判定済み）
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
        -- 3. released_onもseasonもない場合: statusをそのまま返す
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

COMMENT ON VIEW animes_with_computed_status IS 'アニメテーブルのVIEW。released_onおよびseasonを考慮してステータスを動的に計算する。released_onがある場合は日付ベース、ない場合はシーズンベースでフォールバック判定を行う。not_yet_airedはairingとして扱う。';

-- +migrate Down
-- 元のVIEW定義に戻す（DROP + CREATE で型の不整合を回避）
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

COMMENT ON VIEW animes_with_computed_status IS 'アニメテーブルのVIEW。released_onを考慮してステータスを動的に計算する。正規表現でYYYY-MM-DD形式を検証してから日付キャストを行うため、不正な日付文字列によるエラーを防ぐ。';
