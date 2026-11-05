package middleware

import (
	"net/http"
	"net/netip"
	"sync"
	"time"

	"myanimelogs/internal/common"

	"github.com/gin-gonic/gin"
)

const (
	// ipv6PrefixBits は IPv6 を1つの利用者として数える範囲。
	// 利用者は通常 /64 を丸ごと割り当てられ、その中でアドレスを自由に変えられる
	ipv6PrefixBits = 64

	// defaultMaxKeys は記録するキー数の上限。1キー1リクエストなら約9MBで収まる
	defaultMaxKeys = 50000
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
	maxKeys  int
	stopCh   chan struct{}
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if window <= 0 {
		window = time.Second
	}
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
		maxKeys:  defaultMaxKeys,
		stopCh:   make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

// cleanup は定期的に期限切れのキーをマップから削除してメモリリークを防ぐ
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window * 2)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mutex.Lock()
			now := time.Now()
			windowStart := now.Add(-rl.window)
			for key, requests := range rl.requests {
				valid := requests[:0]
				for _, t := range requests {
					if t.After(windowStart) {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(rl.requests, key)
				} else {
					rl.requests[key] = valid
				}
			}
			rl.mutex.Unlock()
		case <-rl.stopCh:
			return
		}
	}
}

// Stop はクリーンアップgoroutineを停止する
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// Len はテスト用: 現在追跡中のキー数を返す
func (rl *RateLimiter) Len() int {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()
	return len(rl.requests)
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Clean old requests
	if requests, exists := rl.requests[key]; exists {
		validRequests := []time.Time{}
		for _, reqTime := range requests {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[key] = validRequests
	}

	// Check if limit exceeded
	if len(rl.requests[key]) >= rl.limit {
		return false
	}

	// 上限に達していたら、どれか1つを消して新しいキーの場所を空ける。
	// 新しいキーを拒否すると、キーを溢れさせるだけで他の利用者を締め出せてしまう。
	// 一番古いキーを探すと、攻撃中に毎回全件を走査することになるので、任意の1つを消す
	if _, exists := rl.requests[key]; !exists && len(rl.requests) >= rl.maxKeys {
		for evicted := range rl.requests {
			delete(rl.requests, evicted)
			break
		}
	}

	// Add current request
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

// rateLimitKey はクライアントIPをレート制限のキーに変換する。
// IPv6 は /64 にまとめ、IPv4 射影アドレスは IPv4 と同じキーにする。
func rateLimitKey(clientIP string) string {
	addr, err := netip.ParseAddr(clientIP)
	if err != nil {
		return clientIP
	}
	addr = addr.Unmap()
	if !addr.Is6() {
		return addr.String()
	}
	prefix, err := addr.Prefix(ipv6PrefixBits)
	if err != nil {
		return clientIP
	}
	return prefix.String()
}

func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := rateLimitKey(c.ClientIP())

		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, common.ErrorResponse{
				Error:   common.ErrMsgTooManyRequests,
				Message: "リクエスト数が上限を超えました。しばらく待ってから再試行してください。",
				Code:    common.ErrCodeTooManyRequests,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
