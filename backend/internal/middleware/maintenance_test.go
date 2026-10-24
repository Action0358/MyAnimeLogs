package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newMaintenanceRouter(enabled bool, endsAt *time.Time) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(MaintenanceMiddleware(enabled, endsAt))
	ok := func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }
	r.GET("/health", ok)
	r.GET("/api/animes", ok)
	r.POST("/api/admin/sync/season", ok)
	r.OPTIONS("/api/animes", ok)
	return r
}

func doRequest(r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(method, path, nil))
	return w
}

func TestMaintenanceMiddleware_Disabled(t *testing.T) {
	r := newMaintenanceRouter(false, nil)

	assert.Equal(t, http.StatusOK, doRequest(r, http.MethodGet, "/api/animes").Code)
}

func TestMaintenanceMiddleware_Enabled(t *testing.T) {
	t.Run("通常のAPIは503 MAINTENANCEを返す", func(t *testing.T) {
		w := doRequest(newMaintenanceRouter(true, nil), http.MethodGet, "/api/animes")

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))
		assert.Equal(t, "300", w.Header().Get("Retry-After"))

		var body map[string]any
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "MAINTENANCE", body["code"])
		assert.NotContains(t, body, "ends_at")
	})

	t.Run("終了予定があれば ends_at と Retry-After に反映する", func(t *testing.T) {
		endsAt := time.Now().Add(10 * time.Minute)
		w := doRequest(newMaintenanceRouter(true, &endsAt), http.MethodGet, "/api/animes")

		var body map[string]any
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, endsAt.Format(time.RFC3339), body["ends_at"])

		retryAfter, err := strconv.Atoi(w.Header().Get("Retry-After"))
		assert.NoError(t, err)
		assert.InDelta(t, 600, retryAfter, 5)
	})

	t.Run("終了予定を過ぎていたら既定の Retry-After にする", func(t *testing.T) {
		endsAt := time.Now().Add(-time.Hour)
		w := doRequest(newMaintenanceRouter(true, &endsAt), http.MethodGet, "/api/animes")

		assert.Equal(t, "300", w.Header().Get("Retry-After"))
	})

	t.Run("/health・管理系・OPTIONSは通す", func(t *testing.T) {
		r := newMaintenanceRouter(true, nil)

		assert.Equal(t, http.StatusOK, doRequest(r, http.MethodGet, "/health").Code)
		assert.Equal(t, http.StatusOK, doRequest(r, http.MethodPost, "/api/admin/sync/season").Code)
		assert.Equal(t, http.StatusOK, doRequest(r, http.MethodOptions, "/api/animes").Code)
	})
}
