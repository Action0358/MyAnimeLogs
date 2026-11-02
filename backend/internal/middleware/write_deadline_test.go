package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testConnTimeout  = 100 * time.Millisecond
	testHandlerSleep = 400 * time.Millisecond
)

// 期限を超えて動くハンドラーを、タイムアウト付きの実サーバーで動かす。
// 戻り値は、クライアント側の結果とハンドラーが見た context のエラー。
func runSlowHandler(t *testing.T, configure func(*http.Server), middlewares ...gin.HandlerFunc) (*http.Response, error, error) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ctxErr := make(chan error, 1)
	r := gin.New()
	r.Use(middlewares...)
	r.POST("/", func(c *gin.Context) {
		_, _ = io.Copy(io.Discard, c.Request.Body)
		select {
		case <-c.Request.Context().Done():
		case <-time.After(testHandlerSleep):
		}
		ctxErr <- c.Request.Context().Err()
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	srv := httptest.NewUnstartedServer(r)
	configure(srv.Config)
	srv.Start()
	defer srv.Close()

	resp, reqErr := srv.Client().Post(srv.URL, "application/json", strings.NewReader(`{}`))
	if resp != nil {
		defer resp.Body.Close()
	}
	return resp, reqErr, <-ctxErr
}

func withWriteTimeout(s *http.Server) { s.WriteTimeout = testConnTimeout }
func withReadTimeout(s *http.Server)  { s.ReadTimeout = testConnTimeout }

func TestClearWriteDeadline(t *testing.T) {
	t.Run("期限を外すとWriteTimeoutを超えても応答が届く", func(t *testing.T) {
		resp, reqErr, ctxErr := runSlowHandler(t, withWriteTimeout, ClearWriteDeadline())

		require.NoError(t, reqErr)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, ctxErr)
	})

	t.Run("期限を外さないとWriteTimeoutを超えた応答は届かない", func(t *testing.T) {
		_, reqErr, _ := runSlowHandler(t, withWriteTimeout)

		assert.Error(t, reqErr)
	})
}

// ReadTimeout を短くできる前提の確認。本文を読み終えた後の処理は止めない。
func TestReadTimeoutDoesNotCancelRunningHandler(t *testing.T) {
	resp, reqErr, ctxErr := runSlowHandler(t, withReadTimeout)

	require.NoError(t, reqErr)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NoError(t, ctxErr)
}
