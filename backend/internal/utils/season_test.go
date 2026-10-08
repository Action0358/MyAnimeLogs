package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertSeasonToDBFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Japanese format conversion
		{"Japanese spring", "2024年春", "2024-spring"},
		{"Japanese summer", "2024年夏", "2024-summer"},
		{"Japanese autumn", "2024年秋", "2024-autumn"},
		{"Japanese winter", "2024年冬", "2024-winter"},
		{"Japanese 2025 spring", "2025年春", "2025-spring"},

		// Already in DB format (passthrough)
		{"DB format spring", "2024-spring", "2024-spring"},
		{"DB format summer", "2024-summer", "2024-summer"},
		{"DB format autumn", "2024-autumn", "2024-autumn"},
		{"DB format winter", "2024-winter", "2024-winter"},

		// Invalid formats (return empty)
		{"Empty string", "", ""},
		{"Invalid format", "2024春", ""},
		{"Invalid season", "2024年abc", ""},
		{"Invalid year", "abcd年春", ""},
		{"Random string", "random", ""},
		{"Partial match", "2024-", ""},
		{"Invalid DB format", "2024-fall", ""}, // fall is not valid, should be autumn
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertSeasonToDBFormat(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertSeasonToDBFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSeasonForTime(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"January - Winter", time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), "2025-winter"},
		{"March - Winter", time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC), "2025-winter"},
		{"April - Spring", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC), "2025-spring"},
		{"June - Spring", time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC), "2025-spring"},
		{"July - Summer", time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), "2025-summer"},
		{"September - Summer", time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC), "2025-summer"},
		{"October - Autumn", time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC), "2025-autumn"},
		{"December - Autumn", time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "2025-autumn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, SeasonForTime(tt.time))
		})
	}
}

func TestPreviousSeasonForTime(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"January - Previous Autumn", time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), "2024-autumn"},
		{"April - Previous Winter", time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC), "2025-winter"},
		{"July - Previous Spring", time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC), "2025-spring"},
		{"October - Previous Summer", time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC), "2025-summer"},
		{"Year boundary", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), "2024-autumn"},
		{"December 31", time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC), "2025-summer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, PreviousSeasonForTime(tt.time))
		})
	}
}

func TestNextSeasonForTime(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"January - Next Spring", time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), "2025-spring"},
		{"April - Next Summer", time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC), "2025-summer"},
		{"July - Next Autumn", time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC), "2025-autumn"},
		{"October - Next Winter (year boundary)", time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC), "2026-winter"},
		{"December 31", time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC), "2026-winter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, NextSeasonForTime(tt.time))
		})
	}
}

func TestGenerateAllSeasons(t *testing.T) {
	tests := []struct {
		name      string
		startYear int
		endYear   int
		endSeason string
		expected  []string
	}{
		{
			name:      "Single year - spring to autumn",
			startYear: 2024,
			endYear:   2024,
			endSeason: "2024-autumn",
			expected:  []string{"2024-spring", "2024-summer", "2024-autumn"},
		},
		{
			name:      "Single year - spring to summer",
			startYear: 2024,
			endYear:   2024,
			endSeason: "2024-summer",
			expected:  []string{"2024-spring", "2024-summer"},
		},
		{
			name:      "Single year - spring only",
			startYear: 2024,
			endYear:   2024,
			endSeason: "2024-spring",
			expected:  []string{"2024-spring"},
		},
		{
			name:      "Edge case - endSeason is winter of startYear (before effective start)",
			startYear: 2024,
			endYear:   2024,
			endSeason: "2024-winter",
			expected:  nil,
		},
		{
			name:      "Cross year - 2024 spring to 2025 winter",
			startYear: 2024,
			endYear:   2025,
			endSeason: "2025-winter",
			expected:  []string{"2024-spring", "2024-summer", "2024-autumn", "2025-winter"},
		},
		{
			name:      "Multiple years - 1979 to 1980 autumn",
			startYear: 1979,
			endYear:   1980,
			endSeason: "1980-autumn",
			expected: []string{
				"1979-spring", "1979-summer", "1979-autumn",
				"1980-winter", "1980-spring", "1980-summer", "1980-autumn",
			},
		},
		{
			name:      "Edge case - endSeason before startYear",
			startYear: 2024,
			endYear:   2024,
			endSeason: "2023-autumn",
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GenerateAllSeasons(tt.startYear, tt.endYear, tt.endSeason))
		})
	}
}

func TestIsSeasonBefore(t *testing.T) {
	tests := []struct {
		name     string
		season1  string
		season2  string
		expected bool
	}{
		{"Same season", "2024-spring", "2024-spring", false},
		{"Earlier year", "2023-autumn", "2024-spring", true},
		{"Later year", "2025-winter", "2024-autumn", false},
		{"Same year - winter before spring", "2024-winter", "2024-spring", true},
		{"Same year - spring before summer", "2024-spring", "2024-summer", true},
		{"Same year - summer before autumn", "2024-summer", "2024-autumn", true},
		{"Same year - autumn after summer", "2024-autumn", "2024-summer", false},

		// Malformed / unknown input must return false (no zero-value fallback)
		{"season1 unparseable", "garbage", "2024-spring", false},
		{"season2 unparseable", "2024-spring", "garbage", false},
		{"non-numeric year", "abcd-spring", "2024-spring", false},
		{"season1 unknown name", "2024-fall", "2024-spring", false},
		{"season2 unknown name", "2024-spring", "2024-fall", false},
		// Unknown season name must return false even when years differ
		// (regression: previously the year comparison short-circuited validation)
		{"season1 unknown name diff year", "2023-fall", "2024-spring", false},
		{"season2 unknown name diff year", "2023-spring", "2024-fall", false},
		{"both empty", "", "", false},
		{"missing season name", "2024-", "2024-spring", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsSeasonBefore(tt.season1, tt.season2))
		})
	}
}

func TestIsDBFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid spring", "2024-spring", true},
		{"Valid summer", "2024-summer", true},
		{"Valid autumn", "2024-autumn", true},
		{"Valid winter", "2024-winter", true},
		{"Invalid fall", "2024-fall", false},
		{"Japanese format", "2024年春", false},
		{"No hyphen", "2024spring", false},
		{"Empty", "", false},
		{"Short year", "24-spring", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDBFormat(tt.input)
			if result != tt.expected {
				t.Errorf("isDBFormat(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
