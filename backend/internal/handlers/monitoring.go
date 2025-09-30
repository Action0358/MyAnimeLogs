package handlers

import (
	"net/http"
	"strconv"

	"myanimelogs/internal/common"

	"github.com/gin-gonic/gin"
)

// MonitoringHandler handles monitoring and health check endpoints
type MonitoringHandler struct {
	*BaseHandler
	errorMonitor *common.ErrorMonitor
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(logger *common.Logger, errorMonitor *common.ErrorMonitor) *MonitoringHandler {
	return &MonitoringHandler{
		BaseHandler:  NewBaseHandlerWithLogger(logger, errorMonitor),
		errorMonitor: errorMonitor,
	}
}

// GetErrorMetrics returns current error metrics
func (h *MonitoringHandler) GetErrorMetrics(c *gin.Context) {
	metrics := h.errorMonitor.GetErrorMetrics()
	statistics := h.errorMonitor.GetErrorStatistics()

	h.response.Success(c, gin.H{
		"metrics":    metrics,
		"statistics": statistics,
	})
}

// GetTopErrors returns the most frequent errors
func (h *MonitoringHandler) GetTopErrors(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	topErrors := h.errorMonitor.GetTopErrors(limit)

	// Calculate total error count for consistency
	totalErrors := 0
	for _, errorMetric := range topErrors {
		totalErrors += errorMetric.Count
	}

	h.response.Success(c, gin.H{
		"top_errors":     topErrors,
		"limit":          limit,
		"returned_count": len(topErrors),
		"total_errors":   totalErrors,
	})
}

// ResetErrorMetrics resets all error metrics (admin only)
func (h *MonitoringHandler) ResetErrorMetrics(c *gin.Context) {
	h.errorMonitor.ResetMetrics()
	h.response.Success(c, gin.H{
		"message": "Error metrics have been reset",
	})
}

// HealthCheck returns application health status
func (h *MonitoringHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "MyAnimeLogs API",
		"version": "1.0.0",
	})
}

// DetailedHealthCheck returns detailed health information
func (h *MonitoringHandler) DetailedHealthCheck(c *gin.Context) {
	statistics := h.errorMonitor.GetErrorStatistics()
	totalErrors := statistics["total_errors"].(int)
	errorTypeCount := statistics["error_type_count"].(int)

	healthScore := "healthy"
	if totalErrors > 100 {
		healthScore = "degraded"
	}
	if totalErrors > 500 {
		healthScore = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"service":      "MyAnimeLogs API",
		"version":      "1.0.0",
		"health_score": healthScore,
		"total_errors": totalErrors,
		"error_types":  errorTypeCount,
		"uptime":       "calculated_uptime", // TODO: implement uptime tracking
	})
}
