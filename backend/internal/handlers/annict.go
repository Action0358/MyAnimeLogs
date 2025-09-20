package handlers

import (
	"fmt"
	"log"
	"myanimelogs/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AnnictHandler struct {
	annictService services.AnnictService
}

func NewAnnictHandler(annictService services.AnnictService) *AnnictHandler {
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
