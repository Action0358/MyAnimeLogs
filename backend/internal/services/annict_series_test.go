package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"myanimelogs/internal/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GraphQLのモックサーバーへ向けたAnnictClientを返す
func newSeriesTestClient(t *testing.T, handler http.HandlerFunc, db *database.DB) *AnnictClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	original := annictGraphQLEndpoint
	annictGraphQLEndpoint = server.URL
	t.Cleanup(func() { annictGraphQLEndpoint = original })

	client := NewAnnictClientWithDB("test-token", db, newTestServiceLogger())
	client.rateLimitDelay = 0
	return client
}

func TestFetchSeriesNames_ParsesSeriesAndSendsIDs(t *testing.T) {
	var gotQuery string
	client := newSeriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]string
		_ = json.Unmarshal(body, &req)
		gotQuery = req["query"]
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"data":{"searchWorks":{"nodes":[
			{"annictId":11107,"seriesList":{"nodes":[{"name":"銀魂"}]}},
			{"annictId":4157,"seriesList":{"nodes":[]}}
		]}}}`))
	}, nil)

	got, err := client.FetchSeriesNames([]int{11107, 4157, 99999})
	require.NoError(t, err)

	assert.Contains(t, gotQuery, "annictIds: [11107,4157,99999]")
	assert.Equal(t, []string{"銀魂"}, got[11107])
	// 所属シリーズが無い作品は空配列、Annictに無い作品はマップに含まれない
	assert.Equal(t, []string{}, got[4157])
	_, found := got[99999]
	assert.False(t, found)
}

func TestFetchSeriesNames_GraphQLErrorIsReturned(t *testing.T) {
	client := newSeriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errors":[{"message":"boom"}]}`))
	}, nil)

	_, err := client.FetchSeriesNames([]int{1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestFetchSeriesNames_RateLimitIsSentinel(t *testing.T) {
	client := newSeriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}, nil)

	_, err := client.FetchSeriesNames([]int{1})
	assert.True(t, errors.Is(err, ErrRateLimitExceeded))
}

// Annictで見つからなかった作品も空配列で保存し、毎回問い合わせ直さないこと
func TestSyncSeriesNames_SavesFoundAndMissingWorks(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	client := newSeriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"searchWorks":{"nodes":[
			{"annictId":16456,"seriesList":{"nodes":[{"name":"銀魂"}]}}
		]}}}`))
	}, &database.DB{DB: sqlDB})

	mock.ExpectQuery(`SELECT annict_id FROM animes`).
		WithArgs(annictSeriesRecheckDays).
		WillReturnRows(sqlmock.NewRows([]string{"annict_id"}).AddRow(16456).AddRow(99999))
	mock.ExpectExec(`UPDATE animes\s+SET annict_series_names = \$1, annict_series_synced_at = NOW\(\)`).
		WithArgs(pq.Array([]string{"銀魂"}), 16456).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE animes\s+SET annict_series_names = \$1, annict_series_synced_at = NOW\(\)`).
		WithArgs(pq.Array([]string{}), 99999).
		WillReturnResult(sqlmock.NewResult(0, 1))

	saved, err := client.SyncSeriesNames(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, saved)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 1バッチの取得に失敗しても、そのバッチは保存せず（次回再取得）エラーで止めないこと
func TestSyncSeriesNames_FetchFailureLeavesBatchUnsaved(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	client := newSeriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}, &database.DB{DB: sqlDB})

	mock.ExpectQuery(`SELECT annict_id FROM animes`).
		WillReturnRows(sqlmock.NewRows([]string{"annict_id"}).AddRow(1))

	saved, err := client.SyncSeriesNames(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, saved)
	assert.NoError(t, mock.ExpectationsWereMet(), "UPDATEが発行されないこと")
}

// 関連作品のクエリが、Annictシリーズの一致を条件に含むこと。
// シリーズの食い違いによる除外はしない（「名探偵コナン」と「劇場版 名探偵コナン」が切れるため）
func TestGetRelatedSeasons_UsesAnnictSeries(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(expected, actual string) error {
		for _, want := range []string{
			"annict_series_names && COALESCE((SELECT annict_series_names FROM animes WHERE id = $2), '{}')",
		} {
			if !strings.Contains(actual, want) {
				return errors.New("query does not contain: " + want)
			}
		}
		return nil
	})))
	require.NoError(t, err)
	defer sqlDB.Close()

	svc := NewAnimeService(&database.DB{DB: sqlDB}, newTestServiceLogger())
	defer svc.Close()

	mock.ExpectQuery("related").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "season", "cover_image_extra_large", "cover_image_large", "cover_image_medium", "image_url"}))

	svc.getRelatedSeasons(1, "銀魂", nil, nil)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// レーベル・放送枠などの括りは作品群として扱わない
func TestIsFranchiseSeries(t *testing.T) {
	for _, name := range []string{"銀魂", "劇場版 名探偵コナン", "科学アドベンチャー", "Project Itoh"} {
		assert.True(t, isFranchiseSeries(name), name)
	}
	for _, name := range []string{"AnimeFestaオリジナル", "「プチプチ・アニメ」放送作品", "世界名作劇場", "P.A.WORKSの“お仕事\"", "アニメミライ2013", "あにめたまご2017"} {
		assert.False(t, isFranchiseSeries(name), name)
	}
}

func TestFetchSeriesNames_DropsNonFranchiseSeries(t *testing.T) {
	client := newSeriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"searchWorks":{"nodes":[
			{"annictId":1,"seriesList":{"nodes":[{"name":"AnimeFestaオリジナル"}]}},
			{"annictId":2,"seriesList":{"nodes":[{"name":"世界名作劇場"},{"name":"フランダースの犬"}]}}
		]}}}`))
	}, nil)

	got, err := client.FetchSeriesNames([]int{1, 2})
	require.NoError(t, err)
	assert.Equal(t, []string{}, got[1], "取得済みとして空配列で保存されること")
	assert.Equal(t, []string{"フランダースの犬"}, got[2])
}
