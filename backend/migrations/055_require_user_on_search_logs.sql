-- +migrate Up
-- 検索ログを「ログイン中の、キーワードありの検索」だけにする。
--
-- 未ログインの検索と、一覧を開いただけの空キーワードの行は今後記録しないため、既存の行も消す。
-- 054 より前の退会者の行は user_id が NULL で未ログインの行と区別できず、
-- プライバシーポリシーの「アカウントに紐付く検索の記録は削除される」に反して残っていたが、
-- ここで一緒に消える。
--
-- デプロイ中は旧バイナリが未ログインの行を書き込むことがある。DELETE と SET NOT NULL の間に
-- 割り込まれるとマイグレーションが失敗するので、先にテーブルをロックする。
-- ロック解除後の旧バイナリの INSERT は NOT NULL で失敗するが、検索ログの失敗は Warn だけで検索には影響しない。
LOCK TABLE search_logs IN ACCESS EXCLUSIVE MODE;

DELETE FROM search_logs
WHERE user_id IS NULL
   OR btrim(query) = '';

ALTER TABLE search_logs ALTER COLUMN user_id SET NOT NULL;

-- +migrate Down
-- 削除した行は元に戻らない。
ALTER TABLE search_logs ALTER COLUMN user_id DROP NOT NULL;
