package handlers

import (
	"fmt"
	"log"
	"myanimelogs/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AnnictHandler struct {
	annictService *services.AnnictClient
}

func NewAnnictHandler(annictService *services.AnnictClient) *AnnictHandler {
	return &AnnictHandler{
		annictService: annictService,
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

	works, err := h.annictService.FetchWorksBySeason(season)
	if err != nil {
		log.Printf("Failed to fetch works by season: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch anime works by season"})
		return
	}

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
	works, err := h.annictService.FetchPopularWorks()
	if err != nil {
		log.Printf("Failed to fetch popular works: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch popular anime works"})
		return
	}
	c.JSON(http.StatusOK, works)
}

// TestConnection tests the Annict API connection
func (h *AnnictHandler) TestConnection(c *gin.Context) {
	err := h.annictService.TestConnection()
	if err != nil {
		log.Printf("Failed to test Annict API connection: %v", err)
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
		log.Printf("Failed to sync works by season: %v", err)
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
		log.Printf("Failed to sync popular works: %v", err)
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
		log.Printf("Failed to create sample episodes: %v", err)
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
			Title:          fmt.Sprintf("第%d話", i),
		}
		episodes = append(episodes, episode)
	}

	return episodes, nil
}
