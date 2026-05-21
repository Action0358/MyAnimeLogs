package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"myanimelogs/internal/common"
	"myanimelogs/internal/models"
)

// AdminAnimeService defines the minimal interface needed by AdminHandler
type AdminAnimeService interface {
	SyncGenresForAllAnimes() (int, int, []string, error)
	ClearAnimeCache() // キャッシュクリア
	SyncAllAnimeStatuses(ctx context.Context) (int, int, error)
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

	successCount, failureCount, err := h.animeService.SyncAllAnimeStatuses(c.Request.Context())
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
