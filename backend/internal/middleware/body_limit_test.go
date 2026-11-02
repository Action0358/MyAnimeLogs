package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newBodyLimitRouter(maxBytes int64, bodyErr *error) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(BodyLimitMiddleware(maxBytes))
	r.POST("/", func(c *gin.Context) {
		_, *bodyErr = io.ReadAll(c.Request.Body)
		c.Status(http.StatusOK)
	})
	return r
}

func TestBodyLimitMiddleware(t *testing.T) {
	t.Run("上限ちょうどの本文は読める", func(t *testing.T) {
		var bodyErr error
		r := newBodyLimitRouter(10, &bodyErr)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("a", 10))))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, bodyErr)
	})

	t.Run("Content-Lengthが上限を超えると413を返しハンドラーを呼ばない", func(t *testing.T) {
		var bodyErr error
		r := newBodyLimitRouter(10, &bodyErr)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("a", 11))))

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
		assert.Contains(t, w.Body.String(), "REQUEST_TOO_LARGE")
		assert.NoError(t, bodyErr)
	})

	t.Run("チャンク転送で上限を超えると読み取りがMaxBytesErrorになる", func(t *testing.T) {
		var bodyErr error
		r := newBodyLimitRouter(10, &bodyErr)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("a", 11)))
		req.ContentLength = -1
		r.ServeHTTP(httptest.NewRecorder(), req)

		var tooLarge *http.MaxBytesError
		assert.True(t, errors.As(bodyErr, &tooLarge))
	})
}
