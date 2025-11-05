package scheduler

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/services"

	"github.com/stretchr/testify/assert"
)

// newTestLogger creates a logger that discards output for testing
func newTestLogger() *common.Logger {
	handler := slog.NewTextHandler(io.Discard, nil)
	return &common.Logger{
		Logger: slog.New(handler),
	}
}

// newTestAnnictClient creates a minimal AnnictClient for testing
func newTestAnnictClient() *services.AnnictClient {
	return services.NewAnnictClient("")
}

func TestGetSeasonForTime(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"January - Winter", time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), "2025-winter"},
		{"March - Winter", time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC), "2025-winter"},
		{"April - Spring", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC), "2025-spring"},
		{"June - Spring", time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC), "2025-spring"},
		{"July - Summer", time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), "2025-summer"},
		{"September - Summer", time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC), "2025-summer"},
		{"October - Autumn", time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC), "2025-autumn"},
		{"December - Autumn", time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "2025-autumn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSeasonForTime(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPreviousSeasonForTime(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"January - Previous Autumn", time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), "2024-autumn"},
		{"April - Previous Winter", time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC), "2025-winter"},
		{"July - Previous Spring", time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC), "2025-spring"},
		{"October - Previous Summer", time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC), "2025-summer"},
		{"Year boundary", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), "2024-autumn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPreviousSeasonForTime(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewAnimeSyncScheduler(t *testing.T) {
	logger := newTestLogger()
	client := newTestAnnictClient()
	scheduler := NewAnimeSyncScheduler(client, logger, true, 3, 0)

	assert.NotNil(t, scheduler)
	assert.True(t, scheduler.enabled)
	assert.NotNil(t, scheduler.stopChan)
	assert.NotNil(t, scheduler.logger)
	assert.NotNil(t, scheduler.annictClient)
	assert.NotNil(t, scheduler.statusTracker)
	assert.Equal(t, 3, scheduler.runHour)
	assert.Equal(t, 0, scheduler.runMinute)
}

func TestSchedulerDisabled(t *testing.T) {
	logger := newTestLogger()
	client := newTestAnnictClient()
	scheduler := NewAnimeSyncScheduler(client, logger, false, 3, 0)

	assert.NotNil(t, scheduler)
	assert.False(t, scheduler.enabled)
	assert.NotNil(t, scheduler.annictClient)

	// Start should not panic and should log that it's disabled
	scheduler.Start()

	// Stop should not panic
	scheduler.Stop()
}

func TestGetStatus(t *testing.T) {
	logger := newTestLogger()
	client := newTestAnnictClient()
	scheduler := NewAnimeSyncScheduler(client, logger, true, 3, 0)

	now := time.Date(2025, 11, 5, 12, 0, 0, 0, time.UTC)
	status := scheduler.GetStatus()

	assert.True(t, status.Enabled)
	assert.Equal(t, "2025-autumn", getSeasonForTime(now))
	assert.Equal(t, "2025-summer", getPreviousSeasonForTime(now))
}

func TestSyncStatusTracker_DayReset(t *testing.T) {
	tracker := NewSyncStatusTracker(true)

	// Set last reset to yesterday
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tracker.mu.Lock()
	tracker.lastResetDate = yesterday
	tracker.totalSyncsToday = 5
	tracker.mu.Unlock()

	// Start sync today - should reset because it's a new day
	nextSync := now.Add(24 * time.Hour)
	tracker.StartSync(nextSync)

	tracker.mu.RLock()
	count := tracker.totalSyncsToday
	tracker.mu.RUnlock()

	assert.Equal(t, 0, count, "Counter should reset on new day")

	// End sync - should be 1
	tracker.EndSync(true, 10, time.Second, nil)

	tracker.mu.RLock()
	count = tracker.totalSyncsToday
	tracker.mu.RUnlock()

	assert.Equal(t, 1, count, "Counter should be 1 after first sync of new day")
}

func TestSyncStatusTracker_SameDayNoReset(t *testing.T) {
	tracker := NewSyncStatusTracker(true)

	// Set last reset to today
	now := time.Now()
	tracker.mu.Lock()
	tracker.lastResetDate = now
	tracker.totalSyncsToday = 5
	tracker.mu.Unlock()

	// Start sync today - should NOT reset
	nextSync := now.Add(2 * time.Hour)
	tracker.StartSync(nextSync)

	tracker.mu.RLock()
	count := tracker.totalSyncsToday
	tracker.mu.RUnlock()

	assert.Equal(t, 5, count, "Counter should not reset on same day")

	// End sync - should increment to 6
	tracker.EndSync(true, 10, time.Second, nil)

	tracker.mu.RLock()
	count = tracker.totalSyncsToday
	tracker.mu.RUnlock()

	assert.Equal(t, 6, count, "Counter should increment")
}

func TestSyncStatusTracker_StatusFields(t *testing.T) {
	tracker := NewSyncStatusTracker(true)

	// Start a sync
	nextSync := time.Now().Add(24 * time.Hour)
	tracker.StartSync(nextSync)

	tracker.mu.RLock()
	isRunning := tracker.isRunning
	tracker.mu.RUnlock()

	assert.True(t, isRunning, "Should be running after StartSync")

	// End the sync
	errors := []string{"test error"}
	tracker.EndSync(false, 42, 5*time.Second, errors)

	tracker.mu.RLock()
	isRunning = tracker.isRunning
	success := tracker.lastSyncSuccess
	count := tracker.lastSyncCount
	duration := tracker.lastSyncDuration
	syncErrors := tracker.lastSyncErrors
	tracker.mu.RUnlock()

	assert.False(t, isRunning, "Should not be running after EndSync")
	assert.False(t, success, "Should record failure")
	assert.Equal(t, 42, count, "Should record count")
	assert.Equal(t, 5*time.Second, duration, "Should record duration")
	assert.Equal(t, errors, syncErrors, "Should record errors")
}

func TestSchedulerWithNilLogger(t *testing.T) {
	client := newTestAnnictClient()
	scheduler := NewAnimeSyncScheduler(client, nil, true, 3, 0)

	// Should not panic - logger is initialized in constructor
	assert.NotNil(t, scheduler.logger)

	// Start and Stop should not panic
	scheduler.Start()
	scheduler.Stop()
}

func TestSchedulerCustomRunTime(t *testing.T) {
	logger := newTestLogger()
	client := newTestAnnictClient()
	scheduler := NewAnimeSyncScheduler(client, logger, true, 14, 30)

	assert.Equal(t, 14, scheduler.runHour)
	assert.Equal(t, 30, scheduler.runMinute)
}
