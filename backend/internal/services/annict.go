package services

import (
	"encoding/json"
	"fmt"
	"io"
	"myanimelogs/internal/database"
	"myanimelogs/internal/models"
	"net/http"
	"net/url"
	"time"
)

const annictAPIEndpoint = "https://api.annict.com/v1/works"

const (
	defaultSortOrder = "desc"
	defaultPerPage   = "50"
	// AnnictAPIのデフォルトフィールド定義
	annictAPIDefaultFields = "id,title,title_kana,title_en,media,media_text,season_name,season_year,released_on,released_on_about,official_site_url,wikipedia_url,twitter_username,twitter_hashtag,mal_anime_id,images,episodes_count,watchers_count,reviews_count,no_episodes"
)

type AnnictClient struct {
	Token      string
	HTTPClient *http.Client
	DB         *database.DB
}

// AnnictService defines the interface for Annict API interactions.
type AnnictService interface {
	FetchWorksBySeason(season string) (*models.AnnictWorkResponse, error)
	FetchPopularWorks() (*models.AnnictWorkResponse, error)
	TestConnection() error
}

func NewAnnictClient(token string) *AnnictClient {
	return &AnnictClient{
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func NewAnnictClientWithDB(token string, db *database.DB) *AnnictClient {
	return &AnnictClient{
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		DB: db,
	}
}

// --- Private Method for DRY principle ---

func (c *AnnictClient) makeRequest(params url.Values) (*models.AnnictWorkResponse, error) {
	if c.Token == "" {
		return nil, fmt.Errorf("annict API token is not set")
	}

	fullURL, err := url.Parse(annictAPIEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base url: %w", err)
	}
	fullURL.RawQuery = params.Encode()

	req, err := http.NewRequest(http.MethodGet, fullURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("User-Agent", "MyAnimeLogs/1.0")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("annict api returned non-200 status: %d, and failed to read response body: %w", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("annict api returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
	}

	var annictResponse models.AnnictWorkResponse
	if err := json.NewDecoder(resp.Body).Decode(&annictResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal annict response: %w", err)
	}

	return &annictResponse, nil
}

// --- Public Methods ---

func (c *AnnictClient) FetchWorksBySeason(season string) (*models.AnnictWorkResponse, error) {
	params := url.Values{}
	params.Add("filter_season", season)
	params.Add("sort_watchers_count", defaultSortOrder)
	params.Add("per_page", defaultPerPage)
	params.Add("fields", annictAPIDefaultFields)
	return c.makeRequest(params)
}

func (c *AnnictClient) FetchPopularWorks() (*models.AnnictWorkResponse, error) {
	params := url.Values{}
	params.Add("sort_watchers_count", defaultSortOrder)
	params.Add("per_page", defaultPerPage)
	params.Add("fields", annictAPIDefaultFields)
	return c.makeRequest(params)
}

func (c *AnnictClient) TestConnection() error {
	params := url.Values{}
	params.Add("per_page", "1")
	_, err := c.makeRequest(params)
	return err
}

// SyncAnimeToDatabase syncs Annict work data to local database
func (c *AnnictClient) SyncAnimeToDatabase(work models.AnnictWork) (*models.Anime, error) {
	if c.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Extract image URL
	var imageURL *string
	if work.Images != nil {
		if work.Images.RecommendedURL != "" {
			imageURL = &work.Images.RecommendedURL
		} else if work.Images.Facebook != nil && work.Images.Facebook.OgImageURL != "" {
			imageURL = &work.Images.Facebook.OgImageURL
		}
	}

	// Convert season format (e.g., "2024-winter" -> "2024-winter")
	season := work.SeasonName
	if season == "" && work.SeasonYear > 0 {
		season = fmt.Sprintf("%d-unknown", work.SeasonYear)
	}

	// Determine status based on season and current date
	status := "finished"
	if work.SeasonName != "" {
		// Simple logic: if it's current year, might be airing
		currentYear := time.Now().Year()
		if work.SeasonYear >= currentYear {
			status = "airing"
		}
	}

	// Upsert anime record using manual SELECT + INSERT/UPDATE approach
	// This is a workaround for ON CONFLICT not working properly
	var anime models.Anime

	// Check if anime already exists
	var existingID int
	err := c.DB.QueryRow("SELECT id FROM animes WHERE annict_id = $1", work.ID).Scan(&existingID)

	if err == nil && existingID > 0 {
		// Anime exists, update it
		err = c.DB.QueryRow(`
			UPDATE animes SET
				title = $1,
				title_kana = $2,
				title_en = $3,
				image_url = $4,
				season = $5,
				season_year = $6,
				status = $7,
				media = $8,
				episodes_count = $9,
				official_site_url = $10,
				wikipedia_url = $11,
				twitter_username = $12,
				twitter_hashtag = $13,
				mal_anime_id = $14,
				watchers_count = $15,
				reviews_count = $16,
				updated_at = CURRENT_TIMESTAMP
			WHERE annict_id = $17
			RETURNING id, title, description, image_url, season, status, annict_id, created_at, updated_at
		`,
			work.Title, work.TitleKana, work.TitleEn, imageURL, season, work.SeasonYear,
			status, work.Media, work.EpisodesCount, work.OfficialSiteURL, work.WikipediaURL,
			work.TwitterUsername, work.TwitterHashtag, work.MalAnimeID,
			work.WatchersCount, work.ReviewsCount, work.ID,
		).Scan(&anime.ID, &anime.Title, &anime.Description, &anime.ImageURL, &anime.Season, &anime.Status, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to update anime (annict_id=%d): %w", work.ID, err)
		}
	} else {
		// Anime doesn't exist, insert it
		// Use RETURNING to get the inserted row in a single query
		err = c.DB.QueryRow(`
			INSERT INTO animes (
				id, title, title_kana, title_en, description, image_url, season, season_year,
				status, media, episodes_count, official_site_url, wikipedia_url,
				twitter_username, twitter_hashtag, mal_anime_id, annict_id,
				watchers_count, reviews_count, created_at, updated_at
			) VALUES (
				DEFAULT, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18,
				CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
			)
			RETURNING id, title, description, image_url, season, status, annict_id, created_at, updated_at
		`,
			work.Title, work.TitleKana, work.TitleEn, "", imageURL, season, work.SeasonYear,
			status, work.Media, work.EpisodesCount, work.OfficialSiteURL, work.WikipediaURL,
			work.TwitterUsername, work.TwitterHashtag, work.MalAnimeID, work.ID,
			work.WatchersCount, work.ReviewsCount,
		).Scan(&anime.ID, &anime.Title, &anime.Description, &anime.ImageURL, &anime.Season, &anime.Status, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to insert anime (annict_id=%d): %w", work.ID, err)
		}
	}

	return &anime, nil
}

// SyncWorksToDatabase syncs multiple works to database
func (c *AnnictClient) SyncWorksToDatabase(works []models.AnnictWork) ([]models.Anime, error) {
	var animes []models.Anime

	for _, work := range works {
		anime, err := c.SyncAnimeToDatabase(work)
		if err != nil {
			// Log error but continue with other works
			fmt.Printf("Failed to sync anime %s (ID: %d): %v\n", work.Title, work.ID, err)
			continue
		}
		animes = append(animes, *anime)
	}

	return animes, nil
}

// FetchAndSyncWorksBySeason fetches works from Annict and syncs to database
func (c *AnnictClient) FetchAndSyncWorksBySeason(season string) ([]models.Anime, error) {
	response, err := c.FetchWorksBySeason(season)
	if err != nil {
		return nil, err
	}

	return c.SyncWorksToDatabase(response.Works)
}

// FetchAndSyncPopularWorks fetches popular works from Annict and syncs to database
func (c *AnnictClient) FetchAndSyncPopularWorks() ([]models.Anime, error) {
	response, err := c.FetchPopularWorks()
	if err != nil {
		return nil, err
	}

	return c.SyncWorksToDatabase(response.Works)
}

// SearchAnimesByKeyword searches for anime by keyword using Annict API
func (c *AnnictClient) SearchAnimesByKeyword(keyword string) (*models.AnnictWorkResponse, error) {
	params := url.Values{}
	params.Add("filter_title", keyword)
	params.Add("sort_watchers_count", defaultSortOrder)
	params.Add("per_page", "20")
	params.Add("fields", annictAPIDefaultFields)
	return c.makeRequest(params)
}
