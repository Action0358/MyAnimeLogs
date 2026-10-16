package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"myanimelogs/internal/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 再放送を除いた最初の放送枠を、日本時間の日付で返すこと
func TestFetchFirstBroadcastDates_SkipsRebroadcastAndUsesJST(t *testing.T) {
	client := newSeriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"searchWorks":{"nodes":[
			{"annictId":1,"programs":{"nodes":[
				{"startedAt":"2026-09-01T12:00:00Z","rebroadcast":true},
				{"startedAt":"2026-10-06T15:30:00Z","rebroadcast":false}
			]}},
			{"annictId":2,"programs":{"nodes":[]}}
		]}}}`))
	}, nil)

	got, err := client.FetchFirstBroadcastDates([]int{1, 2})
	require.NoError(t, err)
	assert.Equal(t, map[int]string{1: "2026-10-07"}, got, "UTC 15:30 は日本時間で翌日0:30")
}

// 年月日がそろった開始日だけを返すこと
func TestGetStartDatesByMALIds_SkipsPartialDates(t *testing.T) {
	var gotIDs []interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Variables map[string]interface{} `json:"variables"`
		}
		_ = json.Unmarshal(body, &req)
		gotIDs, _ = req.Variables["ids"].([]interface{})
		_, _ = w.Write([]byte(`{"data":{"Page":{"media":[
			{"idMal":100,"startDate":{"year":2026,"month":12,"day":11}},
			{"idMal":200,"startDate":{"year":2026,"month":null,"day":null}}
		]}}}`))
	}))
	defer server.Close()

	got, err := NewAniListClientWithEndpoint(server.URL).GetStartDatesByMALIds(context.Background(), []int{100, 200})
	require.NoError(t, err)
	assert.Len(t, gotIDs, 2)
	assert.Equal(t, map[int]string{100: "2026-12-11"}, got)
}

func TestIsPlausibleReleaseDate(t *testing.T) {
	cases := []struct {
		season, date string
		want         bool
	}{
		{"2026-autumn", "2026-10-07", true},
		{"2026-autumn", "2026-09-01", true},  // 先行放送（シーズン開始の1か月前まで）
		{"2026-autumn", "2026-12-11", true},  // シーズン後半公開の映画
		{"2026-autumn", "2027-02-01", true},  // 公開延期
		{"2026-autumn", "2025-12-03", false}, // 海外での先行配信
		{"2026-autumn", "2026-08-15", false},
		{"2026-winter", "2025-12-05", true},
		{"invalid", "2026-10-07", false},
		{"2026-autumn", "2026-10", false},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, isPlausibleReleaseDate(c.season, c.date), c.season+" "+c.date)
	}
}

// Annictの放送枠 → AniList の順で補い、シーズンとかけ離れた日付は採用しないこと
func TestSyncReleasedOnForCurrentSeason_FallsBackAcrossSources(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()
	db := &database.DB{DB: sqlDB}

	annict := newSeriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"searchWorks":{"nodes":[
			{"annictId":10,"programs":{"nodes":[{"startedAt":"2026-10-06T15:30:00Z","rebroadcast":false}]}},
			{"annictId":20,"programs":{"nodes":[]}},
			{"annictId":30,"programs":{"nodes":[{"startedAt":"2025-12-03T03:00:00Z","rebroadcast":false}]}}
		]}}}`))
	}, db)
	anilist := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"Page":{"media":[
			{"idMal":200,"startDate":{"year":2026,"month":12,"day":11}},
			{"idMal":300,"startDate":{"year":2025,"month":12,"day":3}}
		]}}}`))
	}))
	defer anilist.Close()
	jikanCalls := 0
	jikan := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jikanCalls++
		_, _ = w.Write([]byte(`{"data":{"mal_id":300,"aired":{"from":null}}}`))
	}))
	defer jikan.Close()

	svc := NewAnimeService(db, newTestServiceLogger())
	defer svc.Close()
	svc.SetAnnictClient(annict)
	svc.anilistClient = NewAniListClientWithEndpoint(anilist.URL)
	svc.jikanClient.baseURL = jikan.URL

	mock.ExpectQuery(`SELECT id, title, COALESCE\(mal_anime_id, ''\), COALESCE\(annict_id, 0\), season`).
		WithArgs("2026-autumn", "2026-summer").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "mal_anime_id", "annict_id", "season"}).
			AddRow(1, "TV作品", "100", 10, "2026-autumn").
			AddRow(2, "映画", "200", 20, "2026-autumn").
			AddRow(3, "海外先行作品", "300", 30, "2026-autumn"))
	statusErr := errors.New("skip status in test")
	// 1. Annictの放送枠（日本時間）
	mock.ExpectExec(`UPDATE animes SET released_on = \$1\s+WHERE id = \$2 AND \(released_on IS NULL OR released_on = ''\)`).
		WithArgs("2026-10-07", 1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT season, released_on, status`).WithArgs(1).WillReturnError(statusErr)
	// 2. AniList（作品2のみ。作品3の2025-12-03はどちらの補完元でも採用しない）
	mock.ExpectExec(`UPDATE animes SET released_on`).
		WithArgs("2026-12-11", 2).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT season, released_on, status`).WithArgs(2).WillReturnError(statusErr)

	success, failure, err := svc.SyncReleasedOnForCurrentSeason(context.Background(), "2026-autumn", "2026-summer")
	require.NoError(t, err)
	assert.Equal(t, 2, success)
	assert.Equal(t, 0, failure)
	assert.Equal(t, 1, jikanCalls, "AniListでも採用できなかった作品だけJikanへ問い合わせる")
	assert.NoError(t, mock.ExpectationsWereMet())
}
