-- +migrate Up
-- お気に入り（user_favorites）と視聴ステータスを責務分離する。
-- user_favorites は「ハート（好きな作品の記録）」のみを担い、
-- 視聴ステータス（視聴予定/視聴中/完了）は独立した user_watch_status テーブルへ移す。
-- これにより「お気に入りではないが視聴ステータスを持つ作品」が表現可能になる。

-- Step 1: 視聴ステータス用テーブルを作成（anime単位）
CREATE TABLE IF NOT EXISTS user_watch_status (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    anime_id INTEGER NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ステータス値は視聴予定/視聴中/完了の3つのみ
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'user_watch_status_status_check') THEN
        ALTER TABLE user_watch_status ADD CONSTRAINT user_watch_status_status_check CHECK (status IN ('watching', 'completed', 'planned'));
    END IF;
END $$;

-- ユーザー×アニメで一意
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'unique_user_anime_watch_status') THEN
        ALTER TABLE user_watch_status ADD CONSTRAINT unique_user_anime_watch_status UNIQUE(user_id, anime_id);
    END IF;
END $$;

-- インデックス
CREATE INDEX IF NOT EXISTS idx_user_watch_status_user_id ON user_watch_status(user_id);
CREATE INDEX IF NOT EXISTS idx_user_watch_status_anime_id ON user_watch_status(anime_id);
CREATE INDEX IF NOT EXISTS idx_user_watch_status_status ON user_watch_status(status);

-- updated_at 自動更新トリガ（users テーブル由来の既存関数を再利用）
DROP TRIGGER IF EXISTS update_user_watch_status_updated_at ON user_watch_status;
CREATE TRIGGER update_user_watch_status_updated_at BEFORE UPDATE ON user_watch_status
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- RLS有効化 + service_role のみ全操作許可（038_enable_rls_all_tables のパターン踏襲）
ALTER TABLE user_watch_status ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'service_role') THEN
    CREATE ROLE service_role;
  END IF;
END $$;

DROP POLICY IF EXISTS "service_role full access" ON user_watch_status;
CREATE POLICY "service_role full access" ON user_watch_status FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Step 2: 既存の user_favorites.status を user_watch_status へ移行
-- 有効な3ステータスのみ移行（dropped/on_hold はお気に入り＝ハートとして残し、視聴ステータスは付与しない）
-- created_at/updated_at は元の値を保全し、一覧の並び順（updated_at DESC）を維持する
INSERT INTO user_watch_status (user_id, anime_id, status, created_at, updated_at)
SELECT user_id, anime_id, status, created_at, updated_at
FROM user_favorites
WHERE status IN ('watching', 'completed', 'planned')
ON CONFLICT (user_id, anime_id) DO NOTHING;

-- Step 3: user_favorites を純粋なお気に入り（ハート）に。status 関連を撤去
DROP INDEX IF EXISTS idx_user_favorites_status;
ALTER TABLE user_favorites DROP CONSTRAINT IF EXISTS user_favorites_status_check;
ALTER TABLE user_favorites DROP COLUMN IF EXISTS status;

-- +migrate Down
-- Step 3: user_favorites に status を復元
ALTER TABLE user_favorites ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'planned';

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'user_favorites_status_check') THEN
        ALTER TABLE user_favorites ADD CONSTRAINT user_favorites_status_check CHECK (status IN ('watching', 'completed', 'planned', 'dropped', 'on_hold'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_user_favorites_status ON user_favorites(status);

-- Step 2: user_watch_status から status を書き戻す（お気に入りが存在する行のみ更新される）
UPDATE user_favorites uf
SET status = uws.status, updated_at = CURRENT_TIMESTAMP
FROM user_watch_status uws
WHERE uf.user_id = uws.user_id AND uf.anime_id = uws.anime_id;

-- Step 1: 視聴ステータステーブルを削除
DROP POLICY IF EXISTS "service_role full access" ON user_watch_status;
DROP TRIGGER IF EXISTS update_user_watch_status_updated_at ON user_watch_status;
DROP TABLE IF EXISTS user_watch_status;
