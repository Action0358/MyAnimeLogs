-- エピソード同期トラッキング用カラム追加
-- Phase 2: エピソード同期の差分判定に使用
ALTER TABLE animes ADD COLUMN IF NOT EXISTS episodes_synced_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_animes_episodes_synced_at ON animes(episodes_synced_at);
