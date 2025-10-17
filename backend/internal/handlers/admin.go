package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"myanimelogs/internal/common"
	"myanimelogs/internal/models"
)

// AnimeServiceInterface defines the interface for anime service operations
type AnimeServiceInterface interface {
	SyncEpisodeDurations(animeID *int) (int, int, int, int, []string, error)
	SearchAnimes(params models.AnimeSearchParams, userID *int) (*models.PaginatedResponse, error)
	AdvancedSearchAnimes(params models.AdvancedSearchParams, userID *int) (*models.PaginatedResponse, error)
	GetAnimeByID(animeID int, userID *int) (*models.AnimeWithUserData, error)
	GetAnimeEpisodes(animeID int, userID *int) ([]models.Episode, error)
	GetUserFavorites(userID int, status string) ([]models.AnimeWithUserData, error)
	UpdateFavorite(userID, animeID int, status string) (*models.UserFavorite, error)
	DeleteFavorite(userID, animeID int) error
	UpdateWatchStatus(userID, episodeID int, viewed bool) (*models.WatchStatus, error)
	UpdateRating(userID, animeID int, rating float64, comment *string) (*models.UserRating, error)
	DeleteRating(userID, animeID int) error
	GetRecentlyWatched(userID, limit int) ([]models.WatchHistory, error)
	GetFilterOptions() (*models.FilterOptions, error)
	GetPopularSearches(limit int) ([]string, error)
	GetSearchSuggestions(query string, limit int) ([]string, error)
}

// AdminHandler handles admin-related requests
type AdminHandler struct {
	*BaseHandler
	animeService AnimeServiceInterface
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(animeService AnimeServiceInterface, logger *common.Logger, errorMonitor *common.ErrorMonitor) *AdminHandler {
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
	totalAnimes, successCount, failureCount, skippedCount, errors, err := h.animeService.SyncEpisodeDurations(req.AnimeID)
	if err != nil {
		h.logger.Error("Failed to sync episode durations", "error", err)
		h.response.InternalServerError(c, "Failed to sync episode durations")
		return
	}

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
