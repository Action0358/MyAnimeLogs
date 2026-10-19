-- +migrate Up
-- アカウント削除時に検索ログ（search_logs）も削除する。
--
-- これまでは ON DELETE SET NULL で、退会後も検索キーワードが user_id を外した状態で
-- 残っていた。利用規約・プライバシーポリシー・削除確認ダイアログの
-- 「アカウント削除により記録データは削除される」に揃えるため、CASCADE に変更する。
--
-- 未ログイン時の検索（user_id が元から NULL）は対象外で、集計用にそのまま残る。
-- 既に SET NULL で匿名化された過去の退会者の行は未ログイン時の行と区別できないため、
-- そのまま残す。
--
-- 制約名は 008 で自動命名されたもの（search_logs_user_id_fkey）のはずだが、
-- 環境差で名前が違うと DROP を素通りして SET NULL と CASCADE の外部キーが併存するため、
-- 名前ではなく「search_logs.user_id → users」の外部キーを探してすべて外す。
DO $$
DECLARE
  fk record;
BEGIN
  FOR fk IN
    SELECT c.conname
    FROM pg_constraint c
    JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY (c.conkey)
    WHERE c.contype = 'f'
      AND c.conrelid = 'public.search_logs'::regclass
      AND c.confrelid = 'public.users'::regclass
      AND a.attname = 'user_id'
  LOOP
    EXECUTE format('ALTER TABLE search_logs DROP CONSTRAINT %I', fk.conname);
  END LOOP;
END $$;

ALTER TABLE search_logs
  ADD CONSTRAINT search_logs_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- +migrate Down
ALTER TABLE search_logs DROP CONSTRAINT IF EXISTS search_logs_user_id_fkey;
ALTER TABLE search_logs
  ADD CONSTRAINT search_logs_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
