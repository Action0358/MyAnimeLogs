package handlers

import (
	"context"
	"fmt"
	"myanimelogs/internal/common"
	"myanimelogs/internal/models"
	"myanimelogs/internal/services"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/singleflight"
)

type AnnictHandler struct {
	*BaseHandler
	annictService *services.AnnictClient
	anilistClient *services.AniListClient
	imageCache    map[string]*services.CoverImage // MAL ID -> CoverImage のキャッシュ
	cacheMutex    sync.RWMutex
	sfGroup       singleflight.Group // キャッシュスタンピード対策
}

func NewAnnictHandler(annictService *services.AnnictClient, logger *common.Logger, errorMonitor *common.ErrorMonitor) *AnnictHandler {
	return &AnnictHandler{
		BaseHandler:   NewBaseHandlerWithLogger(logger, errorMonitor),
		annictService: annictService,
		anilistClient: services.NewAniListClient(),
		imageCache:    make(map[string]*services.CoverImage),
	}
}

// GetWorksBySeason handles requests to get anime works by season.
// It expects a 'season' query parameter, e.g., ?season=2024-autumn
// If no season is provided, it defaults to the current season
func (h *AnnictHandler) GetWorksBySeason(c *gin.Context) {
	season := c.Query("season")
	if season == "" {
		// Default to current season if not specified
		season = getCurrentSeason()
	}

	var works *models.AnnictWorkResponse // 明示的な型宣言
	works, err := h.annictService.FetchWorksBySeason(season)
	if err != nil {
		h.logger.Error("Failed to fetch works by season", "season", season, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch anime works by season"})
		return
	}

	// Anilistの画像を並列で取得
	h.enrichWorksWithAnilistImages(works.Works)

	// The response from the service is already a struct that can be marshaled to JSON
	c.JSON(http.StatusOK, works)
}

// getCurrentSeason returns the current season in the format expected by Annict API
func getCurrentSeason() string {
	now := time.Now()
	year := now.Year()
	month := now.Month()

	var season string
	switch {
	case month >= time.January && month <= time.March:
		season = "winter"
	case month >= time.April && month <= time.June:
		season = "spring"
	case month >= time.July && month <= time.September:
		season = "summer"
	case month >= time.October && month <= time.December:
		season = "autumn"
	}

	return fmt.Sprintf("%d-%s", year, season)
}

// GetPopularWorks handles requests to get popular anime works.
func (h *AnnictHandler) GetPopularWorks(c *gin.Context) {
	var works *models.AnnictWorkResponse // 明示的な型宣言
	works, err := h.annictService.FetchPopularWorks()
	if err != nil {
		h.logger.Error("Failed to fetch popular works", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch popular anime works"})
		return
	}

	// Anilistの画像を並列で取得
	h.enrichWorksWithAnilistImages(works.Works)

	c.JSON(http.StatusOK, works)
}

// TestConnection tests the Annict API connection
func (h *AnnictHandler) TestConnection(c *gin.Context) {
	err := h.annictService.TestConnection()
	if err != nil {
		h.logger.Error("Failed to test Annict API connection", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to connect to Annict API",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Annict API connection successful",
	})
}

// SyncWorksBySeason syncs anime works by season to database
func (h *AnnictHandler) SyncWorksBySeason(c *gin.Context) {
	season := c.Query("season")
	if season == "" {
		season = getCurrentSeason()
	}

	animes, err := h.annictService.FetchAndSyncWorksBySeason(season)
	if err != nil {
		h.logger.Error("Failed to sync works by season", "season", season, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync anime works by season"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Successfully synced %d anime works for season %s", len(animes), season),
		"data":    animes,
	})
}

// SyncPopularWorks syncs popular anime works to database
func (h *AnnictHandler) SyncPopularWorks(c *gin.Context) {
	animes, err := h.annictService.FetchAndSyncPopularWorks()
	if err != nil {
		h.logger.Error("Failed to sync popular works", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync popular anime works"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Successfully synced %d popular anime works", len(animes)),
		"data":    animes,
	})
}

// SyncEpisodes syncs episodes for a specific anime
func (h *AnnictHandler) SyncEpisodes(c *gin.Context) {
	animeIDStr := c.Param("anime_id")
	animeID, err := strconv.Atoi(animeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid anime ID"})
		return
	}

	// For now, create sample episodes since Annict API doesn't provide episode details in the basic endpoint
	episodes, err := h.createSampleEpisodes(animeID)
	if err != nil {
		h.logger.Error("Failed to create sample episodes", "anime_id", animeID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create episodes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Successfully created %d episodes for anime %d", len(episodes), animeID),
		"data":    episodes,
	})
}

// createSampleEpisodes creates sample episodes for an anime
func (h *AnnictHandler) createSampleEpisodes(animeID int) ([]models.Episode, error) {
	// This is a simplified implementation - in a real scenario, you'd fetch from Annict API
	episodes := []models.Episode{}

	// Create 12 sample episodes (typical for a season)
	for i := 1; i <= 12; i++ {
		episode := models.Episode{
			AnimeID:       animeID,
			EpisodeNumber: i,
			Title:         fmt.Sprintf("第%d話", i),
		}
		episodes = append(episodes, episode)
	}

	return episodes, nil
}

// enrichWorksWithAnilistImages - Anilistから画像を取得してAnnictWorksを拡張
func (h *AnnictHandler) enrichWorksWithAnilistImages(works []models.AnnictWork) {
	// 最初の10件のみ同期的に取得、残りはキャッシュのみ使用
	const priorityCount = 10
	const maxConcurrent = 5

	type result struct {
		index int
		image *services.CoverImage
	}

	// 優先度の高いアイテム（最初の10件）を処理
	priorityResults := make(chan result, priorityCount)
	semaphore := make(chan struct{}, maxConcurrent)

	priorityLimit := len(works)
	if priorityLimit > priorityCount {
		priorityLimit = priorityCount
	}

	for i := 0; i < priorityLimit; i++ {
		go func(index int) {
			var coverImage *services.CoverImage

			// パニックハンドリング: デッドロックを防ぐ
			defer func() {
				if r := recover(); r != nil {
					h.logger.Error("Panic in getCoverImage", "index", index, "panic", r)
					// パニックが発生してもnilを送信してデッドロックを防ぐ
					coverImage = nil
				}
				// 必ず結果を送信（パニック時もnilを送信）
				priorityResults <- result{index: index, image: coverImage}
			}()

			// セマフォを取得
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			work := &works[index]
			coverImage = h.getCoverImage(work)
		}(i)
	}

	// 優先度の高いアイテムの結果を収集
	for i := 0; i < priorityLimit; i++ {
		res := <-priorityResults
		h.applyCoverImage(&works[res.index], res.image)
	}
	close(priorityResults)

	// 残りのアイテムはキャッシュのみを同期的にチェック
	for i := priorityLimit; i < len(works); i++ {
		work := &works[i]
		if work.MalAnimeID != nil && *work.MalAnimeID != "" {
			h.cacheMutex.RLock()
			cachedImage, exists := h.imageCache[*work.MalAnimeID]
			h.cacheMutex.RUnlock()

			if exists {
				h.applyCoverImage(work, cachedImage)
			}
		}
	}
}

// getCoverImage - 2層キャッシュ（インメモリ + データベース）から画像を取得
// singleflightパターンでキャッシュスタンピードを防止
func (h *AnnictHandler) getCoverImage(work *models.AnnictWork) *services.CoverImage {
	// レイヤー1: インメモリキャッシュをチェック（超高速: 0ms）
	cacheKey := ""
	if work.MalAnimeID != nil && *work.MalAnimeID != "" {
		cacheKey = *work.MalAnimeID
		h.cacheMutex.RLock()
		cachedImage, exists := h.imageCache[cacheKey]
		h.cacheMutex.RUnlock()

		if exists {
			return cachedImage
		}
	}

	// Singleflightパターンを使用して、同じキーに対する重複リクエストを防止
	// 複数のgoroutineが同時に同じ画像を要求しても、1回だけ取得される
	if cacheKey != "" {
		result, _, _ := h.sfGroup.Do(cacheKey, func() (interface{}, error) {
			return h.fetchCoverImage(work, cacheKey), nil
		})

		if result != nil {
			return result.(*services.CoverImage)
		}
	}

    // MAL IDがない場合は、タイトルで検索（singleflightを使用）
    if work.Title != "" {
        titleCacheKey := "title:" + work.Title
        result, _, _ := h.sfGroup.Do(titleCacheKey, func() (interface{}, error) {
            return h.fetchCoverImageByTitle(work), nil
        })

        if result != nil {
            return result.(*services.CoverImage)
        }
    }

	return nil
}

// fetchCoverImage - データベースまたはAPIから画像を取得（singleflight内で実行）
func (h *AnnictHandler) fetchCoverImage(work *models.AnnictWork, cacheKey string) *services.CoverImage {
	var coverImage *services.CoverImage

	// レイヤー2: データベースをチェック（高速: 5-10ms）
	if h.annictService.DB != nil {
		var large, medium string
		var updatedAt *time.Time

		err := h.annictService.DB.QueryRow(`
			SELECT cover_image_large, cover_image_medium, cover_image_updated_at
			FROM animes 
			WHERE mal_anime_id = $1 
			AND cover_image_large IS NOT NULL
			AND cover_image_updated_at > NOW() - INTERVAL '7 days'
		`, cacheKey).Scan(&large, &medium, &updatedAt)

		if err == nil && large != "" {
			// データベースから取得成功
			coverImage = &services.CoverImage{
				Large:  large,
				Medium: medium,
			}

			// インメモリキャッシュに保存（次回さらに高速化）
			h.cacheMutex.Lock()
			h.imageCache[cacheKey] = coverImage
			h.cacheMutex.Unlock()

			return coverImage
		}
	}

	// レイヤー3: Anilist APIから取得（遅い: 100-150ms）
	coverImage, err := h.anilistClient.GetCoverImageByMALId(cacheKey)
	if err != nil {
		h.logger.Warn("Failed to fetch cover image from AniList API", "mal_anime_id", cacheKey, "error", err)
		return nil
	}

	if coverImage != nil {
		// データベースに保存（永続化）
		// context.Background()を使用（バックグラウンド処理のため）
		h.saveCoverImageToDB(context.Background(), cacheKey, coverImage)

		// インメモリキャッシュに保存
		h.cacheMutex.Lock()
		h.imageCache[cacheKey] = coverImage
		h.cacheMutex.Unlock()

		return coverImage
	}

	return nil
}

// fetchCoverImageByTitle - タイトルで画像を取得（フォールバック）
func (h *AnnictHandler) fetchCoverImageByTitle(work *models.AnnictWork) *services.CoverImage {
	coverImage, err := h.anilistClient.GetCoverImageByTitle(work.Title)
	if err != nil {
		h.logger.Warn("Failed to fetch cover image by title from AniList API", "title", work.Title, "error", err)
		return nil
	}

	if coverImage != nil {
		// Anilistから返されたMAL IDを使用してデータベースに保存
		if coverImage.MalId != nil {
			malIdStr := strconv.Itoa(*coverImage.MalId)
			// context.Background()を使用（バックグラウンド処理のため）
			h.saveCoverImageToDB(context.Background(), malIdStr, coverImage)

			// インメモリキャッシュに保存
			h.cacheMutex.Lock()
			h.imageCache[malIdStr] = coverImage
			h.cacheMutex.Unlock()
		}
		return coverImage
	}

	return nil
}

// saveCoverImageToDB - 画像URLをデータベースに保存
func (h *AnnictHandler) saveCoverImageToDB(ctx context.Context, malAnimeID string, coverImage *services.CoverImage) {
	if h.annictService.DB == nil {
		return
	}

	// 非同期で保存（メインの処理をブロックしない）
	go func() {
		// UPDATEのみ実行（アニメがデータベースに存在する場合のみ）
		result, err := h.annictService.DB.ExecContext(ctx, `
			UPDATE animes 
			SET cover_image_large = $1,
				cover_image_medium = $2,
				cover_image_updated_at = NOW()
			WHERE mal_anime_id = $3
		`, coverImage.Large, coverImage.Medium, malAnimeID)

		if err != nil {
			h.logger.Error("Failed to save cover image to DB", "mal_anime_id", malAnimeID, "error", err)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			h.logger.Error("Failed to get rows affected", "mal_anime_id", malAnimeID, "error", err)
			return
		}

		if rowsAffected > 0 {
			h.logger.Info("Saved cover image to DB", "mal_anime_id", malAnimeID)
		}
		// アニメが存在しない場合はログを出力しない（Annict同期前の正常な状態）
	}()
}

// applyCoverImage - 画像をworksに適用
func (h *AnnictHandler) applyCoverImage(work *models.AnnictWork, image *services.CoverImage) {
	if image != nil {
		// ExtraLargeがあればそれを使用、なければLargeを使用
		if image.ExtraLarge != "" {
			work.CoverImageLarge = image.ExtraLarge
		} else {
			work.CoverImageLarge = image.Large
		}
		work.CoverImageMedium = image.Medium
	}
}
