package middleware

import (
	"net/http"
	"sync"
	"time"

	"myanimelogs/internal/common"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
	stopCh   chan struct{}
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
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

	// Add current request
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()

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
