package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"

	"myanimelogs/internal/common"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const panicSecret = "secret-detail-in-panic"

func newRecoveryEngine(logs *bytes.Buffer) (*gin.Engine, *common.ErrorMonitor) {
	gin.SetMode(gin.TestMode)
	logger := &common.Logger{Logger: slog.New(slog.NewJSONHandler(logs, nil))}
	monitor := common.NewErrorMonitor(logger, 100, time.Minute)

	r := gin.New()
	r.Use(RecoveryMiddleware(monitor))
	return r, monitor
}

func TestRecoveryMiddleware_RespondsWithJSONWithoutPanicDetails(t *testing.T) {
	var logs bytes.Buffer
	r, monitor := newRecoveryEngine(&logs)
	r.GET("/", func(c *gin.Context) { panic(panicSecret) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var body common.ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, common.ErrCodeInternalServer, body.Code)
	assert.NotContains(t, w.Body.String(), panicSecret)

	// ログには調査に必要な内容が残り、認証情報は残らない
	assert.Contains(t, logs.String(), panicSecret)
	assert.Contains(t, logs.String(), "recovery_test.go")
	assert.NotContains(t, logs.String(), "secret-token")
	assert.Equal(t, 1, monitor.GetTotalErrorCount())
}

func TestRecoveryMiddleware_ConnectionStaysUsableAfterPanic(t *testing.T) {
	var logs bytes.Buffer
	r, _ := newRecoveryEngine(&logs)
	r.GET("/panic", func(c *gin.Context) { panic(panicSecret) })
	r.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })

	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/panic")
	require.NoError(t, err, "panic しても接続を切らずに応答を返す")
	resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	resp, err = server.Client().Get(server.URL + "/ok")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRecoveryMiddleware_AbortsConnectionWhenResponseAlreadyStarted(t *testing.T) {
	var logs bytes.Buffer
	r, monitor := newRecoveryEngine(&logs)
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
		_, _ = c.Writer.WriteString(`{"partial":`)
		c.Writer.Flush()
		panic(panicSecret)
	})

	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()

	// 途中までの本文を完全な応答に見せない
	var buf bytes.Buffer
	_, err = buf.ReadFrom(resp.Body)
	assert.Error(t, err, "接続が切れて本文の読み取りが失敗する")
	assert.NotContains(t, buf.String(), common.ErrCodeInternalServer)
	assert.Equal(t, 1, monitor.GetTotalErrorCount())
}

func TestRecoveryMiddleware_BrokenConnectionIsNotRecordedAsPanic(t *testing.T) {
	var logs bytes.Buffer
	r, monitor := newRecoveryEngine(&logs)
	r.GET("/", func(c *gin.Context) {
		panic(&net.OpError{Err: os.NewSyscallError("write", syscall.EPIPE)})
	})

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 0, monitor.GetTotalErrorCount())
}
