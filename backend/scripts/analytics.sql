-- 利用状況の集計クエリ（#868）
--
-- 目的:
--   新規登録数・利用率・継続・検索傾向を、実データから把握する。
--   アクセス解析（Cloudflare Web Analytics）では取れない「実数」をここで見る。
--
-- 使い方:
--   本番   psql "$PRODUCTION_DATABASE_URL" -f backend/scripts/analytics.sql
--   ローカル docker exec -i supabase_db_myanimelogs-supabase \
--              psql -U postgres -d postgres -f /dev/stdin < backend/scripts/analytics.sql
--
--   全部流すと結果が多いので、普段は必要な節だけコピーして実行するとよい。
--
-- 注意:
--   このファイルは参照専用。UPDATE / DELETE は書かないこと。
--
-- 「登録率」について:
--   本サービスはログインしなくても使えるため、「来たが登録しなかった人」は
--   DBに痕跡を残さない。登録率を出すには分母（訪問者数）が要る。
--   Cloudflare Web Analytics の訪問者数 ÷ 下記 1) の新規登録数 で算出する。

\echo '=== 1) 月別の新規登録数 ==='
-- ベータ公開の告知が効いたかを見る基本の数字。
SELECT
    to_char(date_trunc('month', created_at), 'YYYY-MM') AS month,
    count(*)                                            AS signups
FROM users
GROUP BY 1
ORDER BY 1 DESC;

\echo ''
\echo '=== 2) 利用率（登録者のうち、記録機能を実際に使った割合）==='
-- 登録しただけで何も記録していない人が多いなら、
-- 問題は集客ではなく「登録直後の体験」にある。
WITH activated AS (
    SELECT user_id FROM user_favorites
    UNION
    SELECT user_id FROM user_watch_status
)
SELECT
    count(*)                                                   AS users,
    count(*) FILTER (WHERE u.id IN (SELECT user_id FROM activated)) AS activated,
    round(
        100.0 * count(*) FILTER (WHERE u.id IN (SELECT user_id FROM activated))
        / nullif(count(*), 0),
        1
    )                                                          AS rate_pct
FROM users u;

\echo ''
\echo '=== 3) 登録月ごとの定着（コホート）==='
-- 2) を月別に割ったもの。改善の効果が出ているかは、
-- 全期間の平均ではなく「あとから登録した人ほど定着しているか」で見る。
WITH activated AS (
    SELECT user_id FROM user_favorites
    UNION
    SELECT user_id FROM user_watch_status
)
SELECT
    to_char(date_trunc('month', u.created_at), 'YYYY-MM')            AS cohort,
    count(*)                                                         AS signups,
    count(*) FILTER (WHERE u.id IN (SELECT user_id FROM activated))  AS activated,
    round(
        100.0 * count(*) FILTER (WHERE u.id IN (SELECT user_id FROM activated))
        / nullif(count(*), 0),
        1
    )                                                                AS rate_pct
FROM users u
GROUP BY 1
ORDER BY 1 DESC;

\echo ''
\echo '=== 4) 直近30日に記録操作をしたユーザー数 ==='
-- 継続して使われているか。登録数が伸びてもここが増えないなら定着していない。
-- updated_at を見るのは、既存の記録を «変更» した場合も利用とみなすため。
SELECT count(DISTINCT user_id) AS active_30d
FROM (
    SELECT user_id, updated_at FROM user_favorites
    UNION ALL
    SELECT user_id, updated_at FROM user_watch_status
) x
WHERE updated_at >= now() - interval '30 days';

\echo ''
\echo '=== 5) 記録の内訳（お気に入り / 視聴ステータス）==='
-- どちらの機能が使われているか。片方に偏るなら、
-- もう片方は導線か価値の伝わり方に課題がある。
SELECT 'favorite' AS kind, count(*) AS records, count(DISTINCT user_id) AS users
FROM user_favorites
UNION ALL
SELECT 'watch_status', count(*), count(DISTINCT user_id)
FROM user_watch_status
ORDER BY 1;

\echo ''
\echo '=== 6) 視聴ステータスの分布 ==='
SELECT status, count(*) AS records, count(DISTINCT user_id) AS users
FROM user_watch_status
GROUP BY 1
ORDER BY records DESC;

\echo ''
\echo '=== 7) 検索キーワード 上位30（直近30日）==='
-- search_logs は未ログインの検索も記録している（user_id が NULL）。
-- 登録前の関心が見える数少ない手がかり。
SELECT query, count(*) AS n
FROM search_logs
WHERE query <> ''
  AND searched_at >= now() - interval '30 days'
GROUP BY 1
ORDER BY n DESC, 1
LIMIT 30;

\echo ''
\echo '=== 8) 0件だった検索（直近90日）==='
-- 探したのに見つからなかったもの。作品データの欠落や
-- 表記ゆれ（検索の当たらなさ）を直接示すため、改善ネタとして価値が高い。
SELECT query, count(*) AS n
FROM search_logs
WHERE result_count = 0
  AND query <> ''
  AND searched_at >= now() - interval '90 days'
GROUP BY 1
ORDER BY n DESC, 1
LIMIT 30;

\echo ''
\echo '=== 9) 検索のログイン別内訳（直近30日）==='
-- 未ログインの検索がどれだけあるか。ここが多いなら、
-- 「使われてはいるが登録されていない」状態を意味する。
SELECT
    count(*)                                  AS searches,
    count(user_id)                            AS by_logged_in,
    count(*) - count(user_id)                 AS by_anonymous
FROM search_logs
WHERE searched_at >= now() - interval '30 days';
