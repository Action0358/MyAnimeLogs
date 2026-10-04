-- +migrate Up
-- 劇場版作品でreleased_on（公開日）が未取得のとき、Annict由来のstatusが'airing'だと
-- 「放送中」と誤表示される問題を修正する。
--
-- 例: 「劇場版 薬屋のひとりごと 亡妃の秘宝」はreleased_onが空、season='2026-autumn'、
-- original_status='airing'のため、季節ベースのフォールバックで「放送中」と計算されていた。
-- 公開日が不明な映画は、まだ公開されていない可能性が極めて高い（実データ調査で
-- released_on空・media='movie'のうち873件はstatus='finished'として正しく同期されており、
-- status='airing'の6件のみが本来は「公開前」であるべきケースだった）。
--
-- そのため条件は media = 'movie' だけでなく status = 'airing' も必須にする。
-- status='finished'の映画（released_on未同期なだけで公開済み）を誤って
-- 「放送予定」に変えてしまわないようにするため。
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
        -- 2. 映画でreleased_onが無く、かつ元ステータスが'airing'（Annict側で公開日未確定なまま
        --    'airing'になっているケース）の場合: 公開前とみなす
        WHEN media = 'movie' AND status = 'airing' THEN 'upcoming'
        -- 3. released_onがNULLまたは不正形式で、seasonがある場合: シーズンベースで判定
        WHEN season ~ '^\d{4}-(winter|spring|summer|autumn)$' THEN
            CASE
                WHEN CAST(SUBSTRING(season FROM '^\d{4}') AS int) > EXTRACT(YEAR FROM CURRENT_DATE) THEN 'upcoming'
                WHEN CAST(SUBSTRING(season FROM '^\d{4}') AS int) < EXTRACT(YEAR FROM CURRENT_DATE) THEN
                    COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'finished')
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
        -- 4. released_onもseasonもない場合: statusをそのまま返す
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

ALTER VIEW animes_with_computed_status SET (security_invoker = true);

COMMENT ON VIEW animes_with_computed_status IS 'アニメテーブルのVIEW。released_onおよびseasonを考慮してステータスを動的に計算する。status=upcoming/not_yet_airedの両方を補正対象とする。released_on未確定の映画でstatus=airingの場合はupcomingとして扱う。cover_image_updated_atを含みTTL判定をアプリ側で行う。';

-- +migrate Down
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
                WHEN released_on::date <= CURRENT_DATE AND status IN ('not_yet_aired', 'upcoming') THEN 'airing'
                ELSE COALESCE(status, 'airing')
            END
        WHEN season ~ '^\d{4}-(winter|spring|summer|autumn)$' THEN
            CASE
                WHEN CAST(SUBSTRING(season FROM '^\d{4}') AS int) > EXTRACT(YEAR FROM CURRENT_DATE) THEN 'upcoming'
                WHEN CAST(SUBSTRING(season FROM '^\d{4}') AS int) < EXTRACT(YEAR FROM CURRENT_DATE) THEN
                    COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'finished')
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

ALTER VIEW animes_with_computed_status SET (security_invoker = true);

COMMENT ON VIEW animes_with_computed_status IS 'アニメテーブルのVIEW。released_onおよびseasonを考慮してステータスを動的に計算する。status=upcoming/not_yet_airedの両方を補正対象とする。cover_image_updated_atを含みTTL判定をアプリ側で行う。';
