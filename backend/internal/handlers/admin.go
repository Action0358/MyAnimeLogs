package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"myanimelogs/internal/common"
	"myanimelogs/internal/services"
)

// AdminAnimeService defines the minimal interface needed by AdminHandler
type AdminAnimeService interface {
	SyncEpisodeDurations(animeID *int) (int, int, int, int, []string, error)
	CreateEpisodesForAllAnimes() (int, int, error)
	EnrichEpisodesWithTitles(animeID *int) (int, int, []string, error)
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

// CreateEpisodesResponse represents the response for creating episodes
type CreateEpisodesResponse struct {
	SuccessCount int    `json:"success_count"`
	FailureCount int    `json:"failure_count"`
	Message      string `json:"message"`
}

// CreateEpisodesForAllAnimes creates episodes for all animes that don't have episodes yet
// POST /api/admin/create-episodes
func (h *AdminHandler) CreateEpisodesForAllAnimes(c *gin.Context) {
	h.logger.Info("Starting episode creation for all animes")

	successCount, failureCount, err := h.animeService.CreateEpisodesForAllAnimes()
	if err != nil {
		h.logger.Error("Failed to create episodes", "error", err)
		h.response.InternalServerError(c, "Failed to create episodes")
		return
	}

	response := CreateEpisodesResponse{
		SuccessCount: successCount,
		FailureCount: failureCount,
		Message:      fmt.Sprintf("Episodes creation completed: %d succeeded, %d failed", successCount, failureCount),
	}

	h.logger.Info("Episode creation completed",
		"success_count", successCount,
		"failure_count", failureCount)

	c.JSON(http.StatusOK, response)
}

// EnrichEpisodeTitlesRequest represents the request body
type EnrichEpisodeTitlesRequest struct {
	AnimeID *int `json:"anime_id"`
}

// EnrichEpisodeTitlesResponse represents the response
type EnrichEpisodeTitlesResponse struct {
	SuccessCount int      `json:"success_count"`
	FailureCount int      `json:"failure_count"`
	Errors       []string `json:"errors,omitempty"`
	Message      string   `json:"message"`
}

// EnrichEpisodeTitles enriches episode titles from Annict API
// POST /api/admin/enrich/episode-titles
func (h *AdminHandler) EnrichEpisodeTitles(c *gin.Context) {
	var req EnrichEpisodeTitlesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.ValidationErrorWithErr(c, err)
		return
	}

	h.logger.Info("Starting episode titles enrichment")

	successCount, failureCount, errs, err := h.animeService.EnrichEpisodesWithTitles(req.AnimeID)
	if err != nil {
		h.logger.Error("Failed to enrich episode titles", "error", err)
		h.response.InternalServerError(c, "Failed to enrich episode titles")
		return
	}

	response := EnrichEpisodeTitlesResponse{
		SuccessCount: successCount,
		FailureCount: failureCount,
		Errors:       errs,
		Message:      fmt.Sprintf("Episode titles enrichment completed: %d succeeded, %d failed", successCount, failureCount),
	}

	h.logger.Info("Episode titles enrichment completed",
		"success_count", successCount,
		"failure_count", failureCount)

	c.JSON(http.StatusOK, response)
}
