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

const (
	annictAPIEndpoint = "https://api.annict.com/v1/works"
	defaultSortOrder  = "desc"
	defaultPerPage    = "50"

	// エピソード取得の制限
	annictMaxRetries         = 3   // 最大リトライ回数
	annictMaxPagesPerAnime   = 100 // 最大ページ数（無限ループ防止）
	annictMinEpisodesForSync = 2   // 同期対象の最小エピソード数（劇場版除外）

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

// executeRequestWithRetry executes HTTP request with retry logic
func (c *AnnictClient) executeRequestWithRetry(req *http.Request, workID, page int) (*http.Response, error) {
	var resp *http.Response
	var lastErr error

	for retry := 0; retry < annictMaxRetries; retry++ {
		if retry > 0 {
			backoffDelay := time.Duration(retry) * time.Second
			if c.logger != nil {
				c.logger.Info("Retrying API request",
					"work_id", workID,
					"page", page,
					"retry", retry,
					"backoff_delay", backoffDelay)
			}
			time.Sleep(backoffDelay)
		}

		resp, lastErr = c.HTTPClient.Do(req)
		if lastErr != nil {
			if c.logger != nil {
				c.logger.Warn("HTTP request failed",
					"work_id", workID,
					"page", page,
					"retry", retry,
					"error", lastErr)
			}
			continue
		}

		// レート制限エラー（429）の場合はリトライ
		if resp.StatusCode == http.StatusTooManyRequests {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("rate limit exceeded: %s", string(body))
			if c.logger != nil {
				c.logger.Warn("Rate limit exceeded, retrying",
					"work_id", workID,
					"page", page,
					"retry", retry)
			}
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", annictMaxRetries, lastErr)
}

// returnPartialDataIfAvailable returns partial data with error if available
func (c *AnnictClient) returnPartialDataIfAvailable(allEpisodes []AnnictEpisode, workID, page int, err error, context string) ([]AnnictEpisode, error) {
	if len(allEpisodes) > 0 {
		if c.logger != nil {
			c.logger.Warn(context,
				"work_id", workID,
				"page", page,
				"episodes_fetched", len(allEpisodes))
		}
		return allEpisodes, fmt.Errorf("partial data: %w", err)
	}
	return nil, err
}

// FetchEpisodesByWorkID fetches episode information from Annict API
func (c *AnnictClient) FetchEpisodesByWorkID(workID int) ([]AnnictEpisode, error) {
	if c.Token == "" {
		return nil, fmt.Errorf("annict API token is not set")
	}

	var allEpisodes []AnnictEpisode
	page := 1

	for page <= annictMaxPagesPerAnime {
		c.applyRateLimit()

		req, err := c.buildEpisodeRequest(workID, page)
		if err != nil {
			return nil, err
		}

		resp, err := c.executeRequestWithRetry(req, workID, page)
		if err != nil {
			return c.returnPartialDataIfAvailable(allEpisodes, workID, page, err, "Request failed")
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			err := fmt.Errorf("annict API returned status %d: %s", resp.StatusCode, string(body))
			return c.returnPartialDataIfAvailable(allEpisodes, workID, page, err, "API error")
		}

		var response AnnictEpisodesResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return c.returnPartialDataIfAvailable(allEpisodes, workID, page, err, "Decode error")
		}

		allEpisodes = append(allEpisodes, response.Episodes...)

		if response.NextPage == nil {
			break
		}

		page = *response.NextPage
	}

	if page > annictMaxPagesPerAnime && c.logger != nil {
		c.logger.Warn("Reached maximum page limit",
			"work_id", workID,
			"page", page,
			"episodes_fetched", len(allEpisodes))
	}

	if c.logger != nil {
		c.logger.Info("Fetched episodes from Annict",
			"work_id", workID,
			"episode_count", len(allEpisodes))
	}

	return allEpisodes, nil
}

// buildEpisodeRequest builds HTTP request for fetching episodes
func (c *AnnictClient) buildEpisodeRequest(workID, page int) (*http.Request, error) {
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

	return req, nil
}

// shouldSkipAnime checks if anime should be skipped (movies, shorts, etc.)
func (c *AnnictClient) shouldSkipAnime(animeID, annictID int) (bool, error) {
	var episodeCount int
	err := c.DB.QueryRow(`SELECT COUNT(*) FROM episodes WHERE anime_id = $1`, animeID).Scan(&episodeCount)
	if err != nil {
		return false, fmt.Errorf("failed to count episodes: %w", err)
	}

	if episodeCount < annictMinEpisodesForSync {
		if c.logger != nil {
			c.logger.Info("Skipping anime with too few episodes",
				"anime_id", animeID,
				"annict_id", annictID,
				"episode_count", episodeCount)
		}
		return true, nil
	}

	return false, nil
}

// fetchAnnictEpisodes fetches episodes from Annict API and handles partial data
func (c *AnnictClient) fetchAnnictEpisodes(animeID, annictID int) ([]AnnictEpisode, bool, error) {
	annictEpisodes, err := c.FetchEpisodesByWorkID(annictID)
	isPartialData := false

	if err != nil {
		if len(annictEpisodes) > 0 {
			isPartialData = true
			if c.logger != nil {
				c.logger.Warn("Partial episode data from Annict",
					"anime_id", animeID,
					"annict_id", annictID,
					"episodes_count", len(annictEpisodes),
					"error", err)
			}
		} else {
			return nil, false, fmt.Errorf("failed to fetch episodes from Annict: %w", err)
		}
	}

	if len(annictEpisodes) == 0 {
		return nil, false, fmt.Errorf("no episodes found for work_id %d", annictID)
	}

	return annictEpisodes, isPartialData, nil
}

// getDBEpisodeNumbers retrieves episode numbers from database
func (c *AnnictClient) getDBEpisodeNumbers(animeID int) ([]int, error) {
	rows, err := c.DB.Query(`
		SELECT episode_number FROM episodes 
		WHERE anime_id = $1 
		ORDER BY episode_number
	`, animeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query episodes: %w", err)
	}
	defer rows.Close()

	var episodeNumbers []int
	for rows.Next() {
		var episodeNumber int
		if err := rows.Scan(&episodeNumber); err != nil {
			if c.logger != nil {
				c.logger.Error("Failed to scan episode number",
					"anime_id", animeID,
					"error", err)
			}
			return nil, fmt.Errorf("failed to scan episode number: %w", err)
		}
		episodeNumbers = append(episodeNumbers, episodeNumber)
	}

	if err := rows.Err(); err != nil {
		if c.logger != nil {
			c.logger.Error("Error during rows iteration",
				"anime_id", animeID,
				"error", err)
		}
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	if len(episodeNumbers) == 0 {
		return nil, fmt.Errorf("no episodes found in database for anime_id %d", animeID)
	}

	return episodeNumbers, nil
}

// UpdateEpisodeTitlesFromAnnict updates episode titles from Annict API
func (c *AnnictClient) UpdateEpisodeTitlesFromAnnict(animeID int, annictID int) error {
	// 劇場版・短編アニメをスキップ
	skip, err := c.shouldSkipAnime(animeID, annictID)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}

	// Annict APIからエピソード情報を取得
	annictEpisodes, isPartialData, err := c.fetchAnnictEpisodes(animeID, annictID)
	if err != nil {
		return err
	}

	// データベースの既存エピソード番号を取得
	dbEpisodeNumbers, err := c.getDBEpisodeNumbers(animeID)
	if err != nil {
		return err
	}

	// データベースのエピソード番号とAnnict APIのエピソードを順序でマッチング
	// Annict APIのエピソードは既にsort_numberでソートされているため、
	// 通算エピソード番号（38, 39, 40...）でも正しく対応できる
	successCount := 0
	failureCount := 0

	// エピソード数の不一致を警告
	if len(dbEpisodeNumbers) != len(annictEpisodes) {
		if c.logger != nil {
			c.logger.Warn("Episode count mismatch",
				"anime_id", animeID,
				"db_count", len(dbEpisodeNumbers),
				"annict_count", len(annictEpisodes))
		}
	}

	for i, dbEpisodeNumber := range dbEpisodeNumbers {
		if i >= len(annictEpisodes) {
			if c.logger != nil {
				c.logger.Warn("Not enough episodes from Annict API",
					"anime_id", animeID,
					"db_episode_number", dbEpisodeNumber,
					"annict_episodes_count", len(annictEpisodes))
			}
			break // Annict APIのエピソードが足りない場合は終了
		}

		annictEp := annictEpisodes[i]

		// タイトルが空の場合はスキップ
		if annictEp.Title == "" {
			if c.logger != nil {
				c.logger.Warn("Empty episode title from Annict",
					"anime_id", animeID,
					"db_episode_number", dbEpisodeNumber,
					"annict_index", i)
			}
			failureCount++
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
					"annict_index", i,
					"error", err)
			}
			failureCount++
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			successCount++
			if c.logger != nil {
				c.logger.Debug("Updated episode title",
					"anime_id", animeID,
					"db_episode_number", dbEpisodeNumber,
					"annict_index", i,
					"title", annictEp.Title)
			}
		} else {
			if c.logger != nil {
				c.logger.Warn("No rows affected when updating episode",
					"anime_id", animeID,
					"db_episode_number", dbEpisodeNumber,
					"annict_index", i)
			}
			failureCount++
		}
	}

	if c.logger != nil {
		logLevel := "Info"
		if isPartialData {
			logLevel = "Warn"
		}

		if logLevel == "Warn" {
			c.logger.Warn("Updated episode titles (partial data)",
				"anime_id", animeID,
				"success_count", successCount,
				"failure_count", failureCount,
				"total_annict_episodes", len(annictEpisodes),
				"total_db_episodes", len(dbEpisodeNumbers))
		} else {
			c.logger.Info("Updated episode titles",
				"anime_id", animeID,
				"success_count", successCount,
				"failure_count", failureCount,
				"total_annict_episodes", len(annictEpisodes),
				"total_db_episodes", len(dbEpisodeNumbers))
		}
	}

	// 少なくとも1つのエピソードが更新されていればエラーとしない
	if successCount == 0 && len(dbEpisodeNumbers) > 0 {
		return fmt.Errorf("failed to update any episodes for anime_id %d", animeID)
	}

	return nil
}
