package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/time/rate"

	"myanimelogs/internal/common"
	"myanimelogs/internal/models"
)

const (
	jikanBaseURL        = "https://api.jikan.moe/v4"
	jikanRateLimit      = 3                // 3 requests per second
	jikanRequestTimeout = 10 * time.Second // 10 seconds timeout
	maxRetries          = 3
)

var (
	// ErrJikanNotFound is returned when a resource is not found in Jikan API
	ErrJikanNotFound = errors.New("jikan: resource not found")
	// ErrJikanRateLimit is returned when rate limit is exceeded
	ErrJikanRateLimit = errors.New("jikan: rate limit exceeded")

	// Compiled regex patterns for performance
	durationRegex = regexp.MustCompile(`(?:(\d+)\s*hr)?(?:\s*(\d+)\s*min)?(?:\s*(\d+)\s*sec)?`)
)

// JikanClient handles communication with Jikan API
type JikanClient struct {
	baseURL     string
	httpClient  *http.Client
	logger      *common.Logger
	rateLimiter *rate.Limiter
}

// NewJikanClient creates a new Jikan API client
func NewJikanClient(logger *common.Logger) *JikanClient {
	return &JikanClient{
		baseURL: jikanBaseURL,
		httpClient: &http.Client{
			Timeout: jikanRequestTimeout,
		},
		logger:      logger,
		rateLimiter: rate.NewLimiter(rate.Limit(jikanRateLimit), 1),
	}
}

// GetAnimeDuration fetches the duration per episode for a given MAL anime ID
func (c *JikanClient) GetAnimeDuration(malAnimeID string) (int, error) {
	if malAnimeID == "" {
		return 0, fmt.Errorf("mal_anime_id is required")
	}

	url := fmt.Sprintf("%s/anime/%s", c.baseURL, malAnimeID)

	var response models.JikanAnimeResponse
	err := c.makeRequestWithRetry(url, &response)
	if err != nil {
		return 0, err
	}

	// Parse duration string (e.g., "24 min per ep" -> 24)
	duration := parseDuration(response.Data.Duration)

	return duration, nil
}

// GetAnimeAiredDate fetches the aired date for a given MAL anime ID
// Returns the "from" date in YYYY-MM-DD format (e.g., "2016-01-14")
func (c *JikanClient) GetAnimeAiredDate(malAnimeID string) (string, error) {
	if malAnimeID == "" {
		return "", fmt.Errorf("mal_anime_id is required")
	}

	url := fmt.Sprintf("%s/anime/%s", c.baseURL, malAnimeID)

	var response models.JikanAnimeResponse
	err := c.makeRequestWithRetry(url, &response)
	if err != nil {
		return "", err
	}

	// Check if aired data exists
	if response.Data.Aired == nil || response.Data.Aired.From == nil {
		c.logger.Debug("No aired date available from Jikan API",
			"mal_anime_id", malAnimeID)
		return "", nil
	}

	// Parse ISO 8601 date and convert to YYYY-MM-DD format
	// Example: "2016-01-14T00:00:00+00:00" -> "2016-01-14"
	fromDate := *response.Data.Aired.From
	if len(fromDate) >= 10 {
		releasedOn := fromDate[:10] // Extract YYYY-MM-DD part
		c.logger.Debug("Fetched aired date from Jikan API",
			"mal_anime_id", malAnimeID,
			"released_on", releasedOn)
		return releasedOn, nil
	}

	return "", nil
}

// makeRequestWithRetry makes an HTTP request with exponential backoff retry
func (c *JikanClient) makeRequestWithRetry(url string, result interface{}) error {
	var lastErr error
	ctx := context.Background()

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Wait for rate limiter
		err := c.rateLimiter.Wait(ctx)
		if err != nil {
			return fmt.Errorf("rate limiter error: %w", err)
		}

		// Make the request
		resp, err := c.httpClient.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			c.logger.Warn("Jikan API request failed, retrying",
				"url", url,
				"attempt", attempt+1,
				"error", err)
			c.backoff(attempt)
			continue
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			c.logger.Warn("Failed to read Jikan API response, retrying",
				"url", url,
				"attempt", attempt+1,
				"error", err)
			c.backoff(attempt)
			continue
		}

		// Handle HTTP errors
		switch resp.StatusCode {
		case http.StatusOK:
			// Success - parse JSON
			if err := json.Unmarshal(body, result); err != nil {
				return fmt.Errorf("failed to parse JSON response: %w", err)
			}
			return nil

		case http.StatusNotFound:
			// Anime not found - return sentinel error
			c.logger.Debug("Anime not found in Jikan API", "url", url)
			return ErrJikanNotFound

		case http.StatusTooManyRequests:
			// Rate limit exceeded - retry with backoff
			lastErr = ErrJikanRateLimit
			c.logger.Warn("Jikan API rate limit exceeded, retrying",
				"url", url,
				"attempt", attempt+1)
			c.backoff(attempt)
			continue

		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			// Server error - retry
			lastErr = fmt.Errorf("server error (%d)", resp.StatusCode)
			c.logger.Warn("Jikan API server error, retrying",
				"url", url,
				"status_code", resp.StatusCode,
				"attempt", attempt+1)
			c.backoff(attempt)
			continue

		default:
			// Other errors - don't retry
			return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// backoff implements exponential backoff
func (c *JikanClient) backoff(attempt int) {
	// Exponential backoff: 1s, 2s, 4s
	backoffDuration := time.Duration(math.Pow(2, float64(attempt))) * time.Second
	time.Sleep(backoffDuration)
}

// parseDuration parses duration string from Jikan API using regex
// Examples: "24 min per ep" -> 24, "1 hr 30 min" -> 90, "Unknown" -> 0
func parseDuration(durationStr string) int {
	if durationStr == "" || durationStr == "Unknown" {
		return 0
	}

	// Use pre-compiled regex pattern for performance
	// Matches: "1 hr 30 min", "24 min per ep", "45 min", "2 hr", "90 sec"
	matches := durationRegex.FindStringSubmatch(durationStr)

	if len(matches) < 4 {
		return 0
	}

	var totalMinutes int

	// Parse hours
	if matches[1] != "" {
		if hours, err := strconv.Atoi(matches[1]); err == nil {
			totalMinutes += hours * 60
		}
	}

	// Parse minutes
	if matches[2] != "" {
		if minutes, err := strconv.Atoi(matches[2]); err == nil {
			totalMinutes += minutes
		}
	}

	// Parse seconds (convert to minutes)
	if matches[3] != "" {
		if seconds, err := strconv.Atoi(matches[3]); err == nil {
			totalMinutes += seconds / 60
		}
	}

	return totalMinutes
}

// GetBroadcastInfo fetches broadcast schedule information for a given MAL anime ID
func (c *JikanClient) GetBroadcastInfo(malAnimeID string) (*models.JikanBroadcast, error) {
	if malAnimeID == "" {
		return nil, fmt.Errorf("mal_anime_id is required")
	}

	url := fmt.Sprintf("%s/anime/%s", c.baseURL, malAnimeID)

	var response models.JikanAnimeResponse
	err := c.makeRequestWithRetry(url, &response)
	if err != nil {
		return nil, err
	}

	if response.Data.Broadcast == nil {
		c.logger.Debug("No broadcast info available from Jikan API",
			"mal_anime_id", malAnimeID)
		return nil, nil
	}

	return response.Data.Broadcast, nil
}
