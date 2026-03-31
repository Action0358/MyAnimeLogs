package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"myanimelogs/internal/common"
	"myanimelogs/internal/constants"
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

var (
	// ErrRateLimitExceeded is returned when the API rate limit is exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

type AnnictClient struct {
	Token          string
	HTTPClient     *http.Client
	DB             *database.DB
	logger         *common.Logger
	rateLimitDelay time.Duration
	anilistClient  *AniListClient
	animeService   *AnimeService // Add AnimeService here
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
		rateLimitDelay: 1000 * time.Millisecond, // Increased from 500ms to 1s to avoid rate limiting
	}
}

func NewAnnictClientWithDB(token string, db *database.DB, logger *common.Logger) *AnnictClient { // animeService引数を削除
	return &AnnictClient{
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		DB:             db,
		logger:         logger,
		rateLimitDelay: 1000 * time.Millisecond, // Increased from 500ms to 1s to avoid rate limiting
		anilistClient:  NewAniListClient(),      // Initialize once for reuse
		// animeService は後からSetAnimeServiceで設定する
	}
}

// SetAnimeService - AnimeServiceをAnnictClientに設定する
func (c *AnnictClient) SetAnimeService(as *AnimeService) {
	c.animeService = as
} // applyRateLimit applies rate limiting delay
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
			return nil, fmt.Errorf("annict API returned status %d, and failed to read response body: %w", resp.StatusCode, readErr)
		}

		// Return sentinel error for rate limiting
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("status %d: %w", resp.StatusCode, ErrRateLimitExceeded)
		}

		return nil, fmt.Errorf("annict API returned status %d, body: %s", resp.StatusCode, string(body))
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
	return c.makeRequest(params)
}

func (c *AnnictClient) FetchPopularWorks() (*models.AnnictWorkResponse, error) {
	params := url.Values{}
	params.Add("sort_watchers_count", defaultSortOrder)
	params.Add("per_page", defaultPerPage)
	return c.makeRequest(params)
}

// FetchPopularWorksWithLimit fetches popular works with pagination to get up to the specified limit
// This function makes multiple API calls to fetch more than the default per_page limit
// Returns partial data with error if API fails mid-pagination
// Maximum pages is limited to prevent excessive API calls
func (c *AnnictClient) FetchPopularWorksWithLimit(limit int) (*models.AnnictWorkResponse, error) {
	if limit <= 0 {
		limit = constants.AnnictAPIPerPage // デフォルト
	}

	// 上限を超えるlimitが指定された場合は制限
	if limit > constants.PopularAnimeLimit {
		limit = constants.PopularAnimeLimit
	}

	var allWorks []models.AnnictWork
	var partialErr error
	page := 1
	perPage := constants.AnnictAPIPerPage                   // Annict APIの1ページあたりの最大件数
	maxPages := (constants.PopularAnimeLimit / perPage) + 1 // 最大ページ数（レート制限対策）

	for len(allWorks) < limit && page <= maxPages {
		params := url.Values{}
		params.Add("sort_watchers_count", defaultSortOrder)
		params.Add("per_page", fmt.Sprintf("%d", perPage))
		params.Add("page", fmt.Sprintf("%d", page))

		response, err := c.makeRequest(params)
		if err != nil {
			// 部分的なデータがある場合はデータとエラーの両方を返す
			if len(allWorks) > 0 {
				if c.logger != nil {
					c.logger.Warn("Partial data returned due to API error",
						"page", page,
						"fetched", len(allWorks),
						"error", err)
				}
				partialErr = fmt.Errorf("partial data: fetched %d items before error: %w", len(allWorks), err)
				break
			}
			return nil, err
		}

		if len(response.Works) == 0 {
			break // これ以上データがない
		}

		allWorks = append(allWorks, response.Works...)

		// 次のページがない場合は終了
		if response.NextPage == nil {
			break
		}

		page++
	}

	// 最大ページ数に達した場合の警告
	if page > maxPages && c.logger != nil {
		c.logger.Warn("Reached maximum page limit for popular works",
			"maxPages", maxPages,
			"fetched", len(allWorks))
	}

	// 指定された上限を超えた場合は切り詰める
	if len(allWorks) > limit {
		allWorks = allWorks[:limit]
	}

	if c.logger != nil {
		c.logger.Info("Fetched popular works from Annict",
			"total", len(allWorks),
			"limit", limit)
	}

	return &models.AnnictWorkResponse{
		Works:      allWorks,
		TotalCount: len(allWorks),
	}, partialErr
}

func (c *AnnictClient) TestConnection() error {
	params := url.Values{}
	params.Add("per_page", "1")
	_, err := c.makeRequest(params)
	return err
}

// SyncAnimeToDatabaseResult - SyncAnimeToDatabaseの結果
type SyncAnimeToDatabaseResult struct {
	Anime   *models.Anime
	Updated bool // 実際にDB更新が行われたか（スキップの場合false）
}

// SyncAnimeToDatabase syncs Annict work data to local database
// existingAnime: 呼び出し元で既に取得済みの場合に渡す（nilの場合は内部で取得）
func (c *AnnictClient) SyncAnimeToDatabase(work models.AnnictWork, existingAnime *models.Anime) (*SyncAnimeToDatabaseResult, error) {
	if c.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// 既存データが渡されていない場合のみDBから取得（二重クエリ防止）
	if existingAnime == nil {
		var err error
		existingAnime, err = c.getExistingAnime(work.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing anime: %w", err)
		}
	}

	// 差分判定
	if !c.needsSync(work, existingAnime) {
		// 変更なし - スキップ
		if c.logger != nil {
			c.logger.Debug("Skipping sync - no changes detected",
				"annict_id", work.ID,
				"title", work.Title)
		}
		return &SyncAnimeToDatabaseResult{Anime: existingAnime, Updated: false}, nil
	}

	// Annict更新日時をパース（UPDATE/INSERT共通で使用）
	var annictUpdatedAt *time.Time
	if work.UpdatedAt != "" {
		parsed, parseErr := time.Parse(time.RFC3339, work.UpdatedAt)
		if parseErr == nil {
			annictUpdatedAt = &parsed
		}
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

	// デフォルトステータスを設定（後でSyncAnimeStatusで更新される）
	status := constants.AnimeStatusFinished.String()

	var anime models.Anime

	// Check if anime already exists
	var existingID int
	err := c.DB.QueryRow("SELECT id FROM animes WHERE annict_id = $1", work.ID).Scan(&existingID)

	if err == nil && existingID > 0 {
		// Anime exists, update it
		// Note: statusはSyncAnimeStatusで管理するため、ここでは上書きしない
		// released_onはAnnictの値がある場合のみ更新（COALESCEで既存値を保持）
		err = c.DB.QueryRow(`
			UPDATE animes SET
				title = $1,
				title_kana = $2,
				title_en = $3,
				image_url = $4,
				season = $5,
				media = $6,
				episodes_count = $7,
				official_site_url = $8,
				wikipedia_url = $9,
				twitter_username = $10,
				twitter_hashtag = $11,
				mal_anime_id = $12,
				watchers_count = $13,
				reviews_count = $14,
				released_on = COALESCE(NULLIF($15, ''), released_on),
				last_synced_at = CURRENT_TIMESTAMP,
				annict_updated_at = $17,
				sync_status = $18,
				sync_error = NULL,
				updated_at = CURRENT_TIMESTAMP
			WHERE annict_id = $16
			RETURNING id, title, image_url, season, status, annict_id, created_at, updated_at, last_synced_at, annict_updated_at
		`,
			work.Title, work.TitleKana, work.TitleEn, imageURL, season,
			work.Media, work.EpisodesCount, work.OfficialSiteURL, work.WikipediaURL,
			work.TwitterUsername, work.TwitterHashtag, work.MalAnimeID,
			work.WatchersCount, work.ReviewsCount, work.ReleasedOn, work.ID,
			annictUpdatedAt, string(constants.SyncStatusSynced),
		).Scan(&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.LastSyncedAt, &anime.AnnictUpdatedAt)

		if err != nil {
			// 同期失敗時のステータス更新
			c.updateSyncFailureStatus(work.ID, err)
			return nil, fmt.Errorf("failed to update anime (annict_id=%d): %w", work.ID, err)
		}
	} else {
		// Anime doesn't exist, insert it
		err = c.DB.QueryRow(`
			INSERT INTO animes (
				id, title, title_kana, title_en, image_url, season,
				status, media, episodes_count, official_site_url, wikipedia_url,
				twitter_username, twitter_hashtag, mal_anime_id, annict_id,
				watchers_count, reviews_count, released_on,
				last_synced_at, annict_updated_at, sync_status,
				created_at, updated_at
			) VALUES (
				DEFAULT, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
				CURRENT_TIMESTAMP, $18, $19,
				CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
			)
			RETURNING id, title, image_url, season, status, annict_id, created_at, updated_at, last_synced_at, annict_updated_at
		`,
			work.Title, work.TitleKana, work.TitleEn, imageURL, season,
			status, work.Media, work.EpisodesCount, work.OfficialSiteURL, work.WikipediaURL,
			work.TwitterUsername, work.TwitterHashtag, work.MalAnimeID, work.ID,
			work.WatchersCount, work.ReviewsCount, work.ReleasedOn,
			annictUpdatedAt, string(constants.SyncStatusSynced),
		).Scan(&anime.ID, &anime.Title, &anime.ImageURL, &anime.Season, &anime.Status, &anime.AnnictID, &anime.CreatedAt, &anime.UpdatedAt, &anime.LastSyncedAt, &anime.AnnictUpdatedAt)

		if err != nil {
			if c.logger != nil {
				c.logger.Error("Failed to insert anime",
					"annict_id", work.ID,
					"title", work.Title,
					"error", err)
			}
			return nil, fmt.Errorf("failed to insert anime (annict_id=%d): %w", work.ID, err)
		}
	}

	// Annictからエピソードを作成（主軸）
	// AnimeServiceのCreateEpisodesFromAnnictを使用してエピソードレコードを作成
	if c.animeService != nil && work.EpisodesCount > 0 {
		animeID := anime.ID
		annictID := work.ID
		c.animeService.enqueueBackgroundTask(func() {
			_, err := c.animeService.CreateEpisodesFromAnnict(animeID, annictID)
			if err != nil {
				if c.logger != nil {
					c.logger.Warn("Failed to create episodes from Annict",
						"anime_id", animeID,
						"annict_id", annictID,
						"error", err)
				}
			}
		})
	}

	// Sync genres synchronously if MAL ID is available
	// Note: Must be synchronous to respect AniList rate limiter (90 req/min).
	// Running as goroutines causes burst requests that exceed the rate limit.
	if work.MalAnimeID != nil && *work.MalAnimeID != "" && c.anilistClient != nil {
		ctx := context.Background()

		genres, err := c.anilistClient.GetGenresByMALId(ctx, *work.MalAnimeID)
		if err != nil {
			if c.logger != nil {
				c.logger.Warn("Failed to get genres from AniList",
					"anime_id", anime.ID,
					"mal_anime_id", *work.MalAnimeID,
					"error", err)
			}
		} else {
			err = saveGenresToDatabase(c.DB, c.logger, anime.ID, genres)
			if err != nil {
				if c.logger != nil {
					c.logger.Error("Failed to save genres",
						"anime_id", anime.ID,
						"error", err)
				}
			} else {
				if c.logger != nil {
					c.logger.Info("Successfully saved genres",
						"anime_id", anime.ID,
						"genres", genres)
				}
			}
		}
	}

	// アニメのステータスを非同期で更新（ワーカープールを使用して並行数を制御）
	// SyncAnimeStatusはAnimeServiceのメソッドなので、AnnictClientからはanimeService経由で呼び出す
	if c.animeService != nil {
		animeID := anime.ID
		c.animeService.enqueueBackgroundTask(func() {
			err := c.animeService.SyncAnimeStatus(animeID, true)
			if err != nil {
				c.logger.Error("Failed to sync anime status after Annict sync",
					"anime_id", animeID,
					"error", err)
			} else {
				c.logger.Debug("Anime status sync triggered for", "anime_id", animeID)
			}
		})
	}

	return &SyncAnimeToDatabaseResult{Anime: &anime, Updated: true}, nil
}

// SyncWorksToDatabase syncs multiple works to database
func (c *AnnictClient) SyncWorksToDatabase(works []models.AnnictWork) ([]models.Anime, error) {
	var animes []models.Anime

	for _, work := range works {
		result, err := c.SyncAnimeToDatabase(work, nil)
		if err != nil {
			// Log error but continue with other works
			if c.logger != nil {
				c.logger.Warn("Failed to sync anime",
					"title", work.Title,
					"annict_id", work.ID,
					"error", err)
			}
			continue
		}
		if result.Anime != nil {
			animes = append(animes, *result.Anime)
		}
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
// Uses PopularAnimeLimit (100) to fetch top popular anime
func (c *AnnictClient) FetchAndSyncPopularWorks() ([]models.Anime, error) {
	response, err := c.FetchPopularWorksWithLimit(constants.PopularAnimeLimit)
	if err != nil {
		return nil, err
	}

	return c.SyncWorksToDatabase(response.Works)
}

// SyncAllExistingAnimes re-syncs all existing animes from Annict API using differential update
// Only syncs animes that need updating (based on annict_updated_at or last_synced_at)
// Returns: success count, failure count, error
func (c *AnnictClient) SyncAllExistingAnimes() (int, int, error) {
	if c.DB == nil {
		return 0, 0, fmt.Errorf("database connection not available")
	}

	// Get animes that need syncing (differential update)
	// Priority 1: Never synced (last_synced_at IS NULL)
	// Priority 2: Stale data (last_synced_at older than StaleDataThreshold)
	// Priority 3: Recently updated on Annict (annict_updated_at > last_synced_at)
	staleThreshold := time.Now().Add(-constants.StaleDataThreshold)
	rows, err := c.DB.Query(`
		SELECT id, annict_id, title, last_synced_at, annict_updated_at
		FROM animes 
		WHERE annict_id IS NOT NULL
		AND (
			last_synced_at IS NULL
			OR last_synced_at < $1
			OR (annict_updated_at IS NOT NULL AND annict_updated_at > last_synced_at)
		)
		ORDER BY 
			CASE 
				WHEN last_synced_at IS NULL THEN 1
				WHEN annict_updated_at IS NOT NULL AND annict_updated_at > last_synced_at THEN 2
				ELSE 3
			END,
			watchers_count DESC NULLS LAST, 
			id
	`, staleThreshold)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get animes: %w", err)
	}

	type animeInfo struct {
		ID              int
		AnnictID        int
		Title           string
		LastSyncedAt    *time.Time
		AnnictUpdatedAt *time.Time
	}
	var animes []animeInfo

	for rows.Next() {
		var a animeInfo
		if err := rows.Scan(&a.ID, &a.AnnictID, &a.Title, &a.LastSyncedAt, &a.AnnictUpdatedAt); err != nil {
			continue
		}
		animes = append(animes, a)
	}
	rows.Close()

	if err := rows.Err(); err != nil {
		return 0, 0, fmt.Errorf("error iterating anime rows: %w", err)
	}

	totalAnimes := len(animes)
	if totalAnimes == 0 {
		if c.logger != nil {
			c.logger.Info("No animes need syncing (all up-to-date)")
		}
		return 0, 0, nil
	}

	if c.logger != nil {
		c.logger.Info("Starting Annict differential sync",
			"needs_sync", totalAnimes,
			"note", "Only syncing animes that need updates")
	}

	// Annict rate limit: 5 requests/second
	// Process in batches of 50 with delays
	const batchSize = 50

	successCount := 0
	failureCount := 0
	skippedCount := 0

	for i := 0; i < totalAnimes; i += batchSize {
		end := i + batchSize
		if end > totalAnimes {
			end = totalAnimes
		}
		batch := animes[i:end]

		if c.logger != nil {
			c.logger.Info("Processing Annict batch",
				"progress", fmt.Sprintf("%d/%d", end, totalAnimes),
				"success", successCount,
				"skipped", skippedCount,
				"failure", failureCount)
		}

		for _, anime := range batch {
			// Apply rate limiting
			c.applyRateLimit()

			// Fetch work from Annict
			params := url.Values{}
			params.Add("filter_ids", fmt.Sprintf("%d", anime.AnnictID))
			params.Add("fields", annictAPIDefaultFields)
			params.Add("per_page", "1")

			response, err := c.makeRequest(params)
			if err != nil {
				if c.logger != nil {
					c.logger.Debug("Failed to fetch from Annict",
						"annict_id", anime.AnnictID,
						"title", anime.Title,
						"error", err)
				}
				failureCount++
				continue
			}

			if len(response.Works) == 0 {
				failureCount++
				continue
			}

			// 取得済みデータからexistingAnimeを構築（二重クエリ防止）
			existing := &models.Anime{
				ID:              anime.ID,
				AnnictID:        &anime.AnnictID,
				Title:           anime.Title,
				LastSyncedAt:    anime.LastSyncedAt,
				AnnictUpdatedAt: anime.AnnictUpdatedAt,
			}

			// Sync to database (differential update inside)
			result, err := c.SyncAnimeToDatabase(response.Works[0], existing)
			if err != nil {
				if c.logger != nil {
					c.logger.Debug("Failed to sync anime",
						"annict_id", anime.AnnictID,
						"title", anime.Title,
						"error", err)
				}
				failureCount++
				continue
			}

			// Updated フラグで更新/スキップを正確に区別
			if result.Updated {
				successCount++
			} else {
				skippedCount++
			}
		}
	}

	if c.logger != nil {
		c.logger.Info("Annict differential sync completed",
			"success", successCount,
			"skipped", skippedCount,
			"failure", failureCount,
			"total_processed", totalAnimes)
	}

	return successCount, failureCount, nil
}

// SearchAnimesByKeyword searches for anime by keyword using Annict API
func (c *AnnictClient) SearchAnimesByKeyword(keyword string) (*models.AnnictWorkResponse, error) {
	params := url.Values{}
	params.Add("filter_title", keyword)
	params.Add("sort_watchers_count", defaultSortOrder)
	params.Add("per_page", "20")
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

// executeRequestWithRetry executes HTTP request with exponential backoff retry logic
func (c *AnnictClient) executeRequestWithRetry(req *http.Request, workID, page int) (*http.Response, error) {
	var resp *http.Response
	var lastErr error

	for retry := 0; retry < annictMaxRetries; retry++ {
		if retry > 0 {
			// Exponential backoff: 2^(retry-1) seconds (1s, 2s, 4s, ...)
			backoffDelay := time.Duration(1<<uint(retry-1)) * time.Second
			if c.logger != nil {
				c.logger.Info("Retrying API request with exponential backoff",
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
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				lastErr = fmt.Errorf("rate limit exceeded (status 429), failed to read response body: %w", readErr)
			} else {
				lastErr = fmt.Errorf("rate limit exceeded: %s", string(body))
			}
			if c.logger != nil {
				c.logger.Warn("Rate limit exceeded, retrying with exponential backoff",
					"work_id", workID,
					"page", page,
					"retry", retry,
					"error", lastErr)
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
				"episodes_fetched", len(allEpisodes),
				"error", err)
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

		// Process response and close body immediately to avoid resource leaks
		if resp.StatusCode != http.StatusOK {
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			var err error
			if readErr != nil {
				err = fmt.Errorf("annict API returned status %d, failed to read response body: %w", resp.StatusCode, readErr)
			} else {
				err = fmt.Errorf("annict API returned status %d: %s", resp.StatusCode, string(body))
			}
			return c.returnPartialDataIfAvailable(allEpisodes, workID, page, err, "API error")
		}

		var response AnnictEpisodesResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			return c.returnPartialDataIfAvailable(allEpisodes, workID, page, err, "Decode error")
		}
		resp.Body.Close()

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

	// Annict APIのエピソードは既にsort_numberでソートされている
	// データベースのエピソード番号とAnnict APIのエピソードを順序でマッチング
	// DBのepisode_number=1がAnnictの最初のエピソード(インデックス0)に対応
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
		if isPartialData {
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

// getExistingAnime - 既存のアニメデータを取得
func (c *AnnictClient) getExistingAnime(annictID int) (*models.Anime, error) {
	var anime models.Anime
	err := c.DB.QueryRow(`
		SELECT id, annict_id, title, last_synced_at, annict_updated_at, metadata_synced_at, sync_status
		FROM animes
		WHERE annict_id = $1
	`, annictID).Scan(
		&anime.ID,
		&anime.AnnictID,
		&anime.Title,
		&anime.LastSyncedAt,
		&anime.AnnictUpdatedAt,
		&anime.MetadataSyncedAt,
		&anime.SyncStatus,
	)

	if err == sql.ErrNoRows {
		return nil, nil // 新規作品
	}
	if err != nil {
		return nil, err
	}

	return &anime, nil
}

// needsSync - アニメ作品の同期が必要かを判定
func (c *AnnictClient) needsSync(work models.AnnictWork, existingAnime *models.Anime) bool {
	// 新規作品は同期必須
	if existingAnime == nil {
		return true
	}

	// Annict側の更新日時が新しい場合のみ同期
	if work.UpdatedAt != "" && existingAnime.AnnictUpdatedAt != nil {
		workUpdated, err := time.Parse(time.RFC3339, work.UpdatedAt)
		if err == nil && workUpdated.After(*existingAnime.AnnictUpdatedAt) {
			return true
		}
	}

	// 一定期間同期していない場合は強制同期
	if existingAnime.LastSyncedAt != nil {
		if time.Since(*existingAnime.LastSyncedAt) > constants.StaleDataThreshold {
			return true
		}
	} else {
		// last_synced_atがnullの場合は同期必須（マイグレーション直後）
		return true
	}

	return false
}

// updateSyncFailureStatus - 同期失敗時のステータスを更新
func (c *AnnictClient) updateSyncFailureStatus(annictID int, err error) {
	if c.DB == nil {
		return
	}

	_, updateErr := c.DB.Exec(`
		UPDATE animes
		SET sync_status = $1,
			sync_error = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE annict_id = $3
	`, string(constants.SyncStatusFailed), err.Error(), annictID)

	if updateErr != nil && c.logger != nil {
		c.logger.Error("Failed to update sync failure status",
			"annict_id", annictID,
			"error", updateErr)
	}
}
