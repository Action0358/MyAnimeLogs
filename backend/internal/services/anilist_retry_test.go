package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"myanimelogs/internal/constants"
)

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		wantDur time.Duration
		wantOK  bool
	}{
		{"秒数指定", "5", 5 * time.Second, true},
		{"前後空白", "  3 ", 3 * time.Second, true},
		{"ゼロ秒", "0", 0, true},
		{"空文字", "", 0, false},
		{"非数値(HTTP-date等)", "Wed, 21 Oct 2026 07:28:00 GMT", 0, false},
		{"負数", "-1", 0, false},
		{"上限超はクランプ", "61", constants.AniListMaxRetryBackoff, true},
		{"極端に大きい値もオーバーフローせずクランプ", "10000000000", constants.AniListMaxRetryBackoff, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, ok := parseRetryAfter(tt.header)
			if ok != tt.wantOK || d != tt.wantDur {
				t.Errorf("parseRetryAfter(%q) = (%v, %v), want (%v, %v)", tt.header, d, ok, tt.wantDur, tt.wantOK)
			}
		})
	}
}

func TestRetryBackoff(t *testing.T) {
	// Retry-Afterが指定された場合はそれを下限とし、+10%ジッタの範囲に収まる
	t.Run("Retry-Afterを尊重", func(t *testing.T) {
		base := 10 * time.Second
		for range 50 {
			d := retryBackoff(1, base, true)
			if d < base || d > base+base/10 {
				t.Fatalf("retryBackoff(1, %v, true) = %v, want in [%v, %v]", base, d, base, base+base/10)
			}
		}
	})

	// Retry-After=0（即時リトライ指示）は0を返す
	t.Run("Retry-Afterゼロは即時", func(t *testing.T) {
		if d := retryBackoff(1, 0, true); d != 0 {
			t.Errorf("retryBackoff(1, 0, true) = %v, want 0", d)
		}
	})

	// Retry-Afterが無い場合は指数バックオフ(2s, 4s, 8s...)
	t.Run("指数バックオフ", func(t *testing.T) {
		expectedBase := []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second}
		for attempt := 1; attempt <= 3; attempt++ {
			want := expectedBase[attempt-1]
			d := retryBackoff(attempt, 0, false)
			if d < want || d > want+want/10 {
				t.Errorf("retryBackoff(%d, 0, false) = %v, want in [%v, %v]", attempt, d, want, want+want/10)
			}
		}
	})

	// 最大バックオフでキャップされる
	t.Run("最大値でキャップ", func(t *testing.T) {
		maxBackoff := constants.AniListMaxRetryBackoff
		d := retryBackoff(20, 0, false) // 2^19 * 2s は遥かに上限超
		if d < maxBackoff || d > maxBackoff+maxBackoff/10 {
			t.Errorf("retryBackoff(20, 0, false) = %v, want capped near %v", d, maxBackoff)
		}
	})
}

// TestExecuteGraphQLQuery_RetryAfterThenSuccess は 429(Retry-After) の後に
// リトライして最終的に成功することを検証する。
func TestExecuteGraphQLQuery_RetryAfterThenSuccess(t *testing.T) {
	var calls int32
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			// 1回目はレート制限。Retry-After=1秒
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		// 2回目以降は成功
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"Media": map[string]interface{}{
					"genres": []string{"Action"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	client := NewAniListClientWithEndpoint(mockServer.URL)

	start := time.Now()
	genres, err := client.GetGenresByMALId(context.Background(), "53065")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected success after retry, got error: %v", err)
	}
	if len(genres) != 1 {
		t.Errorf("expected 1 genre, got %d", len(genres))
	}
	if atomic.LoadInt32(&calls) < 2 {
		t.Errorf("expected at least 2 calls (1 retry), got %d", calls)
	}
	// Retry-After=1秒を尊重しているため、少なくとも約1秒は経過しているはず
	if elapsed < 900*time.Millisecond {
		t.Errorf("expected to wait ~Retry-After (1s), but elapsed only %v", elapsed)
	}
}

// TestExecuteGraphQLQuery_RateLimitedExhausted は 429 が継続した場合に
// リトライ上限で rate limited エラーを返すことを検証する。
func TestExecuteGraphQLQuery_RateLimitedExhausted(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 常にレート制限。Retry-After=0でテストを高速化
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer mockServer.Close()

	client := NewAniListClientWithEndpoint(mockServer.URL)

	_, err := client.GetGenresByMALId(context.Background(), "53065")
	if err == nil {
		t.Fatal("expected rate-limited error, got nil")
	}
}
