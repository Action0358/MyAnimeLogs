-- +migrate Up
-- 1. migrationsテーブル（sql-migrateが管理するトラッキングテーブル）にRLSを有効化
-- GoバックエンドはDATABASE_URL（直接PostgreSQL接続）を使用しているためRLSの影響を受けない
ALTER TABLE migrations ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'service_role') THEN
    CREATE ROLE service_role;
  END IF;
END $$;

DROP POLICY IF EXISTS "service_role full access" ON migrations;
CREATE POLICY "service_role full access" ON migrations FOR ALL TO service_role USING (true) WITH CHECK (true);

-- 2. animes_with_computed_statusビューをsecurity_invoker=trueに変更
-- SECURITY DEFINERだとビュー作成者（postgres）のRLSポリシーが適用されSupabaseのアドバイザー警告が発生する
-- security_invoker=trueにすることでクエリ実行ユーザーの権限・RLSポリシーが適用される
ALTER VIEW animes_with_computed_status SET (security_invoker = true);

-- +migrate Down
ALTER VIEW animes_with_computed_status RESET (security_invoker);

DROP POLICY IF EXISTS "service_role full access" ON migrations;
ALTER TABLE migrations DISABLE ROW LEVEL SECURITY;
