package middleware

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)
	defer rl.Stop()
	assert.NotNil(t, rl)
	assert.Equal(t, 10, rl.limit)
	assert.Equal(t, time.Minute, rl.window)
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	defer rl.Stop()

	// 制限内のリクエストは許可
	assert.True(t, rl.Allow("client1"))
	assert.True(t, rl.Allow("client1"))
	assert.True(t, rl.Allow("client1"))

	// 制限超過は拒否
	assert.False(t, rl.Allow("client1"))
}

func TestRateLimiter_Allow_DifferentKeys(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	defer rl.Stop()

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
	defer rl.Stop()

	assert.True(t, rl.Allow("client1"))
	assert.False(t, rl.Allow("client1"))

	// ウィンドウが過ぎるのを待つ
	time.Sleep(60 * time.Millisecond)

	// ウィンドウ経過後は再び許可
	assert.True(t, rl.Allow("client1"))
}

func TestRateLimiter_Cleanup_RemovesExpiredKeys(t *testing.T) {
	// 短いウィンドウでクリーンアップの動作を検証
	rl := NewRateLimiter(1, 30*time.Millisecond)
	defer rl.Stop()

	// 多数のユニークキーを登録
	for i := 0; i < 100; i++ {
		rl.Allow(fmt.Sprintf("client-%d", i))
	}
	assert.Equal(t, 100, rl.Len())

	// ウィンドウ×2（クリーンアップ間隔）+ マージンを待つ
	time.Sleep(100 * time.Millisecond)

	// クリーンアップ後、期限切れキーが削除されている
	assert.Equal(t, 0, rl.Len(), "期限切れのキーがクリーンアップされていない")
}

func TestNewRateLimiter_ZeroWindow(t *testing.T) {
	rl := NewRateLimiter(10, 0)
	defer rl.Stop()
	assert.Equal(t, time.Second, rl.window, "windowが0の場合はデフォルト1秒になるべき")
}

func TestNewRateLimiter_NegativeWindow(t *testing.T) {
	rl := NewRateLimiter(10, -time.Minute)
	defer rl.Stop()
	assert.Equal(t, time.Second, rl.window, "windowが負の場合はデフォルト1秒になるべき")
}
