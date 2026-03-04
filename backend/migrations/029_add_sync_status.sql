-- 同期ステータス管理
ALTER TABLE animes
ADD COLUMN sync_status VARCHAR(20) DEFAULT 'pending',  -- pending/synced/failed
ADD COLUMN sync_error TEXT;                            -- エラー内容

-- インデックス作成
CREATE INDEX idx_animes_sync_status ON animes(sync_status);
