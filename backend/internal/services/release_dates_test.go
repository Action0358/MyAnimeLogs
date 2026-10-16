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
	"github.com/lib/pq"
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
		{"2026-autumn", "2027-10-01", true},  // シーズン開始から1年以内
		{"2026-autumn", "2027-10-02", false}, // 1年超（海外での後年の放送など）
		{"1988-spring", "2017-01-27", false},
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

// 全シーズンの補完: 問い合わせた作品は日付が見つからなくても確認済みにすること
func TestBackfillReleasedOn_SavesAndMarksChecked(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()
	db := &database.DB{DB: sqlDB}

	annict := newSeriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"searchWorks":{"nodes":[
			{"annictId":16456,"programs":{"nodes":[]}},
			{"annictId":2,"programs":{"nodes":[]}}
		]}}}`))
	}, db)
	anilist := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"Page":{"media":[
			{"idMal":62277,"startDate":{"year":2026,"month":2,"day":13}}
		]}}}`))
	}))
	defer anilist.Close()

	svc := NewAnimeService(db, newTestServiceLogger())
	defer svc.Close()
	svc.SetAnnictClient(annict)
	svc.anilistClient = NewAniListClientWithEndpoint(anilist.URL)

	mock.ExpectQuery(`released_on_checked_at IS NULL`).
		WithArgs(releasedOnRecheckDays).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "mal_anime_id", "annict_id", "season"}).
			AddRow(9768, "新劇場版 銀魂 -吉原大炎上-", "62277", 16456, "2026-winter").
			AddRow(5, "日付の無い作品", "500", 2, "2010-spring"))
	mock.ExpectExec(`UPDATE animes SET released_on = \$1`).
		WithArgs("2026-02-13", 9768).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT season, released_on, status`).WithArgs(9768).
		WillReturnError(errors.New("skip status in test"))
	mock.ExpectExec(`UPDATE animes SET released_on_checked_at = NOW\(\) WHERE id = ANY\(\$1\)`).
		WithArgs(pq.Array([]int64{9768, 5})).WillReturnResult(sqlmock.NewResult(0, 2))

	saved, err := svc.BackfillReleasedOn(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, saved)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 問い合わせに失敗したバッチは確認済みにせず、次回再び対象にすること
func TestBackfillReleasedOn_FetchFailureLeavesUnchecked(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()
	db := &database.DB{DB: sqlDB}

	annict := newSeriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}, db)
	anilist := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"Page":{"media":[]}}}`))
	}))
	defer anilist.Close()

	svc := NewAnimeService(db, newTestServiceLogger())
	defer svc.Close()
	svc.SetAnnictClient(annict)
	svc.anilistClient = NewAniListClientWithEndpoint(anilist.URL)

	mock.ExpectQuery(`released_on_checked_at IS NULL`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "mal_anime_id", "annict_id", "season"}).
			AddRow(1, "作品", "100", 10, "2020-spring"))

	saved, err := svc.BackfillReleasedOn(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, saved)
	assert.NoError(t, mock.ExpectationsWereMet(), "確認済みのUPDATEが発行されないこと")
}
