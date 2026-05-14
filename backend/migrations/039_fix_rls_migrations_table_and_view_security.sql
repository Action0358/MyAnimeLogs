-- +migrate Up
-- 前提: security_invoker属性はPostgreSQL 15以上で導入。本プロジェクトはローカル(postgres:16)・
--       Supabase(PostgreSQL 15+)いずれもPG15以上のため全環境で動作する。

-- 1. migrationsテーブル（カスタムmigratorが管理するトラッキングテーブル）にRLSを有効化
-- RLSはスーパーユーザー・テーブルオーナー・BYPASSRLS権限保持者には適用されない。
-- 現在のDATABASE_URLはSupabaseのpostgresスーパーユーザーで接続しているためバックエンドへの影響はない。
-- ただし最小権限ユーザーへ移行する際は別途バックエンド用ポリシーの追加が必要。
ALTER TABLE migrations ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'service_role') THEN
    CREATE ROLE service_role;
  END IF;
END $$;

DROP POLICY IF EXISTS "service_role full access" ON migrations;
CREATE POLICY "service_role full access" ON migrations FOR ALL TO service_role USING (true) WITH CHECK (true);

-- RLSポリシーは行フィルタリングのみ制御する。テーブルへのアクセス自体はGRANTで別途許可が必要。
-- ローカル環境等でservice_roleがスーパーユーザーでない場合も正しく動作させるために明示的に付与する。
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE migrations TO service_role;
-- SERIAL型のidカラムにINSERTする際はシーケンスへのUSAGE権限も必要
GRANT USAGE ON SEQUENCE migrations_id_seq TO service_role;

-- 2. animes_with_computed_statusビューをsecurity_invoker=trueに変更
-- SECURITY DEFINERだとビュー作成者（postgres）のRLSポリシーが適用されSupabaseのアドバイザー警告が発生する
-- security_invoker=trueにすることでクエリ実行ユーザーの権限・RLSポリシーが適用される
ALTER VIEW animes_with_computed_status SET (security_invoker = true);
GRANT SELECT ON animes_with_computed_status TO service_role;

-- +migrate Down
REVOKE SELECT ON animes_with_computed_status FROM service_role;
ALTER VIEW animes_with_computed_status RESET (security_invoker);

REVOKE USAGE ON SEQUENCE migrations_id_seq FROM service_role;
REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLE migrations FROM service_role;
DROP POLICY IF EXISTS "service_role full access" ON migrations;
ALTER TABLE migrations DISABLE ROW LEVEL SECURITY;
