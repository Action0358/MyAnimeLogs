package scheduler

import (
	"sync"
	"time"

	"myanimelogs/internal/models"
)

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

	// Reset daily counter if it's a new day (using Truncate for robust date comparison)
	now := time.Now()
	nowDate := now.Truncate(24 * time.Hour)
	lastResetDate := t.lastResetDate.Truncate(24 * time.Hour)

	if !nowDate.Equal(lastResetDate) {
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
func (t *SyncStatusTracker) GetStatus(now time.Time) models.SyncStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return models.SyncStatus{
		LastSyncTime:     t.lastSyncTime,
		NextSyncTime:     t.nextSyncTime,
		LastSyncDuration: t.lastSyncDuration.String(),
		LastSyncSuccess:  t.lastSyncSuccess,
		LastSyncCount:    t.lastSyncCount,
		LastSyncErrors:   t.lastSyncErrors,
		TotalSyncsToday:  t.totalSyncsToday,
		IsRunning:        t.isRunning,
		Enabled:          t.enabled,
		CurrentSeason:    getSeasonForTime(now),
		PreviousSeason:   getPreviousSeasonForTime(now),
	}
}

// SetNextSyncTime updates the next sync time
func (t *SyncStatusTracker) SetNextSyncTime(nextTime time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.nextSyncTime = nextTime
}
