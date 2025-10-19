package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"myanimelogs/internal/common"
	"myanimelogs/internal/services"
)

// AdminAnimeService defines the minimal interface needed by AdminHandler
type AdminAnimeService interface {
	SyncEpisodeDurations(animeID *int) (int, int, int, int, []string, error)
}

// AdminHandler handles admin-related requests
type AdminHandler struct {
	*BaseHandler
	animeService AdminAnimeService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(animeService AdminAnimeService, logger *common.Logger, errorMonitor *common.ErrorMonitor) *AdminHandler {
	return &AdminHandler{
		BaseHandler:  NewBaseHandlerWithLogger(logger, errorMonitor),
		animeService: animeService,
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
