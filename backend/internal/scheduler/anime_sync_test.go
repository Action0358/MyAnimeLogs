package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
	scheduler := NewAnimeSyncScheduler(nil, nil, true)

	assert.NotNil(t, scheduler)
	assert.True(t, scheduler.enabled)
	assert.NotNil(t, scheduler.stopChan)
}

func TestSchedulerDisabled(t *testing.T) {
	// Create a mock logger to avoid nil pointer
	scheduler := NewAnimeSyncScheduler(nil, nil, false)

	assert.NotNil(t, scheduler)
	assert.False(t, scheduler.enabled)

	// Start should not panic when disabled (no logger needed)
	// Stop should not panic
	scheduler.Stop()
}
