package handlers

import (
	"log"
	"myanimelogs/internal/services"
	"net/http"

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
func (h *AnnictHandler) GetWorksBySeason(c *gin.Context) {
	season := c.Query("season")
	if season == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "season query parameter is required"})
		return
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
