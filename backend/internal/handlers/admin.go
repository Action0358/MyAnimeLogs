package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"myanimelogs/internal/common"
	"myanimelogs/internal/services"
)

// AdminHandler handles admin-related requests
type AdminHandler struct {
	*BaseHandler
	animeService *services.AnimeService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(animeService *services.AnimeService, logger *common.Logger, errorMonitor *common.ErrorMonitor) *AdminHandler {
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
	totalAnimes, successCount, failureCount, errors, err := h.animeService.SyncEpisodeDurations(req.AnimeID)
	if err != nil {
		h.logger.Error("Failed to sync episode durations", "error", err)
		h.response.InternalServerError(c, "Failed to sync episode durations")
		return
	}

	// Calculate skipped count (animes without MAL ID)
	skippedCount := 0

	response := SyncEpisodeDurationsResponse{
		TotalAnimes:  totalAnimes,
		SuccessCount: successCount,
		FailureCount: failureCount,
		SkippedCount: skippedCount,
		Errors:       errors,
	}

	h.logger.Info("Episode duration sync completed",
		"total_animes", totalAnimes,
		"success_count", successCount,
		"failure_count", failureCount)

	c.JSON(http.StatusOK, response)
}
