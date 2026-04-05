package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSyncStatusTracker(t *testing.T) {
	tracker := NewSyncStatusTracker(true)
	assert.NotNil(t, tracker)
	assert.True(t, tracker.enabled)

	tracker2 := NewSyncStatusTracker(false)
	assert.False(t, tracker2.enabled)
}

func TestSyncStatusTracker_StartAndEndSync(t *testing.T) {
	tracker := NewSyncStatusTracker(true)
	nextSync := time.Now().Add(1 * time.Hour)

	tracker.StartSync(nextSync)

	status := tracker.GetStatus(time.Now())
	assert.True(t, status.IsRunning)
	assert.True(t, status.Enabled)

	tracker.EndSync(true, 50, 5*time.Second, nil)

	status = tracker.GetStatus(time.Now())
	assert.False(t, status.IsRunning)
	assert.True(t, status.LastSyncSuccess)
	assert.Equal(t, 50, status.LastSyncCount)
	assert.Equal(t, 1, status.TotalSyncsToday)
}

func TestSyncStatusTracker_EndSync_WithErrors(t *testing.T) {
	tracker := NewSyncStatusTracker(true)
	nextSync := time.Now().Add(1 * time.Hour)

	tracker.StartSync(nextSync)
	errors := []string{"error1", "error2"}
	tracker.EndSync(false, 10, 3*time.Second, errors)

	status := tracker.GetStatus(time.Now())
	assert.False(t, status.LastSyncSuccess)
	assert.Equal(t, 10, status.LastSyncCount)
	assert.Equal(t, errors, status.LastSyncErrors)
}

func TestSyncStatusTracker_MultipleSyncs(t *testing.T) {
	tracker := NewSyncStatusTracker(true)
	nextSync := time.Now().Add(1 * time.Hour)

	tracker.StartSync(nextSync)
	tracker.EndSync(true, 10, time.Second, nil)

	tracker.StartSync(nextSync)
	tracker.EndSync(true, 20, time.Second, nil)

	status := tracker.GetStatus(time.Now())
	assert.Equal(t, 2, status.TotalSyncsToday)
	assert.Equal(t, 20, status.LastSyncCount)
}

func TestSyncStatusTracker_SetNextSyncTime(t *testing.T) {
	tracker := NewSyncStatusTracker(true)
	nextTime := time.Now().Add(2 * time.Hour)

	tracker.SetNextSyncTime(nextTime)

	status := tracker.GetStatus(time.Now())
	assert.Equal(t, nextTime, status.NextSyncTime)
}

func TestSyncStatusTracker_GetStatus_IncludesSeasons(t *testing.T) {
	tracker := NewSyncStatusTracker(true)
	now := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)

	status := tracker.GetStatus(now)
	assert.NotEmpty(t, status.CurrentSeason)
	assert.NotEmpty(t, status.PreviousSeason)
}
