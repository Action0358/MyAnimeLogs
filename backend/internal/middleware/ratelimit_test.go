package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)
	assert.NotNil(t, rl)
	assert.Equal(t, 10, rl.limit)
	assert.Equal(t, time.Minute, rl.window)
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	// 制限内のリクエストは許可
	assert.True(t, rl.Allow("client1"))
	assert.True(t, rl.Allow("client1"))
	assert.True(t, rl.Allow("client1"))

	// 制限超過は拒否
	assert.False(t, rl.Allow("client1"))
}

func TestRateLimiter_Allow_DifferentKeys(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)

	// 異なるキーは独立してカウント
	assert.True(t, rl.Allow("client1"))
	assert.True(t, rl.Allow("client2"))

	// それぞれの制限は独立
	assert.False(t, rl.Allow("client1"))
	assert.False(t, rl.Allow("client2"))
}

func TestRateLimiter_Allow_WindowExpiry(t *testing.T) {
	// 非常に短いウィンドウでテスト
	rl := NewRateLimiter(1, 50*time.Millisecond)

	assert.True(t, rl.Allow("client1"))
	assert.False(t, rl.Allow("client1"))

	// ウィンドウが過ぎるのを待つ
	time.Sleep(60 * time.Millisecond)

	// ウィンドウ経過後は再び許可
	assert.True(t, rl.Allow("client1"))
}
