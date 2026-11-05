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
