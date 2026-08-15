-- Google 認証専用への移行に伴う影響調査スクリプト（#738）
--
-- 目的: 認証手段を Google ソーシャルログインのみに切り替えたとき、
--       ログインできなくなる可能性のある既存ユーザーを事前に洗い出す。
--
-- 実行方法: Supabase ダッシュボードの SQL Editor（本番プロジェクト）で実行する。
--           参照のみで、データを変更するクエリは含まない。
--
-- 【重要な前提】
-- メールアドレスのドメインだけでは Google アカウントの有無を確定できない。
--   - Gmail 以外のドメインでも Google Workspace なら Google アカウントである
--   - 任意のメールアドレス（Yahoo 等）で Google アカウントを作ることもできる
-- したがって本スクリプトの結果は「確定リスト」ではなく「要確認候補」である。
-- 最終判断は対象ユーザーへの個別確認が必要。
--
-- 【移行時に何が起きるか】2026-08 の本番・STG 移行で実測した挙動
-- ユーザーが Google でログインすると、以下のいずれかの経路を通る。
-- いずれの経路でも既存の user_id は維持され、視聴記録・お気に入りは引き継がれる。
--
--   existing … Supabase が「検証済みメールが一致する identity」を既存 auth ユーザーへ
--              自動リンクするため、auth_user_id が変わらないケース。STG がこれだった
--   claimed  … public.users.auth_user_id が NULL の行を、メール一致で紐づけるケース。
--              Supabase Auth 移行（migration 041）以前から存在する行が該当する。
--              auth 側に対応 identity がないため自動リンクでは解決せず、
--              アプリ側（GetOrCreateByAuthUserID）の処理が必要。本番がこれだった
--   relinked … 別の auth_user_id が設定済みの行を張り替えるケース
--   created  … 新規ユーザー。既存データがある場合にこれを通ると引き継ぎ失敗を意味する
--
-- 実際にどの経路を通ったかは、バックエンドのログで確認できる。
--   flyctl logs -a myanimelogs-production --no-tail | grep "linked users row"
-- ※ existing はログを出力しない（通常ログインと区別できないため）

-- ============================================================
-- 1. 全体サマリ: 移行対象の母数を把握する
-- ============================================================
--
-- 【重要】移行の進捗は `provider`（単数）ではなく `providers`（配列）で判定する。
--   - provider  … primary provider。既存ユーザーが Google へ移行しても 'email' のまま変わらない
--   - providers … リンク済み identity の配列。Google をリンクすると 'google' が追加される
-- provider 列だけを見ると「誰も Google を使っていない」と誤読するため、両方を出す。
SELECT
    COUNT(*)                                                    AS auth_users_total,
    COUNT(*) FILTER (WHERE email_confirmed_at IS NOT NULL)      AS email_confirmed,
    COUNT(*) FILTER (WHERE email_confirmed_at IS NULL)          AS email_unconfirmed,
    -- 移行状況の判定はこの2列で行う
    COUNT(*) FILTER (WHERE raw_app_meta_data->'providers' ? 'google')       AS linked_google,
    COUNT(*) FILTER (WHERE NOT (raw_app_meta_data->'providers' ? 'google')) AS not_linked_google,
    -- 参考値（primary provider。移行しても変わらないため進捗判定には使えない）
    COUNT(*) FILTER (WHERE raw_app_meta_data->>'provider' = 'email')  AS primary_provider_email,
    COUNT(*) FILTER (WHERE raw_app_meta_data->>'provider' = 'google') AS primary_provider_google
FROM auth.users;

-- ============================================================
-- 2. メールドメイン別の内訳
--    gmail.com / googlemail.com 以外が「要確認候補」になる
-- ============================================================
SELECT
    LOWER(SPLIT_PART(email, '@', 2)) AS domain,
    COUNT(*)                         AS user_count,
    CASE
        WHEN LOWER(SPLIT_PART(email, '@', 2)) IN ('gmail.com', 'googlemail.com')
            THEN 'Google アカウント確実'
        ELSE '要確認（Workspace の可能性あり）'
    END AS assessment
FROM auth.users
WHERE email IS NOT NULL
GROUP BY 1, 3
ORDER BY user_count DESC;

-- ============================================================
-- 3. 要確認ユーザーの一覧（保有データ量つき）
--
--    データを多く持っているユーザーほど、ログイン不能になったときの
--    影響が大きい。優先的に個別連絡すべき対象を上から並べる。
-- ============================================================
SELECT
    u.id                                    AS user_id,
    u.name,
    u.email,
    a.raw_app_meta_data->>'provider'        AS provider,
    a.email_confirmed_at IS NOT NULL        AS email_confirmed,
    a.last_sign_in_at,
    COALESCE(f.cnt, 0)                      AS favorites,
    COALESCE(w.cnt, 0)                      AS watch_records,
    COALESCE(f.cnt, 0) + COALESCE(w.cnt, 0) AS total_records
FROM users u
JOIN auth.users a ON a.id = u.auth_user_id
LEFT JOIN (
    SELECT user_id, COUNT(*) AS cnt FROM user_favorites GROUP BY user_id
) f ON f.user_id = u.id
LEFT JOIN (
    SELECT user_id, COUNT(*) AS cnt FROM user_watch_status GROUP BY user_id
) w ON w.user_id = u.id
WHERE LOWER(SPLIT_PART(u.email, '@', 2)) NOT IN ('gmail.com', 'googlemail.com')
ORDER BY total_records DESC, u.id;

-- ============================================================
-- 4. メール未確認のユーザー
--
--    Supabase は「検証済みメールが一致する identity」を自動リンクする。
--    メール未確認のアカウントは自動リンクが働かない可能性があり、
--    アプリ側の張り替え（GetOrCreateByAuthUserID の relinked 経路）に
--    依存することになるため、別枠で把握しておく。
-- ============================================================
-- public.users 行が未作成のアカウントも取りこぼさないよう auth.users を起点にする
SELECT
    a.id            AS auth_user_id,
    a.email,
    u.id            AS user_id,   -- NULL なら public.users 行がまだ無い
    u.name,
    a.created_at    AS auth_created_at
FROM auth.users a
LEFT JOIN users u ON u.auth_user_id = a.id
WHERE a.email_confirmed_at IS NULL
ORDER BY a.created_at;

-- ============================================================
-- 5. 整合性チェック: 孤立行の検出
--
--    5-1. auth_user_id が未設定の users 行
--         → 初回ログイン時にメール一致で claimed 経路が走る想定
--
--    【注意】データ保有量が0でも「放置してよい行」とは限らない。
--    2026-08 の本番移行では、データ0件だったこの行が実際のログイン先だった。
--    保有量ではなく、下の 5-4 でメール一致の有無を確認して判断すること。
-- ============================================================
SELECT id AS user_id, name, email, created_at
FROM users
WHERE auth_user_id IS NULL
ORDER BY created_at;

--    5-4. auth_user_id が未設定の行が、auth.users のメールと一致するか
--         matched   → その人がログインした時点で claimed 経路により紐づく（対応不要）
--         no_match  → auth 側に対応が無い完全な孤立行。移行前のテストデータの可能性
SELECT
    (SELECT COUNT(*) FROM users u
     WHERE u.auth_user_id IS NULL
       AND EXISTS (SELECT 1 FROM auth.users a
                   WHERE LOWER(a.email) = LOWER(u.email))) AS orphan_matched_by_email,
    (SELECT COUNT(*) FROM users u
     WHERE u.auth_user_id IS NULL
       AND NOT EXISTS (SELECT 1 FROM auth.users a
                       WHERE LOWER(a.email) = LOWER(u.email))) AS orphan_no_match;

--    5-2. public.users に対応行が無い auth.users
--         → 一度もバックエンド API を叩いていないアカウント
SELECT a.id AS auth_user_id, a.email, a.created_at, a.last_sign_in_at
FROM auth.users a
LEFT JOIN users u ON u.auth_user_id = a.id
WHERE u.id IS NULL
ORDER BY a.created_at;

--    5-3. 同一メールが auth.users に重複していないか
--         → public.users.email は UNIQUE のため、重複があると
--           張り替え時にどちらへ紐づけるかの判断が必要になる
SELECT LOWER(email) AS email, COUNT(*) AS cnt
FROM auth.users
WHERE email IS NOT NULL
GROUP BY 1
HAVING COUNT(*) > 1
ORDER BY cnt DESC;
