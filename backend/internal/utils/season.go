package utils

import (
	"regexp"
	"strings"
)

// seasonMapping maps Japanese season names to English
var seasonMapping = map[string]string{
	"春": "spring",
	"夏": "summer",
	"秋": "autumn",
	"冬": "winter",
}

// validDBSeasons is a set of valid season values in database format
var validDBSeasons = map[string]struct{}{
	"spring": {},
	"summer": {},
	"autumn": {},
	"winter": {},
}

// japaneseSeasonPattern matches "2024年春" format
var japaneseSeasonPattern = regexp.MustCompile(`^(\d{4})年(春|夏|秋|冬)$`)

// ConvertSeasonToDBFormat converts Japanese season format to database format
// "2024年春" -> "2024-spring"
// "2024-spring" -> "2024-spring" (already in correct format)
// Returns empty string if format is invalid
func ConvertSeasonToDBFormat(season string) string {
	if season == "" {
		return ""
	}

	// Check if already in DB format (e.g., "2024-spring")
	if isDBFormat(season) {
		return season
	}

	// Try to convert from Japanese format
	matches := japaneseSeasonPattern.FindStringSubmatch(season)
	if matches == nil {
		return "" // Invalid format
	}

	year := matches[1]
	japaneseSeason := matches[2]
	englishSeason, ok := seasonMapping[japaneseSeason]
	if !ok {
		return "" // Unknown season
	}

	return year + "-" + englishSeason
}

// isDBFormat checks if the season is already in database format
func isDBFormat(season string) bool {
	parts := strings.Split(season, "-")
	if len(parts) != 2 {
		return false
	}

	// Check year part (4 digits)
	if len(parts[0]) != 4 {
		return false
	}
	for _, c := range parts[0] {
		if c < '0' || c > '9' {
			return false
		}
	}

	// Check season part using map lookup
	_, ok := validDBSeasons[parts[1]]
	return ok
}
