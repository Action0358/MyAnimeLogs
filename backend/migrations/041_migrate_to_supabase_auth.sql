-- +migrate Up
-- Supabase Auth移行: 自前JWT認証を廃止し、Supabase Authに委譲する

-- Step 1: sessionsテーブルを削除（Supabaseがセッション管理を担当）
DROP TABLE IF EXISTS sessions;

-- Step 2: usersテーブルから認証関連カラムを削除
ALTER TABLE users DROP COLUMN IF EXISTS password;
ALTER TABLE users DROP COLUMN IF EXISTS failed_login_attempts;
ALTER TABLE users DROP COLUMN IF EXISTS last_failed_login_at;
ALTER TABLE users DROP COLUMN IF EXISTS locked_until;
ALTER TABLE users DROP COLUMN IF EXISTS current_jti;
ALTER TABLE users DROP COLUMN IF EXISTS last_login_at;

-- Step 3: Supabase AuthユーザーIDカラムを追加
ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_user_id UUID;

-- Step 4: auth_user_idにユニークインデックスを作成
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_auth_user_id ON users(auth_user_id)
  WHERE auth_user_id IS NOT NULL;

-- Step 5: Supabase auth.usersへの外部キーを追加（Supabase環境のみ）
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'auth' AND table_name = 'users'
  ) THEN
    IF NOT EXISTS (
      SELECT 1 FROM information_schema.table_constraints
      WHERE constraint_name = 'fk_users_auth_user_id'
        AND table_name = 'users'
        AND table_schema = 'public'
    ) THEN
      ALTER TABLE users
        ADD CONSTRAINT fk_users_auth_user_id
        FOREIGN KEY (auth_user_id)
        REFERENCES auth.users(id)
        ON DELETE CASCADE;
    END IF;
  END IF;
END $$;

-- +migrate Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_auth_user_id;
DROP INDEX IF EXISTS idx_users_auth_user_id;
ALTER TABLE users DROP COLUMN IF EXISTS auth_user_id;
