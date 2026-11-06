package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGenerateRequestID(t *testing.T) {
	id := generateRequestID()

	assert.NotEmpty(t, id)
	// 16バイト → 32文字のhex文字列
	assert.Len(t, id, 32)
}

func TestGenerateRequestID_Uniqueness(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateRequestID()
		assert.False(t, ids[id], "重複するリクエストIDが生成された: %s", id)
		ids[id] = true
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		header   string
		wantKeep bool
	}{
		{name: "指定なし", header: "", wantKeep: false},
		{name: "英数字とハイフン", header: "abc-123-DEF", wantKeep: true},
		{name: "UUID", header: "550e8400-e29b-41d4-a716-446655440000", wantKeep: true},
		{name: "64文字", header: strings.Repeat("a", 64), wantKeep: true},
		{name: "65文字", header: strings.Repeat("a", 65), wantKeep: false},
		{name: "改行を含む", header: "abc\ninjected", wantKeep: false},
		{name: "空白を含む", header: "abc def", wantKeep: false},
		{name: "記号を含む", header: "abc_def", wantKeep: false},
		{name: "マルチバイト文字", header: "リクエスト", wantKeep: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctxID string
			r := gin.New()
			r.Use(RequestIDMiddleware())
			r.GET("/", func(c *gin.Context) {
				ctxID = c.GetString("request_id")
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				// Header.Set を通さずに入れ、改行を含む値もそのまま渡す
				req.Header["X-Request-Id"] = []string{tt.header}
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			respID := w.Header().Get("X-Request-ID")
			assert.Equal(t, ctxID, respID)
			if tt.wantKeep {
				assert.Equal(t, tt.header, respID)
			} else {
				assert.NotEqual(t, tt.header, respID)
				assert.Regexp(t, `^[0-9a-f]{32}$`, respID)
			}
		})
	}
}
