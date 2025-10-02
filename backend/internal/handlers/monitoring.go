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

	// 統計データの妥当性を検証
	if !statistics.IsValid() {
		h.response.InternalServerError(c, "統計データの計算に問題が発生しました")
		return
	}

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

	// Get sort parameter
	sortBy := c.DefaultQuery("sort", "count")

	// Convert string to SortBy type using helper function
	sortCriteria := common.ToSortBy(sortBy)

	topErrors := h.errorMonitor.GetTopErrorsWithSort(limit, sortCriteria)

	// Calculate count in the returned top list
	countInTopList := 0
	for _, errorMetric := range topErrors {
		countInTopList += errorMetric.Count
	}

	// Get system-wide total error count
	systemTotalErrors := h.errorMonitor.GetTotalErrorCount()

	h.response.Success(c, gin.H{
		"top_errors":        topErrors,
		"limit":             limit,
		"sort_by":           sortBy,
		"returned_count":    len(topErrors),
		"count_in_top_list": countInTopList,
		"total_errors":      systemTotalErrors,
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

	// 型安全な構造体から直接値を取得
	totalErrors := statistics.TotalErrors
	errorTypeCount := statistics.ErrorTypeCount

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
