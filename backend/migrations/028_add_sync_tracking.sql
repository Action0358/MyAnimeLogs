-- 同期トラッキング用カラム追加
ALTER TABLE animes
ADD COLUMN last_synced_at TIMESTAMP,           -- 最終同期日時
ADD COLUMN annict_updated_at TIMESTAMP,        -- Annict側の更新日時
ADD COLUMN metadata_synced_at TIMESTAMP;       -- メタデータの最終同期日時

-- インデックス作成（差分検出クエリの高速化）
CREATE INDEX idx_animes_last_synced_at ON animes(last_synced_at);
CREATE INDEX idx_animes_annict_updated_at ON animes(annict_updated_at);
CREATE INDEX idx_animes_metadata_synced_at ON animes(metadata_synced_at);
