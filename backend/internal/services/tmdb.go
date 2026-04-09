package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"myanimelogs/internal/common"
)

const (
	tmdbBaseURL   = "https://api.themoviedb.org/3"
	tmdbTimeout   = 10 * time.Second
	tmdbImageBase = "https://image.tmdb.org/t/p"
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
		return nil, fmt.Errorf("failed to execute request: %w", err)
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
		return nil, fmt.Errorf("failed to execute request: %w", err)
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
		return nil, fmt.Errorf("failed to execute request: %w", err)
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
		return nil, fmt.Errorf("failed to execute request: %w", err)
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
		return nil, fmt.Errorf("failed to execute request: %w", err)
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
	// Try original title first
	result, err := c.SearchTV(ctx, title, "ja-JP")
	if err == nil && len(result.Results) > 0 {
		if show := c.findBestMatch(result.Results); show != nil {
			return show, nil
		}
	}

	// Try normalized title (remove season suffixes like "第3期", "Season 2", etc.)
	normalizedTitle := NormalizeAnimeTitle(title)
	if normalizedTitle != title {
		result, err = c.SearchTV(ctx, normalizedTitle, "ja-JP")
		if err == nil && len(result.Results) > 0 {
			if show := c.findBestMatch(result.Results); show != nil {
				return show, nil
			}
		}
	}

	return nil, fmt.Errorf("no results found for title: %s", title)
}

// findBestMatch finds the best matching TV show from search results
// Prioritizes Japanese anime (origin_country contains "JP")
func (c *TMDBClient) findBestMatch(results []TMDBTVShow) *TMDBTVShow {
	// Prioritize Japanese anime
	for i := range results {
		if slices.Contains(results[i].OriginCountry, "JP") {
			return &results[i]
		}
	}

	// If no Japanese show found, return the first result
	if len(results) > 0 {
		return &results[0]
	}
	return nil
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
	// English season patterns (greedy: removes everything after match)
	regexp.MustCompile(` Season \d+.*`),       // " Season 3 Part.2" etc.
	regexp.MustCompile(` The Final Season.*`), // " The Final Season Part 2"
	regexp.MustCompile(` Final Season.*`),     // " Final Season"
	regexp.MustCompile(` Part[.\s]*\d+$`),     // " Part 2", " Part.2"
	// Roman numerals
	regexp.MustCompile(` [IVX]+$`), // " II", " III", etc.
	// Other common suffixes
	regexp.MustCompile(` \d+nd Season.*`), // " 2nd Season"
	regexp.MustCompile(` \d+rd Season.*`), // " 3rd Season"
	regexp.MustCompile(` \d+th Season.*`), // " 4th Season"
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

	// Preserve the last error from the primary attempt
	lastErr := err

	// Fallback to the other mediaType
	fallbackType := "movie"
	if mediaType == "movie" {
		fallbackType = "tv"
	}

	providers, err = c.GetJapaneseWatchProviders(ctx, tmdbID, fallbackType)
	if err == nil && len(providers) > 0 {
		return providers, nil
	}
	if err != nil {
		lastErr = err
	}

	// Return nil with the last API error (or nil if APIs succeeded but had no results)
	return nil, lastErr
}

// TMDBFindResult represents the result of a find by external ID query
type TMDBFindResult struct {
	TVResults    []TMDBTVShow `json:"tv_results"`
	MovieResults []TMDBMovie  `json:"movie_results"`
}

// FindByMALID searches for a TV show or movie by MyAnimeList ID
// Returns the TMDB ID and media type ("tv" or "movie") if found
func (c *TMDBClient) FindByMALID(ctx context.Context, malID string) (int, string, error) {
	if !c.IsEnabled() {
		return 0, "", fmt.Errorf("TMDB API is not configured")
	}

	if malID == "" {
		return 0, "", fmt.Errorf("MAL ID is empty")
	}

	endpoint := fmt.Sprintf("%s/find/%s", tmdbBaseURL, malID)
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("external_source", "myanimelist_id")

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	var result TMDBFindResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Check TV results first
	if len(result.TVResults) > 0 {
		return result.TVResults[0].ID, "tv", nil
	}

	// Check movie results
	if len(result.MovieResults) > 0 {
		return result.MovieResults[0].ID, "movie", nil
	}

	return 0, "", fmt.Errorf("no TV show or movie found for MAL ID: %s", malID)
}

// FindTMDBID tries to find TMDB ID using MAL ID first, then falls back to title search
// Returns TMDB ID and media type ("tv" or "movie")
// mediaTypeHint: "MOVIE", "TV", or "" (empty for unknown) - from AniList format
func (c *TMDBClient) FindTMDBID(ctx context.Context, malID string, title string, mediaTypeHint string) (int, string, error) {
	// First, try to find by MAL ID (most accurate)
	if malID != "" {
		tmdbID, mediaType, err := c.FindByMALID(ctx, malID)
		if err == nil && tmdbID > 0 {
			return tmdbID, mediaType, nil
		}
		// Log but continue to title search
		if c.logger != nil {
			c.logger.Debug("MAL ID lookup failed, falling back to title search",
				"mal_id", malID,
				"error", err)
		}
	}

	// Fall back to title search based on media type hint
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
	// Try original title first
	result, err := c.SearchMovie(ctx, title, "ja-JP")
	if err == nil && len(result.Results) > 0 {
		if movie := c.findBestMovieMatch(result.Results); movie != nil {
			return movie, nil
		}
	}

	// Try normalized title
	normalizedTitle := NormalizeAnimeTitle(title)
	if normalizedTitle != title {
		result, err = c.SearchMovie(ctx, normalizedTitle, "ja-JP")
		if err == nil && len(result.Results) > 0 {
			if movie := c.findBestMovieMatch(result.Results); movie != nil {
				return movie, nil
			}
		}
	}

	return nil, fmt.Errorf("no movie found for title: %s", title)
}

// findBestMovieMatch finds the best matching movie from search results
// Prioritizes Japanese anime movies (original_language is "ja")
func (c *TMDBClient) findBestMovieMatch(results []TMDBMovie) *TMDBMovie {
	// Prioritize Japanese movies
	for i := range results {
		if results[i].OriginalLanguage == "ja" {
			return &results[i]
		}
	}

	// If no Japanese movie found, return the first result
	if len(results) > 0 {
		return &results[0]
	}
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
