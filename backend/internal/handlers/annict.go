package handlers

import (
	"myanimelogs/internal/services"
	"net/http"

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
func (h *AnnictHandler) GetWorksBySeason(c *gin.Context) {
	season := c.Query("season")
	if season == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "season query parameter is required"})
		return
	}

	works, err := h.annictService.FetchWorksBySeason(season)
	if err != nil {
		// In a real app, you might want more sophisticated error handling
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// The response from the service is already a struct that can be marshaled to JSON
	c.JSON(http.StatusOK, works)
}

// GetPopularWorks handles requests to get popular anime works.
func (h *AnnictHandler) GetPopularWorks(c *gin.Context) {
	works, err := h.annictService.FetchPopularWorks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, works)
}

// TestConnection tests the Annict API connection
func (h *AnnictHandler) TestConnection(c *gin.Context) {
	err := h.annictService.TestConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Annict API connection successful",
	})
}
