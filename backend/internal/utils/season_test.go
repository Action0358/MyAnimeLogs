package utils

import "testing"

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
