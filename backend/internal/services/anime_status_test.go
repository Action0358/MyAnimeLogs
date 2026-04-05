package services

import (
	"fmt"
	"testing"
	"time"

	"myanimelogs/internal/constants"

	"github.com/stretchr/testify/assert"
)

func TestIsPastSeason(t *testing.T) {
	tests := []struct {
		name     string
		season   string
		expected bool
	}{
		{
			name:     "過去シーズン（2016-spring）は true",
			season:   "2016-spring",
			expected: true,
		},
		{
			name:     "過去シーズン（2024-autumn）は true",
			season:   "2024-autumn",
			expected: true,
		},
		{
			name:     "未来シーズンは false",
			season:   "2099-winter",
			expected: false,
		},
		{
			name:     "空文字列は false",
			season:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPastSeason(tt.season)
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

func TestDetermineStatusBySeason(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineStatusBySeason(tt.seasonName, tt.releasedOn)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsPastSeason_SeasonComparisonLogic はシーズン文字列の比較ロジックを検証する
// issueの核心: TMDBが "Returning Series" を返しても、過去シーズンの作品は finished になるべき
func TestIsPastSeason_SeasonComparisonLogic(t *testing.T) {
	// 現在のシーズンを取得
	currentSeason := getSeasonStringFromTime(time.Now())

	tests := []struct {
		name     string
		season   string
		expected bool
	}{
		{
			name:     "このすば1期(2016-winter)は過去シーズン",
			season:   "2016-winter",
			expected: true,
		},
		{
			name:     "このすば2期(2017-winter)は過去シーズン",
			season:   "2017-winter",
			expected: true,
		},
		{
			name:     "現在のシーズンは過去ではない",
			season:   currentSeason,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPastSeason(tt.season)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsPastSeason_SameYearSeasonOrder は同じ年内のシーズン順序が正しく判定されることを検証する
// アルファベット順（autumn < spring < summer < winter）と時系列順（winter < spring < summer < autumn）が
// 異なるため、文字列比較ではなく正確な時系列比較が必要
func TestIsPastSeason_SameYearSeasonOrder(t *testing.T) {
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	// 現在の月から各シーズンが過去かどうかを正確に判定
	type seasonInfo struct {
		name       string
		startMonth int
		endMonth   int
	}
	seasons := []seasonInfo{
		{"winter", 1, 3},
		{"spring", 4, 6},
		{"summer", 7, 9},
		{"autumn", 10, 12},
	}

	for _, s := range seasons {
		seasonStr := fmt.Sprintf("%d-%s", currentYear, s.name)
		// 現在月がシーズン終了月より後なら過去シーズン
		expectedPast := currentMonth > s.endMonth

		t.Run(fmt.Sprintf("今年の%sシーズン", s.name), func(t *testing.T) {
			result := isPastSeason(seasonStr)
			assert.Equal(t, expectedPast, result,
				"season=%s, currentMonth=%d, endMonth=%d", seasonStr, currentMonth, s.endMonth)
		})
	}
}

// TestStatusCorrectionForPastSeason はissueのシナリオを再現する統合的なテスト
// TMDBが "Returning Series" を返した場合に、過去シーズンの作品が finished になることを検証
func TestStatusCorrectionForPastSeason(t *testing.T) {
	tests := []struct {
		name           string
		tmdbStatus     string
		season         string
		expectedStatus string
	}{
		{
			name:           "過去シーズン + Returning Series → finished",
			tmdbStatus:     "Returning Series",
			season:         "2016-winter",
			expectedStatus: constants.AnimeStatusFinished.String(),
		},
		{
			name:           "未来シーズン + Returning Series → airing",
			tmdbStatus:     "Returning Series",
			season:         "2099-spring",
			expectedStatus: constants.AnimeStatusAiring.String(),
		},
		{
			name:           "過去シーズン + Ended → finished",
			tmdbStatus:     "Ended",
			season:         "2020-spring",
			expectedStatus: constants.AnimeStatusFinished.String(),
		},
		{
			name:           "過去シーズン + In Production → finished",
			tmdbStatus:     "In Production",
			season:         "2020-spring",
			expectedStatus: constants.AnimeStatusFinished.String(),
		},
		{
			name:           "シーズン情報なし + Returning Series → airing（補正しない）",
			tmdbStatus:     "Returning Series",
			season:         "",
			expectedStatus: constants.AnimeStatusAiring.String(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmdbStatus := mapTMDBStatus(tt.tmdbStatus)

			var status string
			if tmdbStatus == constants.AnimeStatusAiring.String() && isPastSeason(tt.season) {
				status = constants.AnimeStatusFinished.String()
			} else {
				status = tmdbStatus
			}

			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}
