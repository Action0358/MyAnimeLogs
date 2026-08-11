package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"myanimelogs/internal/common"
)

const (
	tmdbBaseURL = "https://api.themoviedb.org/3"
	tmdbTimeout = 10 * time.Second
	// tmdbFallbackTimeout bounds the opposite-media-type retry in
	// GetJapaneseWatchProvidersWithFallback. See that function for why.
	tmdbFallbackTimeout = 2 * time.Second
	tmdbImageBase       = "https://image.tmdb.org/t/p"
)

// TMDBClient handles communication with TMDB API
type TMDBClient struct {
	apiKey     string
	httpClient *http.Client
	logger     *common.Logger
}

// TMDBSearchResult represents a TV show search result from TMDB
type TMDBSearchResult struct {
	Page         int          `json:"page"`
	Results      []TMDBTVShow `json:"results"`
	TotalPages   int          `json:"total_pages"`
	TotalResults int          `json:"total_results"`
}

// TMDBMovieSearchResult represents a movie search result from TMDB
type TMDBMovieSearchResult struct {
	Page         int         `json:"page"`
	Results      []TMDBMovie `json:"results"`
	TotalPages   int         `json:"total_pages"`
	TotalResults int         `json:"total_results"`
}

// TMDBMovie represents a movie from TMDB
type TMDBMovie struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	PosterPath       string  `json:"poster_path"`
	BackdropPath     string  `json:"backdrop_path"`
	ReleaseDate      string  `json:"release_date"`
	OriginalLanguage string  `json:"original_language"`
	Popularity       float64 `json:"popularity"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	GenreIDs         []int   `json:"genre_ids"`
}

// TMDBTVShow represents a TV show from TMDB
type TMDBTVShow struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	OriginalName     string   `json:"original_name"`
	Overview         string   `json:"overview"`
	PosterPath       string   `json:"poster_path"`
	BackdropPath     string   `json:"backdrop_path"`
	FirstAirDate     string   `json:"first_air_date"`
	OriginCountry    []string `json:"origin_country"`
	OriginalLanguage string   `json:"original_language"`
	Popularity       float64  `json:"popularity"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
	GenreIDs         []int    `json:"genre_ids"`
}

// TMDBTVDetails represents detailed TV show information from TMDB
type TMDBTVDetails struct {
	ID               int           `json:"id"`
	Name             string        `json:"name"`
	OriginalName     string        `json:"original_name"`
	Overview         string        `json:"overview"`
	PosterPath       string        `json:"poster_path"`
	BackdropPath     string        `json:"backdrop_path"`
	FirstAirDate     string        `json:"first_air_date"`
	LastAirDate      string        `json:"last_air_date"`
	Homepage         string        `json:"homepage"`
	InProduction     bool          `json:"in_production"`
	NumberOfEpisodes int           `json:"number_of_episodes"`
	NumberOfSeasons  int           `json:"number_of_seasons"`
	OriginCountry    []string      `json:"origin_country"`
	OriginalLanguage string        `json:"original_language"`
	Popularity       float64       `json:"popularity"`
	VoteAverage      float64       `json:"vote_average"`
	VoteCount        int           `json:"vote_count"`
	Status           string        `json:"status"`
	Tagline          string        `json:"tagline"`
	Type             string        `json:"type"`
	Genres           []TMDBGenre   `json:"genres"`
	Networks         []TMDBNetwork `json:"networks"`
}

// TMDBGenre represents a genre from TMDB
type TMDBGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TMDBNetwork represents a network from TMDB
type TMDBNetwork struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	LogoPath      string `json:"logo_path"`
	OriginCountry string `json:"origin_country"`
}

// TMDBWatchProviders represents watch provider information from TMDB
type TMDBWatchProviders struct {
	ID      int                          `json:"id"`
	Results map[string]TMDBCountryResult `json:"results"`
}

// TMDBCountryResult represents watch providers for a specific country
type TMDBCountryResult struct {
	Link     string         `json:"link"`
	Flatrate []TMDBProvider `json:"flatrate"`
	Rent     []TMDBProvider `json:"rent"`
	Buy      []TMDBProvider `json:"buy"`
}

// TMDBProvider represents a streaming provider
type TMDBProvider struct {
	LogoPath        string `json:"logo_path"`
	ProviderID      int    `json:"provider_id"`
	ProviderName    string `json:"provider_name"`
	DisplayPriority int    `json:"display_priority"`
}

// NewTMDBClient creates a new TMDB API client
func NewTMDBClient(logger *common.Logger) *TMDBClient {
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		if logger != nil {
			logger.Warn("TMDB_API_KEY not set, TMDB features will be disabled")
		}
	}

	return &TMDBClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: tmdbTimeout,
		},
		logger: logger,
	}
}

// IsEnabled returns true if TMDB API is configured
func (c *TMDBClient) IsEnabled() bool {
	return c.apiKey != ""
}

// SearchTV searches for TV shows by title
// language: ISO 639-1 code (e.g., "ja" for Japanese, "en" for English)
func (c *TMDBClient) SearchTV(ctx context.Context, query string, language string) (*TMDBSearchResult, error) {
	if !c.IsEnabled() {
		return nil, fmt.Errorf("TMDB API is not configured")
	}

	if language == "" {
		language = "ja-JP"
	}

	endpoint := fmt.Sprintf("%s/search/tv", tmdbBaseURL)
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("query", query)
	params.Set("language", language)
	params.Set("include_adult", "false")

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", c.redactAPIKey(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	var result TMDBSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// SearchMovie searches for movies by title
// language: ISO 639-1 code (e.g., "ja" for Japanese, "en" for English)
func (c *TMDBClient) SearchMovie(ctx context.Context, query string, language string) (*TMDBMovieSearchResult, error) {
	if !c.IsEnabled() {
		return nil, fmt.Errorf("TMDB API is not configured")
	}

	if language == "" {
		language = "ja-JP"
	}

	endpoint := fmt.Sprintf("%s/search/movie", tmdbBaseURL)
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("query", query)
	params.Set("language", language)
	params.Set("include_adult", "false")

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", c.redactAPIKey(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	var result TMDBMovieSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetMovieDetails gets detailed information about a movie
func (c *TMDBClient) GetMovieDetails(ctx context.Context, tmdbID int, language string) (*TMDBMovie, error) {
	if !c.IsEnabled() {
		return nil, fmt.Errorf("TMDB API is not configured")
	}

	if language == "" {
		language = "ja-JP"
	}

	endpoint := fmt.Sprintf("%s/movie/%d", tmdbBaseURL, tmdbID)
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("language", language)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", c.redactAPIKey(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	var result TMDBMovie
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetTVDetails gets detailed information about a TV show
// language: ISO 639-1 code (e.g., "ja-JP" for Japanese)
func (c *TMDBClient) GetTVDetails(ctx context.Context, tmdbID int, language string) (*TMDBTVDetails, error) {
	if !c.IsEnabled() {
		return nil, fmt.Errorf("TMDB API is not configured")
	}

	if language == "" {
		language = "ja-JP"
	}

	endpoint := fmt.Sprintf("%s/tv/%d", tmdbBaseURL, tmdbID)
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("language", language)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", c.redactAPIKey(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	var result TMDBTVDetails
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetWatchProviders gets streaming provider information for a TV show or movie
// mediaType: "tv" or "movie"
// region: ISO 3166-1 code (e.g., "JP" for Japan)
func (c *TMDBClient) GetWatchProviders(ctx context.Context, tmdbID int, mediaType string) (*TMDBWatchProviders, error) {
	if !c.IsEnabled() {
		return nil, fmt.Errorf("TMDB API is not configured")
	}

	// Default to TV if not specified
	if mediaType == "" {
		mediaType = "tv"
	}

	endpoint := fmt.Sprintf("%s/%s/%d/watch/providers", tmdbBaseURL, mediaType, tmdbID)
	params := url.Values{}
	params.Set("api_key", c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", c.redactAPIKey(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	var result TMDBWatchProviders
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// SearchTVByTitle searches for a TV show and returns the best match
// Prioritizes Japanese anime (origin_country contains "JP")
// Tries multiple search strategies: original title, normalized title (without season suffix)
func (c *TMDBClient) SearchTVByTitle(ctx context.Context, title string) (*TMDBTVShow, error) {
	// Try original title first. Propagate real API/config errors immediately so they are
	// not masked by the generic "no results" error below (a successful-but-empty response
	// returns no error and falls through to the normalized search).
	result, err := c.SearchTV(ctx, title, "ja-JP")
	if err != nil {
		return nil, err
	}
	if len(result.Results) > 0 {
		if show := c.findBestMatch(result.Results, title); show != nil {
			return show, nil
		}
	}

	// Try normalized title (remove season suffixes like "第3期", "Season 2", etc.).
	// Skip when normalization is a no-op or empties the title to avoid a redundant /
	// invalid (empty-query) API call.
	normalizedTitle := NormalizeAnimeTitle(title)
	if normalizedTitle == "" || normalizedTitle == title {
		return nil, fmt.Errorf("no results found for title: %s", title)
	}

	result, err = c.SearchTV(ctx, normalizedTitle, "ja-JP")
	if err != nil {
		return nil, err
	}
	if len(result.Results) > 0 {
		if show := c.findBestMatch(result.Results, normalizedTitle); show != nil {
			return show, nil
		}
	}

	return nil, fmt.Errorf("no results found for title: %s", title)
}

// findBestMatch finds the best matching TV show from search results.
// It prioritizes Japanese anime (origin_country contains "JP"), but only accepts a
// result whose title is sufficiently similar to the query, so an unrelated work is not
// returned just because it happened to be first in the search results.
func (c *TMDBClient) findBestMatch(results []TMDBTVShow, title string) *TMDBTVShow {
	query := normalizeForComparison(title)

	// First pass: prefer a Japanese-origin result with a similar title.
	for i := range results {
		if slices.Contains(results[i].OriginCountry, "JP") &&
			titlesSimilar(query, results[i].Name, results[i].OriginalName) {
			return &results[i]
		}
	}

	// Second pass: any result with a similar title.
	for i := range results {
		if titlesSimilar(query, results[i].Name, results[i].OriginalName) {
			return &results[i]
		}
	}

	// No sufficiently similar match; returning nil lets the caller fall back rather
	// than surface a wrong work's data.
	return nil
}

// titleSimilarityThreshold is the minimum Levenshtein-based similarity ratio (0..1)
// required to accept a fuzzy title match.
const titleSimilarityThreshold = 0.5

// minContainmentRunes is the minimum length (in runes) of the shorter string for a
// substring-containment match to count. Without this, very short normalized strings
// (e.g. "no", "in") would falsely match any title that happens to contain them.
const minContainmentRunes = 3

// normalizeForComparison lowercases a title and strips everything except letters and
// numbers (including spaces and punctuation) so titles can be compared loosely.
func normalizeForComparison(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// titlesSimilar reports whether the normalized query matches either of the candidate
// titles by containment (either direction) or Levenshtein similarity.
func titlesSimilar(normalizedQuery, name, originalName string) bool {
	if normalizedQuery == "" {
		return false
	}
	return titleSimilar(normalizedQuery, name) || titleSimilar(normalizedQuery, originalName)
}

// titleSimilar reports whether a single candidate title matches the normalized query
// by containment (either direction) or Levenshtein similarity.
func titleSimilar(normalizedQuery, candidate string) bool {
	c := normalizeForComparison(candidate)
	if c == "" {
		return false
	}
	if containmentMatch(normalizedQuery, c) {
		return true
	}
	return similarityRatio(normalizedQuery, c) >= titleSimilarityThreshold
}

// containmentMatch reports whether the shorter of the two strings is a substring of the
// longer one, requiring the shorter string to be at least minContainmentRunes long to
// avoid false positives from very short strings.
func containmentMatch(a, b string) bool {
	shorter, longer := a, b
	if utf8.RuneCountInString(a) > utf8.RuneCountInString(b) {
		shorter, longer = b, a
	}
	if utf8.RuneCountInString(shorter) < minContainmentRunes {
		return false
	}
	return strings.Contains(longer, shorter)
}

// similarityRatio returns a 0..1 similarity score based on the Levenshtein distance.
func similarityRatio(a, b string) float64 {
	ra, rb := []rune(a), []rune(b)
	maxLen := max(len(ra), len(rb))
	if maxLen == 0 {
		return 1
	}
	return 1 - float64(levenshtein(ra, rb))/float64(maxLen)
}

// levenshtein computes the edit distance between two rune slices using a single
// reused row, keeping space at O(min(len(a), len(b))).
func levenshtein(a, b []rune) int {
	// Ensure b is the shorter slice so the row stays small.
	if len(a) < len(b) {
		a, b = b, a
	}
	if len(b) == 0 {
		return len(a)
	}

	row := make([]int, len(b)+1)
	for j := range row {
		row[j] = j
	}
	for i := 1; i <= len(a); i++ {
		prevDiag := row[0] // distance for (i-1, j-1)
		row[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur := min(row[j]+1, min(row[j-1]+1, prevDiag+cost))
			prevDiag = row[j]
			row[j] = cur
		}
	}
	return row[len(b)]
}

// Pre-compiled regex patterns for anime title normalization (performance optimization)
var animeTitlePatterns = []*regexp.Regexp{
	// Japanese season patterns (Arabic numerals)
	regexp.MustCompile(` 第\d+期.*`),    // " 第3期" and anything after
	regexp.MustCompile(`第\d+期.*`),     // "第3期" (no space)
	regexp.MustCompile(` 第\d+シーズン.*`), // " 第2シーズン"
	regexp.MustCompile(`第\d+シーズン.*`),  // "第2シーズン"
	regexp.MustCompile(` 第\d+クール.*`),  // " 第2クール"
	regexp.MustCompile(`第\d+クール.*`),   // "第2クール"
	// Japanese season patterns (Kanji numerals)
	regexp.MustCompile(` 第[一二三四五六七八九十百]+期.*`),    // " 第二期"
	regexp.MustCompile(`第[一二三四五六七八九十百]+期.*`),     // "第二期" (no space)
	regexp.MustCompile(` 第[一二三四五六七八九十百]+シーズン.*`), // " 第二シーズン"
	regexp.MustCompile(`第[一二三四五六七八九十百]+シーズン.*`),  // "第二シーズン"
	// English season patterns (case-insensitive, greedy: removes everything after match)
	// (?i) is required because titles use mixed casing such as "4th season" / "Final Season".
	regexp.MustCompile(`(?i) season \d+.*`),                // " Season 3 Part.2", " season 3" etc.
	regexp.MustCompile(`(?i) the final season.*`),          // " The Final Season Part 2"
	regexp.MustCompile(`(?i) final season.*`),              // " Final Season"
	regexp.MustCompile(`(?i) \d+(?:st|nd|rd|th) season.*`), // " 2nd Season", " 4th season 喪失編"
	regexp.MustCompile(`(?i) part[.\s]*\d+\s*$`),           // " Part 2", " Part.2" (allow trailing spaces)
	// Roman numerals (uppercase only to avoid over-matching lowercase i/v/x)
	regexp.MustCompile(` [IVX]+\s*$`), // " II", " III", etc. (allow trailing spaces)
}

// NormalizeAnimeTitle removes season suffixes and other variations from anime titles
// Examples:
// - "ワンパンマン 第3期" -> "ワンパンマン"
// - "葬送のフリーレン 第2期" -> "葬送のフリーレン"
// - "【推しの子】第3期" -> "【推しの子】"
// - "進撃の巨人 The Final Season Part 2" -> "進撃の巨人"
func NormalizeAnimeTitle(title string) string {
	result := title
	// Apply all patterns to handle titles with multiple suffixes
	// e.g., "進撃の巨人 The Final Season Part 2" -> "進撃の巨人"
	for _, re := range animeTitlePatterns {
		result = re.ReplaceAllString(result, "")
	}
	return strings.TrimSpace(result)
}

// GetJapaneseOverview gets the Japanese overview for a TV show
// Falls back to English if Japanese is not available
func (c *TMDBClient) GetJapaneseOverview(ctx context.Context, tmdbID int) (string, error) {
	// First try Japanese
	details, err := c.GetTVDetails(ctx, tmdbID, "ja-JP")
	if err != nil {
		return "", err
	}

	if details.Overview != "" {
		return details.Overview, nil
	}

	// Fall back to English if Japanese overview is empty
	details, err = c.GetTVDetails(ctx, tmdbID, "en-US")
	if err != nil {
		return "", err
	}

	return details.Overview, nil
}

// FindTMDBIDByTitle searches for a TV show and returns its TMDB ID
func (c *TMDBClient) FindTMDBIDByTitle(ctx context.Context, title string) (int, error) {
	show, err := c.SearchTVByTitle(ctx, title)
	if err != nil {
		return 0, err
	}
	return show.ID, nil
}

// GetOverviewByTitle searches for a TV show by title and returns its Japanese overview
func (c *TMDBClient) GetOverviewByTitle(ctx context.Context, title string) (string, int, error) {
	show, err := c.SearchTVByTitle(ctx, title)
	if err != nil {
		return "", 0, err
	}

	// If the search result already has a Japanese overview, use it
	if show.Overview != "" {
		return show.Overview, show.ID, nil
	}

	// Otherwise, fetch the details to get the overview
	overview, err := c.GetJapaneseOverview(ctx, show.ID)
	if err != nil {
		return "", show.ID, err
	}

	return overview, show.ID, nil
}

// GetProviderLogoURL returns the full URL for a provider logo
func GetProviderLogoURL(logoPath string, size string) string {
	if logoPath == "" {
		return ""
	}
	if size == "" {
		size = "w92"
	}
	return fmt.Sprintf("%s/%s%s", tmdbImageBase, size, logoPath)
}

// redactAPIKey strips the API key from an error message. Transport errors wrap a
// *url.Error that carries the full request URL, including the api_key query
// parameter, which would otherwise be written to the logs verbatim.
func (c *TMDBClient) redactAPIKey(err error) error {
	if err == nil || c.apiKey == "" {
		return err
	}
	msg := err.Error()
	redacted := strings.ReplaceAll(msg, c.apiKey, "REDACTED")
	if redacted == msg {
		return err
	}
	return errors.New(redacted)
}

// GetJapaneseWatchProviders gets streaming providers available in Japan
// mediaType: "tv" or "movie"
func (c *TMDBClient) GetJapaneseWatchProviders(ctx context.Context, tmdbID int, mediaType string) ([]TMDBProvider, error) {
	providers, err := c.GetWatchProviders(ctx, tmdbID, mediaType)
	if err != nil {
		return nil, err
	}

	jpProviders, ok := providers.Results["JP"]
	if !ok {
		return nil, nil // No providers in Japan
	}

	// Return flatrate (subscription) providers
	return jpProviders.Flatrate, nil
}

// GetJapaneseWatchProvidersWithFallback searches for watch providers using the specified mediaType,
// and falls back to the other type if no providers are found.
func (c *TMDBClient) GetJapaneseWatchProvidersWithFallback(ctx context.Context, tmdbID int, mediaType string) ([]TMDBProvider, error) {
	// Try the specified mediaType first
	providers, err := c.GetJapaneseWatchProviders(ctx, tmdbID, mediaType)
	if err == nil && len(providers) > 0 {
		return providers, nil
	}
	firstErr := err

	// Fallback to the other mediaType
	fallbackType := "movie"
	if mediaType == "movie" {
		fallbackType = "tv"
	}

	// The fallback queries an ID against the opposite media type, which TMDB may never
	// answer. Left unbounded it held anime detail responses open until the caller's
	// deadline (~10s). Give it a short budget: the fallback is a best-effort retry, so
	// timing out simply means falling back to whatever the first call produced.
	fallbackCtx, cancel := context.WithTimeout(ctx, tmdbFallbackTimeout)
	defer cancel()

	fallbackProviders, fallbackErr := c.GetJapaneseWatchProviders(fallbackCtx, tmdbID, fallbackType)
	if fallbackErr != nil {
		if firstErr != nil {
			// Neither media type produced a usable answer.
			return nil, firstErr
		}
		// The first call succeeded but found nothing; keep that as the answer.
		return providers, nil
	}
	return fallbackProviders, nil
}

// FindTMDBID finds a TMDB ID and media type ("tv" or "movie") by title search.
// mediaTypeHint: "MOVIE", "TV", or "" (empty for unknown) - from AniList format.
//
// Note: the malID argument is retained for API compatibility but is no longer used for a
// TMDB /find lookup. TMDB's /find endpoint does not support "myanimelist_id" as an
// external_source (only imdb_id / tvdb_id / wikidata_id etc.), so that lookup always
// returned 0 results and wasted an API call. Title search is the only working path.
func (c *TMDBClient) FindTMDBID(ctx context.Context, malID string, title string, mediaTypeHint string) (int, string, error) {
	_ = malID // unused: see note above

	// Search by title based on media type hint.
	// If hint is MOVIE, search movie first; otherwise search TV first
	if mediaTypeHint == "MOVIE" {
		// Search movie first for movies
		movie, err := c.SearchMovieByTitle(ctx, title)
		if err == nil && movie != nil {
			return movie.ID, "movie", nil
		}
		// Fall back to TV search
		show, err := c.SearchTVByTitle(ctx, title)
		if err == nil && show != nil {
			return show.ID, "tv", nil
		}
	} else {
		// Search TV first for TV shows (default)
		show, err := c.SearchTVByTitle(ctx, title)
		if err == nil && show != nil {
			return show.ID, "tv", nil
		}
		// Fall back to movie search
		movie, err := c.SearchMovieByTitle(ctx, title)
		if err == nil && movie != nil {
			return movie.ID, "movie", nil
		}
	}

	return 0, "", fmt.Errorf("no TV show or movie found for title: %s", title)
}

// SearchMovieByTitle searches for a movie and returns the best match
// Prioritizes Japanese anime movies (original_language is "ja")
func (c *TMDBClient) SearchMovieByTitle(ctx context.Context, title string) (*TMDBMovie, error) {
	// Try original title first. Propagate real API/config errors immediately rather than
	// swallowing them (an empty-but-successful response falls through to normalized search).
	result, err := c.SearchMovie(ctx, title, "ja-JP")
	if err != nil {
		return nil, err
	}
	if len(result.Results) > 0 {
		if movie := c.findBestMovieMatch(result.Results, title); movie != nil {
			return movie, nil
		}
	}

	// Try normalized title. Skip when normalization is a no-op or empties the title to
	// avoid a redundant / invalid (empty-query) API call.
	normalizedTitle := NormalizeAnimeTitle(title)
	if normalizedTitle == "" || normalizedTitle == title {
		return nil, fmt.Errorf("no movie found for title: %s", title)
	}

	result, err = c.SearchMovie(ctx, normalizedTitle, "ja-JP")
	if err != nil {
		return nil, err
	}
	if len(result.Results) > 0 {
		if movie := c.findBestMovieMatch(result.Results, normalizedTitle); movie != nil {
			return movie, nil
		}
	}

	return nil, fmt.Errorf("no movie found for title: %s", title)
}

// findBestMovieMatch finds the best matching movie from search results.
// It prioritizes Japanese anime movies (original_language is "ja"), but only accepts a
// result whose title is sufficiently similar to the query, mirroring findBestMatch for
// TV shows so an unrelated movie is not returned just because it was first.
func (c *TMDBClient) findBestMovieMatch(results []TMDBMovie, title string) *TMDBMovie {
	query := normalizeForComparison(title)

	// First pass: prefer a Japanese-language movie with a similar title.
	for i := range results {
		if results[i].OriginalLanguage == "ja" &&
			titlesSimilar(query, results[i].Title, results[i].OriginalTitle) {
			return &results[i]
		}
	}

	// Second pass: any movie with a similar title.
	for i := range results {
		if titlesSimilar(query, results[i].Title, results[i].OriginalTitle) {
			return &results[i]
		}
	}

	// No sufficiently similar match; let the caller fall back.
	return nil
}

// GetJapaneseMovieOverview gets the Japanese overview for a movie
// Falls back to English if Japanese is not available
func (c *TMDBClient) GetJapaneseMovieOverview(ctx context.Context, tmdbID int) (string, error) {
	// First try Japanese
	details, err := c.GetMovieDetails(ctx, tmdbID, "ja-JP")
	if err != nil {
		return "", err
	}

	if details.Overview != "" {
		return details.Overview, nil
	}

	// Fall back to English if Japanese overview is empty
	details, err = c.GetMovieDetails(ctx, tmdbID, "en-US")
	if err != nil {
		return "", err
	}

	return details.Overview, nil
}

// GetOverviewByMALIDOrTitle gets Japanese overview using MAL ID first, then title
// Supports both TV shows and movies
// mediaTypeHint: "MOVIE", "TV", or "" (empty for unknown) - from AniList format
func (c *TMDBClient) GetOverviewByMALIDOrTitle(ctx context.Context, malID string, title string, mediaTypeHint string) (string, int, error) {
	tmdbID, mediaType, err := c.FindTMDBID(ctx, malID, title, mediaTypeHint)
	if err != nil {
		return "", 0, err
	}

	var overview string
	if mediaType == "movie" {
		overview, err = c.GetJapaneseMovieOverview(ctx, tmdbID)
	} else {
		overview, err = c.GetJapaneseOverview(ctx, tmdbID)
	}

	if err != nil {
		return "", tmdbID, err
	}

	return overview, tmdbID, nil
}
