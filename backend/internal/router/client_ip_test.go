package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// clientIPFor は configureClientIP を適用したエンジンで、c.ClientIP() が返す値を調べる。
func clientIPFor(t *testing.T, environment string, remoteAddr string, headers map[string]string) string {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	configureClientIP(r, environment)

	var got string
	r.GET("/", func(c *gin.Context) {
		got = c.ClientIP()
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	r.ServeHTTP(httptest.NewRecorder(), req)
	return got
}

func TestConfigureClientIP(t *testing.T) {
	const proxyAddr = "172.16.0.1:4321" // Fly のプロキシ（接続元として見えるアドレス）

	t.Run("Fly上では Fly-Client-IP を採用し、偽装した X-Forwarded-For は無視する", func(t *testing.T) {
		for _, env := range []string{"production", "staging"} {
			got := clientIPFor(t, env, proxyAddr, map[string]string{
				"Fly-Client-IP":   "203.0.113.7",
				"X-Forwarded-For": "198.51.100.99, 203.0.113.7",
			})
			assert.Equal(t, "203.0.113.7", got, env)
		}
	})

	t.Run("Fly上でも X-Forwarded-For だけでは偽装できない", func(t *testing.T) {
		got := clientIPFor(t, "production", proxyAddr, map[string]string{
			"X-Forwarded-For": "198.51.100.99",
			"X-Real-IP":       "198.51.100.98",
		})
		assert.Equal(t, "172.16.0.1", got)
	})

	t.Run("ローカルでは Fly-Client-IP も X-Forwarded-For も無視し、接続元を使う", func(t *testing.T) {
		got := clientIPFor(t, "development", "192.0.2.10:5555", map[string]string{
			"Fly-Client-IP":   "198.51.100.1",
			"X-Forwarded-For": "198.51.100.2",
		})
		assert.Equal(t, "192.0.2.10", got)
	})
}
