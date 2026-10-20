package handlers

import (
	"fmt"
	"myanimelogs/internal/common"
	"myanimelogs/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AnnictHandler は Annict との同期と接続確認を担う。
//
// 以前は /api/annict/works・/api/annict/popular-works で Annict をそのまま中継していたが、
// 認証なしで外部APIを叩けて Annict のレート制限（5 req/秒）を使い切られるため撤去した。
// 利用者向けの一覧は同期済みのDBから /api/animes 系で返す。
type AnnictHandler struct {
	*BaseHandler
	annictService *services.AnnictClient
}

func NewAnnictHandler(annictService *services.AnnictClient, logger *common.Logger, errorMonitor *common.ErrorMonitor) *AnnictHandler {
	return &AnnictHandler{
		BaseHandler:   NewBaseHandlerWithLogger(logger, errorMonitor),
		annictService: annictService,
	}
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
