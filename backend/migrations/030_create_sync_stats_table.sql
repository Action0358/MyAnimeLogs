-- 同期統計テーブル作成
CREATE TABLE IF NOT EXISTS sync_stats (
    id SERIAL PRIMARY KEY,
    sync_type VARCHAR(20) NOT NULL,           -- full/incremental
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP NOT NULL,
    duration INTEGER NOT NULL,                -- 秒
    total_animes INTEGER NOT NULL,
    updated_animes INTEGER NOT NULL,          -- 実際に更新された件数
    skipped_animes INTEGER NOT NULL,          -- スキップされた件数
    failed_animes INTEGER NOT NULL,
    metadata_synced INTEGER NOT NULL,         -- メタデータ同期件数
    success BOOLEAN NOT NULL,
    error_count INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- インデックス作成
CREATE INDEX idx_sync_stats_started_at ON sync_stats(started_at);
CREATE INDEX idx_sync_stats_sync_type ON sync_stats(sync_type);
CREATE INDEX idx_sync_stats_success ON sync_stats(success);
