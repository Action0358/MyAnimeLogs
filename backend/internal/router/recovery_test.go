package router

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"myanimelogs/internal/common"
	"myanimelogs/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Setup のミドルウェアの並びで、panic したリクエストの応答とログを確かめる。
func TestSetup_PanicBecomes500WithCORSAndLog(t *testing.T) {
	const origin = "https://myanimelogs.com"
	t.Setenv("CORS_ALLOWED_ORIGINS", origin)

	var logs bytes.Buffer
	logger := &common.Logger{Logger: slog.New(slog.NewJSONHandler(&logs, nil))}
	cfg := &config.Config{Environment: "production", RateLimitPerMinute: 100, ErrorAlertThreshold: 100, ErrorTimeWindow: 1}

	r := Setup(nil, cfg, nil, logger, nil)
	r.GET("/test-panic", func(c *gin.Context) { panic("boom") })

	req := httptest.NewRequest(http.MethodGet, "/test-panic", nil)
	req.Header.Set("Origin", origin)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"), "ブラウザが本文を読めるよう CORS ヘッダーを残す")
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)

	var recorded, completed map[string]any
	for _, line := range bytes.Split(logs.Bytes(), []byte("\n")) {
		var entry map[string]any
		if json.Unmarshal(line, &entry) != nil {
			continue
		}
		switch entry["msg"] {
		case "Error Recorded":
			recorded = entry
		case "HTTP Request Completed":
			completed = entry
		}
	}
	require.NotNil(t, recorded, "panic の内容がログに残る")
	assert.Equal(t, requestID, recorded["request_id"])
	require.NotNil(t, completed, "panic したリクエストもアクセスログに残る")
	assert.EqualValues(t, http.StatusInternalServerError, completed["status"])
}

// Setup のミドルウェアの並びで、途中で打ち切られた応答にもセキュリティ用のヘッダーが付くことを確かめる。
func TestSetup_SecurityHeaders(t *testing.T) {
	const origin = "https://myanimelogs.com"
	t.Setenv("CORS_ALLOWED_ORIGINS", origin)

	newRouter := func(rateLimit int) *gin.Engine {
		logger := &common.Logger{Logger: slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))}
		cfg := &config.Config{Environment: "production", RateLimitPerMinute: rateLimit, ErrorAlertThreshold: 100, ErrorTimeWindow: 1}
		r := Setup(nil, cfg, nil, logger, nil)
		r.GET("/test-ok", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
		r.GET("/test-panic", func(c *gin.Context) { panic("boom") })
		return r
	}
	get := func(path string) *http.Request { return httptest.NewRequest(http.MethodGet, path, nil) }

	cases := []struct {
		name      string
		rateLimit int
		// 同じルーターに先に送るリクエストの数(レート制限を使い切るため)
		warmup int
		req    func() *http.Request
		want   int
	}{
		{"正常", 100, 0, func() *http.Request { return get("/test-ok") }, http.StatusOK},
		{"panic", 100, 0, func() *http.Request { return get("/test-panic") }, http.StatusInternalServerError},
		{"存在しないパス", 100, 0, func() *http.Request { return get("/no-such-path") }, http.StatusNotFound},
		{"本文が大きすぎる", 100, 0, func() *http.Request {
			req := httptest.NewRequest(http.MethodPost, "/test-ok", nil)
			req.ContentLength = 1 << 20
			return req
		}, http.StatusRequestEntityTooLarge},
		{"CORS のプリフライト", 100, 0, func() *http.Request {
			req := httptest.NewRequest(http.MethodOptions, "/test-ok", nil)
			req.Header.Set("Origin", origin)
			req.Header.Set("Access-Control-Request-Method", http.MethodGet)
			return req
		}, http.StatusNoContent},
		{"レート制限", 1, 1, func() *http.Request { return get("/test-ok") }, http.StatusTooManyRequests},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newRouter(tc.rateLimit)
			for i := 0; i < tc.warmup; i++ {
				r.ServeHTTP(httptest.NewRecorder(), tc.req())
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, tc.req())

			assert.Equal(t, tc.want, w.Code)
			assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
			assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))
		})
	}
}
