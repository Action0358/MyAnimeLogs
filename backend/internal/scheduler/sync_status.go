package scheduler

import (
	"sync"
	"time"
)

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

// SyncStatusTracker tracks the status of anime synchronization
type SyncStatusTracker struct {
	mu               sync.RWMutex
	lastSyncTime     time.Time
	nextSyncTime     time.Time
	lastSyncDuration time.Duration
	lastSyncSuccess  bool
	lastSyncCount    int
	lastSyncErrors   []string
	totalSyncsToday  int
	lastResetDate    time.Time
	isRunning        bool
	enabled          bool
}

// NewSyncStatusTracker creates a new sync status tracker
func NewSyncStatusTracker(enabled bool) *SyncStatusTracker {
	return &SyncStatusTracker{
		enabled:       enabled,
		lastResetDate: time.Now(),
	}
}

// StartSync marks the beginning of a sync operation
func (t *SyncStatusTracker) StartSync(nextSyncTime time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.isRunning = true
	t.nextSyncTime = nextSyncTime

	// Reset daily counter if it's a new day
	now := time.Now()
	if now.Day() != t.lastResetDate.Day() || now.Month() != t.lastResetDate.Month() {
		t.totalSyncsToday = 0
		t.lastResetDate = now
	}
}

// EndSync marks the end of a sync operation
func (t *SyncStatusTracker) EndSync(success bool, count int, duration time.Duration, errors []string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.isRunning = false
	t.lastSyncTime = time.Now()
	t.lastSyncDuration = duration
	t.lastSyncSuccess = success
	t.lastSyncCount = count
	t.lastSyncErrors = errors
	t.totalSyncsToday++
}

// GetStatus returns the current sync status
func (t *SyncStatusTracker) GetStatus() SyncStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return SyncStatus{
		LastSyncTime:     t.lastSyncTime,
		NextSyncTime:     t.nextSyncTime,
		LastSyncDuration: t.lastSyncDuration.String(),
		LastSyncSuccess:  t.lastSyncSuccess,
		LastSyncCount:    t.lastSyncCount,
		LastSyncErrors:   t.lastSyncErrors,
		TotalSyncsToday:  t.totalSyncsToday,
		IsRunning:        t.isRunning,
		Enabled:          t.enabled,
		CurrentSeason:    getCurrentSeason(),
		PreviousSeason:   getPreviousSeason(),
	}
}

// SetNextSyncTime updates the next sync time
func (t *SyncStatusTracker) SetNextSyncTime(nextTime time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.nextSyncTime = nextTime
}
