package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRateLimitedTransport(limit rate.Limit, burst int) *rateLimitedTransport {
	return &rateLimitedTransport{
		base:    http.DefaultTransport,
		limiter: rate.NewLimiter(limit, burst),
	}
}

func doGet(t *testing.T, client *http.Client, url string, timeout time.Duration) error {
	t.Helper()
	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func TestNewTMDBRateLimitedTransport_Settings(t *testing.T) {
	tr := newTMDBRateLimitedTransport(http.DefaultTransport)

	// 40リクエスト/10秒の90% = 3.6リクエスト/秒
	assert.InDelta(t, 3.6, float64(tr.limiter.Limit()), 0.001)
	assert.Equal(t, 5, tr.limiter.Burst())
}

// 上限を使い切った後は、待ち時間が ctx の期限を超えるなら TMDB を呼ばずにすぐ失敗する。
// 作品IDを次々に開かれても TMDB への問い合わせが上限を超えないことの確認。
func TestRateLimitedTransport_FailsFastWhenExhausted(t *testing.T) {
	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
	}))
	defer server.Close()

	// 最初の2回は通し、その後は1時間に1回しか補充されない
	client := &http.Client{Transport: newTestRateLimitedTransport(rate.Every(time.Hour), 2)}

	assert.NoError(t, doGet(t, client, server.URL, time.Second))
	assert.NoError(t, doGet(t, client, server.URL, time.Second))

	start := time.Now()
	err := doGet(t, client, server.URL, 5*time.Second)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "tmdb rate limit")
	assert.Less(t, time.Since(start), time.Second, "期限まで待たずにすぐ失敗する")
	assert.EqualValues(t, 2, atomic.LoadInt32(&hits), "上限超過のリクエストは TMDB に届かない")
}

// 期限のないバックグラウンド処理（夜間同期など）は、失敗せず順番を待って全件通る。
func TestRateLimitedTransport_WaitsWithoutDeadline(t *testing.T) {
	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
	}))
	defer server.Close()

	// 50リクエスト/秒（20msおき）・瞬間の許容は1
	client := &http.Client{Transport: newTestRateLimitedTransport(50, 1)}

	start := time.Now()
	for i := 0; i < 3; i++ {
		assert.NoError(t, doGet(t, client, server.URL, 0))
	}

	assert.EqualValues(t, 3, atomic.LoadInt32(&hits))
	assert.GreaterOrEqual(t, time.Since(start), 30*time.Millisecond, "2回ぶんの待ちが入る")
}

// 本番で使う NewTMDBClient が、実際にレートリミッター付きの通信を使っていること。
// ここが外れると、上のテストが通っていても TMDB の上限は守られない。
func TestNewTMDBClient_UsesRateLimitedTransport(t *testing.T) {
	c := NewTMDBClient(nil)

	_, ok := c.httpClient.Transport.(*rateLimitedTransport)

	assert.True(t, ok)
}
