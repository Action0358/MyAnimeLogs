package scheduler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentSeason(t *testing.T) {
	// Test that current season returns a valid format
	result := getCurrentSeason()

	// Should contain year and season
	assert.Regexp(t, `^\d{4}-(winter|spring|summer|autumn)$`, result)
}

func TestGetPreviousSeason(t *testing.T) {
	// Test that previous season returns a valid format
	result := getPreviousSeason()

	// Should contain year and season
	assert.Regexp(t, `^\d{4}-(winter|spring|summer|autumn)$`, result)

	// Previous season should be different from current
	current := getCurrentSeason()
	assert.NotEqual(t, current, result)
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
