package services

import (
	"testing"
	"time"

	"myanimelogs/internal/constants"

	"github.com/stretchr/testify/assert"
)

// fixedNow はテスト用の固定時刻（2025年5月15日 = spring シーズン中）
var fixedNow = time.Date(2025, 5, 15, 12, 0, 0, 0, time.UTC)

func TestIsPastSeasonAt(t *testing.T) {
	tests := []struct {
		name     string
		season   string
		expected bool
	}{
		{"過去シーズン（2016-spring）は true", "2016-spring", true},
		{"過去シーズン（2024-autumn）は true", "2024-autumn", true},
		{"過去シーズン（2025-winter）は true", "2025-winter", true},
		{"現在シーズン（2025-spring）は false", "2025-spring", false},
		{"未来シーズン（2025-summer）は false", "2025-summer", false},
		{"未来シーズン（2099-winter）は false", "2099-winter", false},
		{"空文字列は false", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPastSeasonAt(tt.season, fixedNow)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapTMDBStatus(t *testing.T) {
	tests := []struct {
		name       string
		tmdbStatus string
		expected   string
	}{
		{"Returning Series → airing", "Returning Series", constants.AnimeStatusAiring.String()},
		{"In Production → airing", "In Production", constants.AnimeStatusAiring.String()},
		{"Planned → upcoming", "Planned", constants.AnimeStatusUpcoming.String()},
		{"Ended → finished", "Ended", constants.AnimeStatusFinished.String()},
		{"Canceled → finished", "Canceled", constants.AnimeStatusFinished.String()},
		{"Unknown → finished (default)", "Unknown", constants.AnimeStatusFinished.String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapTMDBStatus(tt.tmdbStatus)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSeasonStringFromTime(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"1月 → winter", time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), "2025-winter"},
		{"3月 → winter", time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC), "2025-winter"},
		{"4月 → spring", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC), "2025-spring"},
		{"6月 → spring", time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC), "2025-spring"},
		{"7月 → summer", time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), "2025-summer"},
		{"9月 → summer", time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC), "2025-summer"},
		{"10月 → autumn", time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC), "2025-autumn"},
		{"12月 → autumn", time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "2025-autumn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSeasonStringFromTime(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetermineStatusBySeasonAt(t *testing.T) {
	futureDate := "2099-01-01"
	pastDate := "2020-01-01"

	tests := []struct {
		name       string
		seasonName string
		releasedOn *string
		expected   string
	}{
		{"released_onが未来 → upcoming", "2025-spring", &futureDate, constants.AnimeStatusUpcoming.String()},
		{"過去シーズン → finished", "2020-winter", &pastDate, constants.AnimeStatusFinished.String()},
		{"未来シーズン → upcoming", "2099-spring", nil, constants.AnimeStatusUpcoming.String()},
		{"空シーズン → finished", "", nil, constants.AnimeStatusFinished.String()},
		{"不正シーズン → finished", "invalid", nil, constants.AnimeStatusFinished.String()},
		{"現在シーズン → airing", "2025-spring", nil, constants.AnimeStatusAiring.String()},
		{"今年の過去シーズン → finished", "2025-winter", nil, constants.AnimeStatusFinished.String()},
		{"今年の未来シーズン → upcoming", "2025-autumn", nil, constants.AnimeStatusUpcoming.String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineStatusBySeasonAt(tt.seasonName, tt.releasedOn, fixedNow)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsPastSeasonAt_SeasonComparisonLogic はissueの核心シナリオを固定時刻で検証する
func TestIsPastSeasonAt_SeasonComparisonLogic(t *testing.T) {
	tests := []struct {
		name     string
		season   string
		expected bool
	}{
		{"このすば1期(2016-winter)は過去シーズン", "2016-winter", true},
		{"このすば2期(2017-winter)は過去シーズン", "2017-winter", true},
		{"現在のシーズン(2025-spring)は過去ではない", "2025-spring", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPastSeasonAt(tt.season, fixedNow)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsPastSeasonAt_SameYearSeasonOrder は同じ年内のシーズン順序が正しく判定されることを検証する
// 固定時刻: 2025年5月15日（springシーズン中）
func TestIsPastSeasonAt_SameYearSeasonOrder(t *testing.T) {
	tests := []struct {
		name     string
		season   string
		expected bool
	}{
		{"2025-winter（1-3月）→ 過去", "2025-winter", true},
		{"2025-spring（4-6月）→ 現在（過去ではない）", "2025-spring", false},
		{"2025-summer（7-9月）→ 未来（過去ではない）", "2025-summer", false},
		{"2025-autumn（10-12月）→ 未来（過去ではない）", "2025-autumn", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPastSeasonAt(tt.season, fixedNow)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsPastSeasonAt_SeasonBoundaries はシーズン境界での判定を検証する
func TestIsPastSeasonAt_SeasonBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		now      time.Time
		season   string
		expected bool
	}{
		{"3月末: 2025-winterは現在", time.Date(2025, 3, 31, 23, 59, 0, 0, time.UTC), "2025-winter", false},
		{"4月初: 2025-winterは過去", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC), "2025-winter", true},
		{"6月末: 2025-springは現在", time.Date(2025, 6, 30, 23, 59, 0, 0, time.UTC), "2025-spring", false},
		{"7月初: 2025-springは過去", time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), "2025-spring", true},
		{"12月末: 2025-autumnは現在", time.Date(2025, 12, 31, 23, 59, 0, 0, time.UTC), "2025-autumn", false},
		{"翌年1月: 2025-autumnは過去", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), "2025-autumn", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPastSeasonAt(tt.season, tt.now)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStatusCorrectionForPastSeason はissueのシナリオを固定時刻で再現する統合テスト
func TestStatusCorrectionForPastSeason(t *testing.T) {
	tests := []struct {
		name           string
		tmdbStatus     string
		season         string
		expectedStatus string
	}{
		{"過去シーズン + Returning Series → finished", "Returning Series", "2016-winter", constants.AnimeStatusFinished.String()},
		{"未来シーズン + Returning Series → airing", "Returning Series", "2099-spring", constants.AnimeStatusAiring.String()},
		{"過去シーズン + Ended → finished", "Ended", "2020-spring", constants.AnimeStatusFinished.String()},
		{"過去シーズン + In Production → finished", "In Production", "2020-spring", constants.AnimeStatusFinished.String()},
		{"シーズン情報なし + Returning Series → airing（補正しない）", "Returning Series", "", constants.AnimeStatusAiring.String()},
		{"現在シーズン + Returning Series → airing", "Returning Series", "2025-spring", constants.AnimeStatusAiring.String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmdbStatus := mapTMDBStatus(tt.tmdbStatus)

			var status string
			if tmdbStatus == constants.AnimeStatusAiring.String() && isPastSeasonAt(tt.season, fixedNow) {
				status = constants.AnimeStatusFinished.String()
			} else {
				status = tmdbStatus
			}

			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}
