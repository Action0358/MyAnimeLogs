package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// seasonOrder defines the chronological order of seasons within a year
var seasonOrder = []string{"winter", "spring", "summer", "autumn"}

// seasonIndex maps season names to their chronological order within a year
var seasonIndex = map[string]int{"winter": 0, "spring": 1, "summer": 2, "autumn": 3}

// SeasonForTime returns the season string for a given time (e.g., "2025-autumn")
// This is a pure function that takes time as input for easier testing.
// Shared by the scheduler and AnimeService so season boundaries stay in one place.
func SeasonForTime(t time.Time) string {
	year := t.Year()
	month := t.Month()

	var season string
	switch {
	case month >= 1 && month <= 3:
		season = "winter"
	case month >= 4 && month <= 6:
		season = "spring"
	case month >= 7 && month <= 9:
		season = "summer"
	default:
		season = "autumn"
	}

	return fmt.Sprintf("%d-%s", year, season)
}

// PreviousSeasonForTime returns the previous season string for a given time
// This is a pure function that takes time as input for easier testing
func PreviousSeasonForTime(t time.Time) string {
	year := t.Year()
	month := t.Month()

	var season string
	switch {
	case month >= 1 && month <= 3:
		season = "autumn"
		year--
	case month >= 4 && month <= 6:
		season = "winter"
	case month >= 7 && month <= 9:
		season = "spring"
	default:
		season = "summer"
	}

	return fmt.Sprintf("%d-%s", year, season)
}

// GenerateAllSeasons generates all season strings from startYear-spring to endSeason
// For startYear=1979, it starts from 1979-spring (skipping 1979-winter)
// Returns empty slice if endSeason is before the effective start season (startYear-spring)
func GenerateAllSeasons(startYear, endYear int, endSeason string) []string {
	var result []string

	// Effective start season is spring of startYear (winter is skipped)
	effectiveStartSeason := fmt.Sprintf("%d-spring", startYear)

	// Check if endSeason is before effective start season
	// This handles cases like GenerateAllSeasons(2024, 2024, "2024-winter")
	if IsSeasonBefore(endSeason, effectiveStartSeason) {
		return result
	}

	for year := startYear; year <= endYear; year++ {
		for _, season := range seasonOrder {
			// Skip winter of the start year (e.g., 1979-winter)
			if year == startYear && season == "winter" {
				continue
			}

			seasonStr := fmt.Sprintf("%d-%s", year, season)
			result = append(result, seasonStr)

			// Stop if we've reached the end season
			if seasonStr == endSeason {
				return result
			}
		}
	}

	return result
}

// IsSeasonBefore returns true if season1 is chronologically before season2.
// Malformed input (unparseable "YYYY-season" or an unknown season name) is treated
// as "not before" and returns false, so callers never rely on zero-value fallbacks.
func IsSeasonBefore(season1, season2 string) bool {
	var year1, year2 int
	var seasonName1, seasonName2 string

	if n, _ := fmt.Sscanf(season1, "%d-%s", &year1, &seasonName1); n != 2 {
		return false
	}
	if n, _ := fmt.Sscanf(season2, "%d-%s", &year2, &seasonName2); n != 2 {
		return false
	}

	if year1 != year2 {
		return year1 < year2
	}

	// Compare season order within the same year (unknown names are not comparable)
	idx1, ok1 := seasonIndex[seasonName1]
	idx2, ok2 := seasonIndex[seasonName2]
	if !ok1 || !ok2 {
		return false
	}
	return idx1 < idx2
}

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
	// japaneseSeasonは正規表現により春|夏|秋|冬のいずれかが保証されている
	englishSeason := seasonMapping[japaneseSeason]

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
