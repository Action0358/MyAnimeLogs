package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"myanimelogs/internal/common"
	"myanimelogs/internal/models"
	"myanimelogs/internal/services"
)

// AdminAnimeService defines the minimal interface needed by AdminHandler
type AdminAnimeService interface {
	SyncEpisodeDurations(animeID *int) (int, int, int, int, []string, error)
	SyncEpisodesFromJikan(animeID *int) (int, int, []string, error)
	GetEpisodeSyncStatus(animeID int) (*models.EpisodeSyncStatus, error)
	SyncGenresForAllAnimes() (int, int, []string, error)
	ClearAnimeCache() // キャッシュクリア
	SyncAllAnimeStatuses() (int, int, error)
	SyncReleasedOnFromJikan() (int, int, int, error)
}

// SyncScheduler defines the interface for sync scheduler
type SyncScheduler interface {
	GetStatus() models.SyncStatus
}

// AdminHandler handles admin-related requests
type AdminHandler struct {
	*BaseHandler
	animeService  AdminAnimeService
	syncScheduler SyncScheduler
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(animeService AdminAnimeService, logger *common.Logger, errorMonitor *common.ErrorMonitor, syncScheduler SyncScheduler) *AdminHandler {
	return &AdminHandler{
		BaseHandler:   NewBaseHandlerWithLogger(logger, errorMonitor),
		animeService:  animeService,
		syncScheduler: syncScheduler,
	}
}

// SyncEpisodeDurationsRequest represents the request body for syncing episode durations
type SyncEpisodeDurationsRequest struct {
	AnimeID *int `json:"anime_id"`
}

// SyncEpisodeDurationsResponse represents the response for syncing episode durations
type SyncEpisodeDurationsResponse struct {
	TotalAnimes  int      `json:"total_animes"`
	SuccessCount int      `json:"success_count"`
	FailureCount int      `json:"failure_count"`
	SkippedCount int      `json:"skipped_count"`
	Errors       []string `json:"errors,omitempty"`
}

// SyncEpisodeDurations synchronizes episode durations from Jikan API
// POST /api/admin/sync/episode-durations
func (h *AdminHandler) SyncEpisodeDurations(c *gin.Context) {
	var req SyncEpisodeDurationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.ValidationErrorWithErr(c, err)
		return
	}

	// Log the sync request
	if req.AnimeID != nil {
		h.logger.Info("Starting episode duration sync for specific anime", "anime_id", *req.AnimeID)
	} else {
		h.logger.Info("Starting episode duration sync for all animes")
	}

	// Perform sync
	totalAnimes, successCount, failureCount, skippedCount, errs, err := h.animeService.SyncEpisodeDurations(req.AnimeID)
	if err != nil {
		h.logger.Error("Failed to sync episode durations", "error", err)

		// Check for specific error types
		if errors.Is(err, services.ErrAnimeNotFound) {
			h.response.NotFoundError(c, "指定されたアニメが見つかりません")
			return
		}

		h.response.InternalServerError(c, "Failed to sync episode durations")
		return
	}

	response := SyncEpisodeDurationsResponse{
		TotalAnimes:  totalAnimes,
		SuccessCount: successCount,
		FailureCount: failureCount,
		SkippedCount: skippedCount,
		Errors:       errs,
	}

	h.logger.Info("Episode duration sync completed",
		"total_animes", totalAnimes,
		"success_count", successCount,
		"failure_count", failureCount)

	c.JSON(http.StatusOK, response)
}

// SyncEpisodesRequest represents the request body
type SyncEpisodesRequest struct {
	AnimeID *int `json:"anime_id"`
}

// SyncEpisodesResponse represents the response
type SyncEpisodesResponse struct {
	SuccessCount int      `json:"success_count"`
	FailureCount int      `json:"failure_count"`
	Errors       []string `json:"errors,omitempty"`
	Message      string   `json:"message"`
}

// SyncEpisodesFromJikan syncs episodes from Jikan API
// POST /api/admin/sync/episodes
func (h *AdminHandler) SyncEpisodesFromJikan(c *gin.Context) {
	var req SyncEpisodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.ValidationErrorWithErr(c, err)
		return
	}

	h.logger.Info("Starting episode sync from Jikan API")

	successCount, failureCount, errs, err := h.animeService.SyncEpisodesFromJikan(req.AnimeID)
	if err != nil {
		h.logger.Error("Failed to sync episodes from Jikan", "error", err)
		h.response.InternalServerError(c, "Failed to sync episodes")
		return
	}

	response := SyncEpisodesResponse{
		SuccessCount: successCount,
		FailureCount: failureCount,
		Errors:       errs,
		Message:      fmt.Sprintf("Episode sync completed: %d succeeded, %d failed", successCount, failureCount),
	}

	h.logger.Info("Episode sync from Jikan completed",
		"success_count", successCount,
		"failure_count", failureCount)

	c.JSON(http.StatusOK, response)
}

// GetSyncStatus returns the current sync status
// GET /api/admin/sync/status
func (h *AdminHandler) GetSyncStatus(c *gin.Context) {
	if h.syncScheduler == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Sync scheduler is not available",
		})
		return
	}

	status := h.syncScheduler.GetStatus()
	c.JSON(http.StatusOK, status)
}

// GetEpisodeSyncStatusRequest represents the request for checking episode sync status
type GetEpisodeSyncStatusRequest struct {
	AnimeID int `form:"anime_id" binding:"required"`
}

// GetEpisodeSyncStatus returns the sync status for a specific anime's episodes
// GET /api/admin/episode-sync-status?anime_id=69
func (h *AdminHandler) GetEpisodeSyncStatus(c *gin.Context) {
	var req GetEpisodeSyncStatusRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.response.ValidationErrorWithErr(c, err)
		return
	}

	status, err := h.animeService.GetEpisodeSyncStatus(req.AnimeID)
	if err != nil {
		h.logger.Error("Failed to get episode sync status", "anime_id", req.AnimeID, "error", err)

		if errors.Is(err, services.ErrAnimeNotFound) {
			h.response.NotFoundError(c, "アニメが見つかりません")
			return
		}

		h.response.InternalServerError(c, "エピソード同期状況の取得に失敗しました")
		return
	}

	c.JSON(http.StatusOK, status)
}

// SyncGenresResponse represents the response for syncing genres
type SyncGenresResponse struct {
	SuccessCount int      `json:"success_count"`
	FailureCount int      `json:"failure_count"`
	Errors       []string `json:"errors,omitempty"`
	Message      string   `json:"message"`
}

// SyncGenresForAllAnimes syncs genre information for all animes from AniList API
// POST /api/admin/sync/genres
func (h *AdminHandler) SyncGenresForAllAnimes(c *gin.Context) {
	h.logger.Info("Starting genre sync for all animes")

	successCount, failureCount, errs, err := h.animeService.SyncGenresForAllAnimes()
	if err != nil {
		h.logger.Error("Failed to sync genres", "error", err)
		h.response.InternalServerError(c, "Failed to sync genres")
		return
	}

	response := SyncGenresResponse{
		SuccessCount: successCount,
		FailureCount: failureCount,
		Errors:       errs,
		Message:      fmt.Sprintf("Genre sync completed: %d succeeded, %d failed", successCount, failureCount),
	}

	h.logger.Info("Genre sync completed",
		"success_count", successCount,
		"failure_count", failureCount)

	c.JSON(http.StatusOK, response)
}

// ClearCache clears all anime-related caches
// POST /api/admin/cache/clear
func (h *AdminHandler) ClearCache(c *gin.Context) {
	h.logger.Info("Clearing anime cache via admin API")

	h.animeService.ClearAnimeCache()

	c.JSON(http.StatusOK, gin.H{
		"message": "Cache cleared successfully",
	})
}

// SyncAnimeStatuses - 全アニメのステータスを同期
// POST /api/admin/sync/statuses
func (h *AdminHandler) SyncAnimeStatuses(c *gin.Context) {
	h.logger.Info("Starting anime status sync via admin API")

	successCount, failureCount, err := h.animeService.SyncAllAnimeStatuses()
	if err != nil {
		h.logger.Error("Failed to sync anime statuses", "error", err)
		h.response.InternalServerError(c, "ステータス同期に失敗しました")
		return
	}

	h.response.Success(c, gin.H{
		"success_count": successCount,
		"failure_count": failureCount,
		"message":       fmt.Sprintf("ステータス同期完了: 成功 %d件, 失敗 %d件", successCount, failureCount),
	})
}

// SyncReleasedOn - Jikan APIからreleased_onを一括取得
// POST /api/admin/sync/released-on
func (h *AdminHandler) SyncReleasedOn(c *gin.Context) {
	h.logger.Info("Starting released_on sync from Jikan API via admin API")

	successCount, notFoundCount, failureCount, err := h.animeService.SyncReleasedOnFromJikan()
	if err != nil {
		h.logger.Error("Failed to sync released_on", "error", err)
		h.response.InternalServerError(c, "released_on同期に失敗しました")
		return
	}

	h.response.Success(c, gin.H{
		"success_count":   successCount,
		"not_found_count": notFoundCount,
		"failure_count":   failureCount,
		"message":         fmt.Sprintf("released_on同期完了: 成功 %d件, 未検出 %d件, 失敗 %d件", successCount, notFoundCount, failureCount),
	})
}
