package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"myanimelogs/internal/constants"
	"myanimelogs/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// annictRewriteTransport はAnnictへのリクエストをテストサーバーへ向ける
type annictRewriteTransport struct {
	base string
}

func (t *annictRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newURL := strings.Replace(req.URL.String(), annictAPIEndpoint, t.base, 1)
	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, newURL, req.Body)
	if err != nil {
		return nil, err
	}
	newReq.Header = req.Header
	return http.DefaultTransport.RoundTrip(newReq)
}

// newSeasonServer は視聴者数の降順に並んだ作品を per_page ごとに返すサーバー。
// failPage に1以上を渡すと、そのページで500を返す。
func newSeasonServer(t *testing.T, watchers []int, failPage int) (*httptest.Server, *[]int) {
	t.Helper()
	var requested []int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
		requested = append(requested, page)
		if page == failPage {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		start := (page - 1) * perPage
		end := start + perPage
		if end > len(watchers) {
			end = len(watchers)
		}
		resp := models.AnnictWorkResponse{TotalCount: len(watchers)}
		for i := start; i < end; i++ {
			resp.Works = append(resp.Works, models.AnnictWork{ID: i + 1, Title: fmt.Sprintf("work%d", i+1), WatchersCount: watchers[i]})
		}
		if end < len(watchers) {
			next := page + 1
			resp.NextPage = &next
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)
	return server, &requested
}

func newTestAnnictClient(serverURL string) *AnnictClient {
	return &AnnictClient{
		Token:      "test-token",
		HTTPClient: &http.Client{Transport: &annictRewriteTransport{base: serverURL}},
	}
}

// 降順の視聴者数列を作る: high件は下限以上、low件は下限未満
func descendingWatchers(high, low int) []int {
	min := constants.AnnictSeasonMinWatchersCount
	var ws []int
	for i := 0; i < high; i++ {
		ws = append(ws, min+high-i)
	}
	for i := 0; i < low; i++ {
		ws = append(ws, min-1-i%(min-1))
	}
	return ws
}

// 1シーズン50件を超える作品も取り込めること（以前は1ページ目しか取っていなかった）
func TestFetchAllWorksBySeason_FollowsPagesAboveMinWatchers(t *testing.T) {
	server, requested := newSeasonServer(t, descendingWatchers(120, 80), 0)
	client := newTestAnnictClient(server.URL)

	resp, err := client.FetchAllWorksBySeason("2026-winter")

	require.NoError(t, err)
	assert.Len(t, resp.Works, 120)
	for _, w := range resp.Works {
		assert.GreaterOrEqual(t, w.WatchersCount, constants.AnnictSeasonMinWatchersCount)
	}
	// 3ページ目の途中で下限を下回るので、4ページ目は取りに行かない
	assert.Equal(t, []int{1, 2, 3}, *requested)
}

// 放送前の新作は視聴者数が少ないため、上位50件は下限に関係なく取り込む
func TestFetchAllWorksBySeason_KeepsFirstPageRegardlessOfWatchers(t *testing.T) {
	server, requested := newSeasonServer(t, descendingWatchers(10, 70), 0)
	client := newTestAnnictClient(server.URL)

	resp, err := client.FetchAllWorksBySeason("2027-spring")

	require.NoError(t, err)
	assert.Len(t, resp.Works, constants.AnnictAPIPerPage)
	assert.Equal(t, []int{1, 2}, *requested)
}

func TestFetchAllWorksBySeason_StopsWhenNoNextPage(t *testing.T) {
	server, requested := newSeasonServer(t, descendingWatchers(30, 0), 0)
	client := newTestAnnictClient(server.URL)

	resp, err := client.FetchAllWorksBySeason("1990-spring")

	require.NoError(t, err)
	assert.Len(t, resp.Works, 30)
	assert.Equal(t, []int{1}, *requested)
}

// 途中のページで失敗しても、取得済みの分は返す
func TestFetchAllWorksBySeason_ReturnsPartialDataOnLaterPageError(t *testing.T) {
	server, _ := newSeasonServer(t, descendingWatchers(150, 0), 2)
	client := newTestAnnictClient(server.URL)

	resp, err := client.FetchAllWorksBySeason("2026-spring")

	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Works, constants.AnnictAPIPerPage)
}

func TestFetchAllWorksBySeason_ReturnsErrorOnFirstPageError(t *testing.T) {
	server, _ := newSeasonServer(t, descendingWatchers(10, 0), 1)
	client := newTestAnnictClient(server.URL)

	resp, err := client.FetchAllWorksBySeason("2026-spring")

	require.Error(t, err)
	assert.Nil(t, resp)
}
