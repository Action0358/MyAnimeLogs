package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// preflightStatus は Origin 付きのプリフライトを CORSMiddleware に通し、結果を返す。
func preflightStatus(origin string) (int, string) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/api/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodOptions, "/api/health", nil)
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code, w.Header().Get("Access-Control-Allow-Origin")
}

func TestCORSMiddleware_AllowedOrigins(t *testing.T) {
	tests := []struct {
		name   string
		env    string
		origin string
		want   bool
	}{
		// 本番: 設定したオリジンだけを許可する
		{"production: own origin", "https://myanimelogs.com,https://www.myanimelogs.com", "https://myanimelogs.com", true},
		{"production: second origin with space", "https://myanimelogs.com, https://www.myanimelogs.com", "https://www.myanimelogs.com", true},
		{"production: staging origin", "https://myanimelogs.com", "https://stg.myanimelogs.com", false},
		{"production: pages apex", "https://myanimelogs.com", "https://myanimelogs.pages.dev", false},
		{"production: pages preview", "https://myanimelogs.com", "https://abc123.myanimelogs.pages.dev", false},
		{"production: staging pages apex", "https://myanimelogs.com", "https://myanimelogs-staging.pages.dev", false},
		{"production: staging pages alias", "https://myanimelogs.com", "https://staging.myanimelogs-staging.pages.dev", false},
		{"production: localhost", "https://myanimelogs.com", "http://localhost:3000", false},

		// staging: 本番のオリジンと pages.dev は許可しない
		{"staging: own origin", "https://stg.myanimelogs.com", "https://stg.myanimelogs.com", true},
		{"staging: production origin", "https://stg.myanimelogs.com", "https://myanimelogs.com", false},
		{"staging: pages apex", "https://stg.myanimelogs.com", "https://myanimelogs-staging.pages.dev", false},
		{"staging: pages alias", "https://stg.myanimelogs.com", "https://staging.myanimelogs-staging.pages.dev", false},

		// 完全一致のみ: 前後を変えたものは通さない
		{"subdomain of allowed origin", "https://myanimelogs.com", "https://evil.myanimelogs.com", false},
		{"allowed origin as subdomain of attacker", "https://myanimelogs.com", "https://myanimelogs.com.evil.com", false},
		{"http instead of https", "https://myanimelogs.com", "http://myanimelogs.com", false},
		{"unrelated origin", "https://myanimelogs.com", "https://evil.example.com", false},

		// 未設定: ローカル開発用のオリジンだけを許可する
		{"unset: localhost 3000", "", "http://localhost:3000", true},
		{"unset: localhost 3001", "", "http://localhost:3001", true},
		{"unset: production origin", "", "https://myanimelogs.com", false},
		{"unset: pages apex", "", "https://myanimelogs.pages.dev", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CORS_ALLOWED_ORIGINS", tt.env)

			code, allowOrigin := preflightStatus(tt.origin)

			if tt.want {
				if code != http.StatusNoContent || allowOrigin != tt.origin {
					t.Errorf("origin %q: status = %d, Allow-Origin = %q, want 204 and the origin", tt.origin, code, allowOrigin)
				}
				return
			}
			if code != http.StatusForbidden || allowOrigin != "" {
				t.Errorf("origin %q: status = %d, Allow-Origin = %q, want 403 and empty", tt.origin, code, allowOrigin)
			}
		})
	}
}
