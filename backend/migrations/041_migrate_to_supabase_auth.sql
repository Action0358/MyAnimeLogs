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
-- Step 5: Supabase auth外部キーを削除
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_auth_user_id;

-- Step 4: auth_user_idインデックスを削除
DROP INDEX IF EXISTS idx_users_auth_user_id;

-- Step 3: auth_user_idカラムを削除
ALTER TABLE users DROP COLUMN IF EXISTS auth_user_id;

-- Step 2: 認証関連カラムを復元
-- ※ パスワードハッシュ等のデータは復元不可。スキーマ構造のみ復元する
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS password VARCHAR(255),
  ADD COLUMN IF NOT EXISTS failed_login_attempts INT DEFAULT 0,
  ADD COLUMN IF NOT EXISTS last_failed_login_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS locked_until TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_users_locked_until ON users(locked_until)
  WHERE locked_until IS NOT NULL;

-- Step 1: sessionsテーブルを復元
CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(44) NOT NULL UNIQUE,
    access_token_jti VARCHAR(36) NOT NULL,
    user_agent TEXT,
    ip_address VARCHAR(45),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_last_activity_at ON sessions(last_activity_at);
CREATE INDEX IF NOT EXISTS idx_sessions_access_token_jti ON sessions(access_token_jti);
