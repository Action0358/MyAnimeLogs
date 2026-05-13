-- +migrate Up
-- Supabase PostgREST API経由の不正アクセスを防ぐため、全テーブルにRLSを有効化する
-- GoバックエンドはDATABASE_URL（直接PostgreSQL接続）を使用しており、RLSの影響を受けない

ALTER TABLE users           ENABLE ROW LEVEL SECURITY;
ALTER TABLE animes          ENABLE ROW LEVEL SECURITY;
ALTER TABLE episodes        ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_favorites  ENABLE ROW LEVEL SECURITY;
ALTER TABLE watch_statuses  ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_ratings    ENABLE ROW LEVEL SECURITY;
ALTER TABLE watch_history   ENABLE ROW LEVEL SECURITY;
ALTER TABLE anime_genres    ENABLE ROW LEVEL SECURITY;
ALTER TABLE search_logs     ENABLE ROW LEVEL SECURITY;
ALTER TABLE sessions        ENABLE ROW LEVEL SECURITY;
ALTER TABLE sync_stats      ENABLE ROW LEVEL SECURITY;

-- service_roleのみ全操作を許可（PostgREST経由のanonアクセスを拒否）
CREATE POLICY "service_role full access" ON users           FOR ALL TO service_role USING (true) WITH CHECK (true);
CREATE POLICY "service_role full access" ON animes          FOR ALL TO service_role USING (true) WITH CHECK (true);
CREATE POLICY "service_role full access" ON episodes        FOR ALL TO service_role USING (true) WITH CHECK (true);
CREATE POLICY "service_role full access" ON user_favorites  FOR ALL TO service_role USING (true) WITH CHECK (true);
CREATE POLICY "service_role full access" ON watch_statuses  FOR ALL TO service_role USING (true) WITH CHECK (true);
CREATE POLICY "service_role full access" ON user_ratings    FOR ALL TO service_role USING (true) WITH CHECK (true);
CREATE POLICY "service_role full access" ON watch_history   FOR ALL TO service_role USING (true) WITH CHECK (true);
CREATE POLICY "service_role full access" ON anime_genres    FOR ALL TO service_role USING (true) WITH CHECK (true);
CREATE POLICY "service_role full access" ON search_logs     FOR ALL TO service_role USING (true) WITH CHECK (true);
CREATE POLICY "service_role full access" ON sessions        FOR ALL TO service_role USING (true) WITH CHECK (true);
CREATE POLICY "service_role full access" ON sync_stats      FOR ALL TO service_role USING (true) WITH CHECK (true);

-- +migrate Down
DROP POLICY IF EXISTS "service_role full access" ON users;
DROP POLICY IF EXISTS "service_role full access" ON animes;
DROP POLICY IF EXISTS "service_role full access" ON episodes;
DROP POLICY IF EXISTS "service_role full access" ON user_favorites;
DROP POLICY IF EXISTS "service_role full access" ON watch_statuses;
DROP POLICY IF EXISTS "service_role full access" ON user_ratings;
DROP POLICY IF EXISTS "service_role full access" ON watch_history;
DROP POLICY IF EXISTS "service_role full access" ON anime_genres;
DROP POLICY IF EXISTS "service_role full access" ON search_logs;
DROP POLICY IF EXISTS "service_role full access" ON sessions;
DROP POLICY IF EXISTS "service_role full access" ON sync_stats;

ALTER TABLE users           DISABLE ROW LEVEL SECURITY;
ALTER TABLE animes          DISABLE ROW LEVEL SECURITY;
ALTER TABLE episodes        DISABLE ROW LEVEL SECURITY;
ALTER TABLE user_favorites  DISABLE ROW LEVEL SECURITY;
ALTER TABLE watch_statuses  DISABLE ROW LEVEL SECURITY;
ALTER TABLE user_ratings    DISABLE ROW LEVEL SECURITY;
ALTER TABLE watch_history   DISABLE ROW LEVEL SECURITY;
ALTER TABLE anime_genres    DISABLE ROW LEVEL SECURITY;
ALTER TABLE search_logs     DISABLE ROW LEVEL SECURITY;
ALTER TABLE sessions        DISABLE ROW LEVEL SECURITY;
ALTER TABLE sync_stats      DISABLE ROW LEVEL SECURITY;
