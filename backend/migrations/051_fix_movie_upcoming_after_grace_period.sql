-- +migrate Up
-- released_on未確定の映画を「放送予定」とする分岐（048）を、元ステータスが
-- 'upcoming' / 'not_yet_aired' の場合にも適用する。
--
-- 048はstatus='airing'の映画だけを対象にしていたが、050の猶予期間により
-- Go側のステータス同期（determineStatusBySeason）はシーズン開始月の1〜14日に
-- 未確定作品のstatusを'upcoming'で保存する。猶予期間が明けた15日以降、次の
-- 日次同期でstatusが'airing'に書き換わるまでの間、映画はこの分岐を通らず
-- シーズンベースの判定で「放送中」になっていた
-- （例: 劇場版 薬屋のひとりごと 亡妃の秘宝、2026-10-15）。
--
-- 公開済みとして同期された映画（status='finished'）は従来どおり対象外。
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
        -- 2. 映画でreleased_onが無く、かつ元ステータスが公開済み（finished）でない場合: 公開前とみなす
        WHEN media = 'movie' AND status IN ('airing', 'upcoming', 'not_yet_aired') THEN 'upcoming'
        -- 3. released_onがNULLまたは不正形式で、seasonがある場合: シーズンベースで判定
        WHEN season ~ '^\d{4}-(winter|spring|summer|autumn)$' THEN
            CASE
                WHEN CAST(SUBSTRING(season FROM '^\d{4}') AS int) > EXTRACT(YEAR FROM CURRENT_DATE) THEN 'upcoming'
                WHEN CAST(SUBSTRING(season FROM '^\d{4}') AS int) < EXTRACT(YEAR FROM CURRENT_DATE) THEN
                    COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'finished')
                WHEN season ~ '-winter$' THEN
                    CASE
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) = 1 AND EXTRACT(DAY FROM CURRENT_DATE) <= 14 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) <= 3 THEN COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'airing')
                        ELSE COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'finished')
                    END
                WHEN season ~ '-spring$' THEN
                    CASE
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) < 4 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) = 4 AND EXTRACT(DAY FROM CURRENT_DATE) <= 14 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) <= 6 THEN COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'airing')
                        ELSE COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'finished')
                    END
                WHEN season ~ '-summer$' THEN
                    CASE
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) < 7 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) = 7 AND EXTRACT(DAY FROM CURRENT_DATE) <= 14 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) <= 9 THEN COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'airing')
                        ELSE COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'finished')
                    END
                WHEN season ~ '-autumn$' THEN
                    CASE
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) < 10 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) = 10 AND EXTRACT(DAY FROM CURRENT_DATE) <= 14 THEN 'upcoming'
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

COMMENT ON VIEW animes_with_computed_status IS 'アニメテーブルのVIEW。released_onおよびseasonを考慮してステータスを動的に計算する。status=upcoming/not_yet_airedの両方を補正対象とする。released_on未確定の映画はstatus=finishedでなければupcomingとして扱う。シーズン開始月の最初の14日間はreleased_on未確定の作品を放送予定として扱う。cover_image_updated_atを含みTTL判定をアプリ側で行う。';

-- +migrate Down
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
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) = 1 AND EXTRACT(DAY FROM CURRENT_DATE) <= 14 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) <= 3 THEN COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'airing')
                        ELSE COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'finished')
                    END
                WHEN season ~ '-spring$' THEN
                    CASE
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) < 4 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) = 4 AND EXTRACT(DAY FROM CURRENT_DATE) <= 14 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) <= 6 THEN COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'airing')
                        ELSE COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'finished')
                    END
                WHEN season ~ '-summer$' THEN
                    CASE
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) < 7 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) = 7 AND EXTRACT(DAY FROM CURRENT_DATE) <= 14 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) <= 9 THEN COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'airing')
                        ELSE COALESCE(NULLIF(NULLIF(status, 'not_yet_aired'), 'upcoming'), 'finished')
                    END
                WHEN season ~ '-autumn$' THEN
                    CASE
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) < 10 THEN 'upcoming'
                        WHEN EXTRACT(MONTH FROM CURRENT_DATE) = 10 AND EXTRACT(DAY FROM CURRENT_DATE) <= 14 THEN 'upcoming'
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

COMMENT ON VIEW animes_with_computed_status IS 'アニメテーブルのVIEW。released_onおよびseasonを考慮してステータスを動的に計算する。status=upcoming/not_yet_airedの両方を補正対象とする。released_on未確定の映画でstatus=airingの場合はupcomingとして扱う。シーズン開始月の最初の14日間はreleased_on未確定の作品を放送予定として扱う。cover_image_updated_atを含みTTL判定をアプリ側で行う。';

