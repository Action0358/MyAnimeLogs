package services

import (
	"encoding/json"
	"fmt"
	"io"
	"myanimelogs/internal/common"
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
	annictAPIDefaultFields = "id,title,title_kana,title_en,media,media_text,season_name,released_on,released_on_about,official_site_url,wikipedia_url,twitter_username,twitter_hashtag,mal_anime_id,images,episodes_count,watchers_count,reviews_count,no_episodes"
)

type AnnictClient struct {
	Token          string
	HTTPClient     *http.Client
	DB             *database.DB
	logger         *common.Logger
	rateLimitDelay time.Duration
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
		rateLimitDelay: 500 * time.Millisecond,
	}
}

func NewAnnictClientWithDB(token string, db *database.DB, logger *common.Logger) *AnnictClient {
	return &AnnictClient{
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		DB:             db,
		logger:         logger,
		rateLimitDelay: 500 * time.Millisecond,
	}
}

// applyRateLimit applies rate limiting delay
func (c *AnnictClient) applyRateLimit() {
	if c.rateLimitDelay > 0 {
		time.Sleep(c.rateLimitDelay)
	}
}

// --- Private Method for DRY principle ---

func (c *AnnictClient) makeRequest(params url.Values) (*models.AnnictWorkResponse, error) {
	if c.Token == "" {
		return nil, fmt.Errorf("annict API token is not set")
	}

	// Apply rate limiting before making the request
	c.applyRateLimit()

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

	// Use season name directly (format: "YYYY-season", e.g., "2024-winter")
	season := work.SeasonName

	// Determine status based on season and current date
	status := "finished"
	if work.SeasonName != "" {
		// Extract year from season name (format: "YYYY-season")
		var seasonYear int
		_, err := fmt.Sscanf(work.SeasonName, "%d-", &seasonYear)
		if err == nil {
			currentYear := time.Now().Year()
			if seasonYear >= currentYear {
				status = "airing"
			}
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
				status = $6,
				media = $7,
				episodes_count = $8,
				official_site_url = $9,
				wikipedia_url = $10,
				twitter_username = $11,
				twitter_hashtag = $12,
				mal_anime_id = $13,
				watchers_count = $14,
				reviews_count = $15,
				updated_at = CURRENT_TIMESTAMP
			WHERE annict_id = $16
			RETURNING id, title, image_url, season, status, annict_id, created_at, updated_at
		`,
			work.Title, work.TitleKana, work.TitleEn, imageURL, season,
			status, work.Media, work.EpisodesCount, work.OfficialSiteURL, work.WikipediaURL,
			work.TwitterUsername, work.TwitterHashtag, work.MalAnimeID,
			work.WatchersCount, work.ReviewsCount, work.ID,
		).Scan(&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to update anime (annict_id=%d): %w", work.ID, err)
		}
	} else {
		// Anime doesn't exist, insert it
		// Use RETURNING to get the inserted row in a single query
		err = c.DB.QueryRow(`
			INSERT INTO animes (
				id, title, title_kana, title_en, image_url, season,
				status, media, episodes_count, official_site_url, wikipedia_url,
				twitter_username, twitter_hashtag, mal_anime_id, annict_id,
				watchers_count, reviews_count, created_at, updated_at
			) VALUES (
				DEFAULT, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
				CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
			)
			RETURNING id, title, image_url, season, status, annict_id, created_at, updated_at
		`,
			work.Title, work.TitleKana, work.TitleEn, imageURL, season,
			status, work.Media, work.EpisodesCount, work.OfficialSiteURL, work.WikipediaURL,
			work.TwitterUsername, work.TwitterHashtag, work.MalAnimeID, work.ID,
			work.WatchersCount, work.ReviewsCount,
		).Scan(&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to insert anime (annict_id=%d): %w", work.ID, err)
		}
	}

	// Note: Episode creation is now handled by AnimeService.createEpisodesForAnime
	// This avoids code duplication and centralizes episode management

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

// AnnictEpisode represents episode data from Annict API
type AnnictEpisode struct {
	ID         int     `json:"id"`
	Number     float64 `json:"number"`
	SortNumber int     `json:"sort_number"`
	Title      string  `json:"title"`
}

// AnnictEpisodesResponse represents the response from Annict episodes endpoint
type AnnictEpisodesResponse struct {
	Episodes   []AnnictEpisode `json:"episodes"`
	TotalCount int             `json:"total_count"`
	NextPage   *int            `json:"next_page"`
	PrevPage   *int            `json:"prev_page"`
}

// FetchEpisodesByWorkID fetches episode information from Annict API
func (c *AnnictClient) FetchEpisodesByWorkID(workID int) ([]AnnictEpisode, error) {
	if c.Token == "" {
		return nil, fmt.Errorf("annict API token is not set")
	}

	var allEpisodes []AnnictEpisode
	page := 1

	for {
		// Apply rate limiting before each API request
		c.applyRateLimit()

		params := url.Values{}
		params.Add("filter_work_id", fmt.Sprintf("%d", workID))
		params.Add("fields", "id,number,title,sort_number")
		params.Add("per_page", "50")
		params.Add("page", fmt.Sprintf("%d", page))
		params.Add("sort_sort_number", "asc")

		fullURL, err := url.Parse("https://api.annict.com/v1/episodes")
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}
		fullURL.RawQuery = params.Encode()

		req, err := http.NewRequest(http.MethodGet, fullURL.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("User-Agent", "MyAnimeLogs/1.0")

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("annict API returned status %d: %s", resp.StatusCode, string(body))
		}

		var response AnnictEpisodesResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		allEpisodes = append(allEpisodes, response.Episodes...)

		if response.NextPage == nil {
			break
		}

		page = *response.NextPage
	}

	if c.logger != nil {
		c.logger.Info("Fetched episodes from Annict",
			"work_id", workID,
			"episode_count", len(allEpisodes))
	}

	return allEpisodes, nil
}

// UpdateEpisodeTitlesFromAnnict updates episode titles from Annict API
func (c *AnnictClient) UpdateEpisodeTitlesFromAnnict(animeID int, annictID int) error {
	// Annict APIからエピソード情報を取得
	annictEpisodes, err := c.FetchEpisodesByWorkID(annictID)
	if err != nil {
		return fmt.Errorf("failed to fetch episodes from Annict: %w", err)
	}

	if len(annictEpisodes) == 0 {
		return fmt.Errorf("no episodes found for work_id %d", annictID)
	}

	// データベースの既存エピソード番号を取得
	rows, err := c.DB.Query(`
		SELECT episode_number FROM episodes 
		WHERE anime_id = $1 
		ORDER BY episode_number
	`, animeID)
	if err != nil {
		return fmt.Errorf("failed to query episodes: %w", err)
	}
	defer rows.Close()

	var dbEpisodeNumbers []int
	for rows.Next() {
		var episodeNumber int
		if err := rows.Scan(&episodeNumber); err != nil {
			// Database scan errors indicate serious problems and should not be ignored
			if c.logger != nil {
				c.logger.Error("Failed to scan episode number",
					"anime_id", animeID,
					"error", err)
			}
			return fmt.Errorf("failed to scan episode number: %w", err)
		}
		dbEpisodeNumbers = append(dbEpisodeNumbers, episodeNumber)
	}

	// Check for errors that occurred during iteration
	if err := rows.Err(); err != nil {
		if c.logger != nil {
			c.logger.Error("Error during rows iteration",
				"anime_id", animeID,
				"error", err)
		}
		return fmt.Errorf("error during rows iteration: %w", err)
	}

	if len(dbEpisodeNumbers) == 0 {
		return fmt.Errorf("no episodes found in database for anime_id %d", animeID)
	}

	// データベースのエピソード番号とAnnict APIのエピソードを順序でマッチング
	// Annict APIのエピソードは既にsort_numberでソートされているため、
	// 通算エピソード番号（38, 39, 40...）でも正しく対応できる
	successCount := 0
	for i, dbEpisodeNumber := range dbEpisodeNumbers {
		if i >= len(annictEpisodes) {
			break // Annict APIのエピソードが足りない場合は終了
		}

		annictEp := annictEpisodes[i]

		// Validate that Number is a positive value
		if annictEp.Number <= 0 {
			if c.logger != nil {
				c.logger.Warn("Invalid episode number from Annict",
					"anime_id", animeID,
					"annict_number", annictEp.Number,
					"title", annictEp.Title)
			}
			continue
		}

		// エピソードを更新
		result, err := c.DB.Exec(`
			UPDATE episodes
			SET 
				title = $1,
				updated_at = CURRENT_TIMESTAMP
			WHERE anime_id = $2 AND episode_number = $3
		`, annictEp.Title, animeID, dbEpisodeNumber)

		if err != nil {
			if c.logger != nil {
				c.logger.Warn("Failed to update episode",
					"anime_id", animeID,
					"db_episode_number", dbEpisodeNumber,
					"annict_number", annictEp.Number,
					"error", err)
			}
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			successCount++
			if c.logger != nil {
				c.logger.Debug("Updated episode title",
					"anime_id", animeID,
					"db_episode_number", dbEpisodeNumber,
					"annict_number", annictEp.Number,
					"title", annictEp.Title)
			}
		}
	}

	if c.logger != nil {
		c.logger.Info("Updated episode titles",
			"anime_id", animeID,
			"success_count", successCount,
			"total_episodes", len(annictEpisodes))
	}

	return nil
}
