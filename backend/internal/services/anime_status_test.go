package services

import (
	"testing"
	"time"

	"myanimelogs/internal/constants"

	"github.com/stretchr/testify/assert"
)

// fixedNow はテスト用の固定時刻（2025年5月15日 = spring シーズン中）
var fixedNow = time.Date(2025, 5, 15, 12, 0, 0, 0, time.UTC)

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
	tests := []struct {
		name       string
		seasonName string
		releasedOn *string
		expected   string
	}{
		// released_on ベースの判定
		{"released_onが未来 → upcoming", "2025-spring", strPtr("2099-01-01"), constants.AnimeStatusUpcoming.String()},
		{"released_onが過去 + 今期シーズン → airing", "2025-spring", strPtr("2025-01-01"), constants.AnimeStatusAiring.String()},
		{"released_onが今日 + 今期シーズン → airing", "2025-spring", strPtr("2025-05-15"), constants.AnimeStatusAiring.String()},

		// シーズンベースの判定（released_onなし）
		{"過去シーズン → finished", "2020-winter", nil, constants.AnimeStatusFinished.String()},
		{"今年の過去シーズン → finished", "2025-winter", nil, constants.AnimeStatusFinished.String()},
		{"現在シーズン → airing", "2025-spring", nil, constants.AnimeStatusAiring.String()},
		{"今年の未来シーズン → upcoming", "2025-autumn", nil, constants.AnimeStatusUpcoming.String()},
		{"未来シーズン → upcoming", "2099-spring", nil, constants.AnimeStatusUpcoming.String()},

		// エッジケース
		{"空シーズン → finished", "", nil, constants.AnimeStatusFinished.String()},
		{"不正シーズン → finished", "invalid", nil, constants.AnimeStatusFinished.String()},
		{"released_onが空文字 → シーズンベース", "2025-spring", strPtr(""), constants.AnimeStatusAiring.String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineStatusBySeasonAt(tt.seasonName, tt.releasedOn, fixedNow)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDetermineStatusBySeasonAt_SeasonBoundaries はシーズン境界での判定を検証する
func TestDetermineStatusBySeasonAt_SeasonBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		now      time.Time
		season   string
		expected string
	}{
		{"3月末: 2025-winterはairing", time.Date(2025, 3, 31, 23, 59, 0, 0, time.UTC), "2025-winter", constants.AnimeStatusAiring.String()},
		{"4月初: 2025-winterはfinished", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC), "2025-winter", constants.AnimeStatusFinished.String()},
		{"6月末: 2025-springはairing", time.Date(2025, 6, 30, 23, 59, 0, 0, time.UTC), "2025-spring", constants.AnimeStatusAiring.String()},
		{"7月初: 2025-springはfinished", time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), "2025-spring", constants.AnimeStatusFinished.String()},
		{"12月末: 2025-autumnはairing", time.Date(2025, 12, 31, 23, 59, 0, 0, time.UTC), "2025-autumn", constants.AnimeStatusAiring.String()},
		{"翌年1月: 2025-autumnはfinished", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), "2025-autumn", constants.AnimeStatusFinished.String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineStatusBySeasonAt(tt.season, nil, tt.now)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDetermineStatusBySeasonAt_SeriesScenarios はシリーズ作品のシナリオを検証する
// issueの核心: 過去シーズンの作品は常にfinished、今期はairing
func TestDetermineStatusBySeasonAt_SeriesScenarios(t *testing.T) {
	tests := []struct {
		name     string
		season   string
		expected string
	}{
		{"このすば1期(2016-winter) → finished", "2016-winter", constants.AnimeStatusFinished.String()},
		{"このすば2期(2017-winter) → finished", "2017-winter", constants.AnimeStatusFinished.String()},
		{"このすば3期(2024-spring) → finished", "2024-spring", constants.AnimeStatusFinished.String()},
		{"今期の作品(2025-spring) → airing", "2025-spring", constants.AnimeStatusAiring.String()},
		{"来期の作品(2025-summer) → upcoming", "2025-summer", constants.AnimeStatusUpcoming.String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineStatusBySeasonAt(tt.season, nil, fixedNow)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDetermineStatusBySeasonAt_ReleasedOnOverride はreleased_onによる上書きを検証する
func TestDetermineStatusBySeasonAt_ReleasedOnOverride(t *testing.T) {
	tests := []struct {
		name       string
		season     string
		releasedOn string
		expected   string
	}{
		{"released_on未来 → upcoming（シーズンに関係なく）", "2025-spring", "2025-06-15", constants.AnimeStatusUpcoming.String()},
		{"released_on過去 + 今期 → airing", "2025-spring", "2025-01-01", constants.AnimeStatusAiring.String()},
		{"released_on過去 + 過去シーズン → finished", "2024-spring", "2024-01-01", constants.AnimeStatusFinished.String()},
		{"released_on今日 + 今期 → airing", "2025-spring", "2025-05-15", constants.AnimeStatusAiring.String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			releasedOn := tt.releasedOn
			result := determineStatusBySeasonAt(tt.season, &releasedOn, fixedNow)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func strPtr(s string) *string {
	return &s
}
