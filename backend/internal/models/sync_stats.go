package models

import "time"

// SyncStats represents synchronization statistics
type SyncStats struct {
	ID             int       `json:"id" db:"id"`
	SyncType       string    `json:"sync_type" db:"sync_type"` // full/incremental
	StartedAt      time.Time `json:"started_at" db:"started_at"`
	CompletedAt    time.Time `json:"completed_at" db:"completed_at"`
	Duration       int       `json:"duration" db:"duration"` // 秒
	TotalAnimes    int       `json:"total_animes" db:"total_animes"`
	UpdatedAnimes  int       `json:"updated_animes" db:"updated_animes"` // 実際に更新された件数
	SkippedAnimes  int       `json:"skipped_animes" db:"skipped_animes"` // スキップされた件数
	FailedAnimes   int       `json:"failed_animes" db:"failed_animes"`
	MetadataSynced int       `json:"metadata_synced" db:"metadata_synced"` // メタデータ同期件数
	Success        bool      `json:"success" db:"success"`
	ErrorCount     int       `json:"error_count" db:"error_count"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
