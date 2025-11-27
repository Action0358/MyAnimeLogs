package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestAniListClientRateLimiter tests that the rate limiter is properly initialized
func TestAniListClientRateLimiter(t *testing.T) {
	client := NewAniListClient()

	assert.NotNil(t, client.rateLimiter, "Rate limiter should be initialized")
}

// TestAniListClientRateLimiterThrottling tests that the rate limiter actually throttles requests
func TestAniListClientRateLimiterThrottling(t *testing.T) {
	client := NewAniListClient()
	ctx := context.Background()

	// Measure time for 3 consecutive requests
	start := time.Now()

	// Make 3 requests (should be throttled by rate limiter)
	// This mirrors the production usage in executeGraphQLQuery
	for i := 0; i < 3; i++ {
		err := client.rateLimiter.Wait(ctx)
		assert.NoError(t, err, "Wait should not return an error")
	}

	elapsed := time.Since(start)

	// With 85 requests/minute, each request should take ~706ms
	// 3 requests should take at least ~1.4 seconds (2 * 706ms)
	// We use a lower bound to account for timing variations
	minExpectedDuration := 1200 * time.Millisecond

	assert.GreaterOrEqual(t, elapsed, minExpectedDuration,
		"Rate limiter should throttle requests (elapsed: %v, expected: >= %v)",
		elapsed, minExpectedDuration)
}

// TestAniListClientRateLimiterBurst tests that the rate limiter allows burst of 1
func TestAniListClientRateLimiterBurst(t *testing.T) {
	client := NewAniListClient()

	// First request should be immediate (burst = 1)
	start := time.Now()
	allowed := client.rateLimiter.Allow()
	elapsed := time.Since(start)

	assert.True(t, allowed, "First request should be allowed immediately")
	assert.Less(t, elapsed, 10*time.Millisecond, "First request should not be delayed")

	// Second immediate request should be blocked (no burst available)
	allowed = client.rateLimiter.Allow()
	assert.False(t, allowed, "Second immediate request should be blocked")
}
