package models

import "time"

// SyncStatus represents the current status of anime synchronization
type SyncStatus struct {
	LastSyncTime     time.Time `json:"last_sync_time"`
	NextSyncTime     time.Time `json:"next_sync_time"`
	LastSyncDuration string    `json:"last_sync_duration"`
	LastSyncSuccess  bool      `json:"last_sync_success"`
	LastSyncCount    int       `json:"last_sync_count"`
	LastSyncErrors   []string  `json:"last_sync_errors,omitempty"`
	TotalSyncsToday  int       `json:"total_syncs_today"`
	IsRunning        bool      `json:"is_running"`
	Enabled          bool      `json:"enabled"`
	CurrentSeason    string    `json:"current_season"`
	PreviousSeason   string    `json:"previous_season"`
}
