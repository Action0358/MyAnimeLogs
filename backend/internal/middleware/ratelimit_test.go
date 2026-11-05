package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
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

func TestRateLimiter_MaxKeys_BoundsMemoryWithoutRejectingNewClients(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	defer rl.Stop()
	rl.maxKeys = 10

	for i := 0; i < 1000; i++ {
		assert.True(t, rl.Allow(fmt.Sprintf("client-%d", i)), "新しいキーは上限に達していても拒否しない")
		assert.LessOrEqual(t, rl.Len(), 10)
	}
}

func TestRateLimiter_MaxKeys_ExistingKeyDoesNotEvict(t *testing.T) {
	rl := NewRateLimiter(100, time.Minute)
	defer rl.Stop()
	rl.maxKeys = 3

	for _, key := range []string{"a", "b", "c"} {
		assert.True(t, rl.Allow(key))
	}

	// 上限ちょうどの状態で既存のキーが来ても、他のキーを消さない
	for i := 0; i < 50; i++ {
		assert.True(t, rl.Allow("a"))
	}
	assert.Equal(t, 3, rl.Len())
	assert.Contains(t, rl.requests, "b")
	assert.Contains(t, rl.requests, "c")
}

func TestRateLimitKey(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want string
	}{
		{"IPv4 is unchanged", "203.0.113.7", "203.0.113.7"},
		{"IPv4-mapped IPv6 becomes IPv4", "::ffff:203.0.113.7", "203.0.113.7"},
		{"IPv6 is grouped by /64", "2001:db8:1:2:aaaa:bbbb:cccc:dddd", "2001:db8:1:2::/64"},
		{"IPv6 notation does not matter", "2001:0DB8:0001:0002:0:0:0:1", "2001:db8:1:2::/64"},
		{"not an IP is kept as is", "not-an-ip", "not-an-ip"},
		{"empty is kept as is", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, rateLimitKey(tt.ip))
		})
	}
}

func TestRateLimitMiddleware_CountsByClientNetwork(t *testing.T) {
	const (
		sameNetworkFirst  = "2001:db8:1:2::1"
		sameNetworkSecond = "2001:db8:1:2:ffff:ffff:ffff:ffff"
		otherNetwork      = "2001:db8:1:3::1"
		ipv4              = "203.0.113.7"
		ipv4Mapped        = "::ffff:203.0.113.7"
	)

	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(1, time.Minute)
	defer rl.Stop()
	r := gin.New()
	r.TrustedPlatform = "Fly-Client-IP"
	r.Use(RateLimitMiddleware(rl))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	request := func(ip string) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Fly-Client-IP", ip)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	assert.Equal(t, http.StatusOK, request(sameNetworkFirst))
	assert.Equal(t, http.StatusTooManyRequests, request(sameNetworkSecond), "同じ /64 ならアドレスを変えても制限される")
	assert.Equal(t, http.StatusOK, request(otherNetwork), "別の /64 は別に数える")

	assert.Equal(t, http.StatusOK, request(ipv4))
	assert.Equal(t, http.StatusTooManyRequests, request(ipv4Mapped), "IPv4 射影アドレスは IPv4 と同じに数える")
}

// BenchmarkRateLimiter_AllowAtMaxKeys は、上限に達した状態で新しいキーが続いても
// 1回の判定が全件走査にならないことを確認する。
func BenchmarkRateLimiter_AllowAtMaxKeys(b *testing.B) {
	rl := NewRateLimiter(1, time.Minute)
	defer rl.Stop()
	for i := 0; i < rl.maxKeys; i++ {
		rl.Allow(fmt.Sprintf("seed-%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow(fmt.Sprintf("new-%d", i))
	}
}
