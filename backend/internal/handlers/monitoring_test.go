package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"myanimelogs/internal/common"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMonitoringHandler_GetTopErrors_ResponseFields(t *testing.T) {
	// Setup
	logger := common.NewLogger(common.LogLevelDebug, "text")
	monitor := common.NewErrorMonitor(logger, 10, 5*time.Minute)
	handler := NewMonitoringHandler(logger, monitor)

	// テスト用のエラーを記録
	ctx := context.Background()

	// ERROR_A: 5回
	for i := 0; i < 5; i++ {
		monitor.RecordError(ctx, "ERROR_A", errors.New("error a"), map[string]interface{}{})
	}

	// ERROR_B: 3回
	for i := 0; i < 3; i++ {
		monitor.RecordError(ctx, "ERROR_B", errors.New("error b"), map[string]interface{}{})
	}

	// ERROR_C: 2回
	for i := 0; i < 2; i++ {
		monitor.RecordError(ctx, "ERROR_C", errors.New("error c"), map[string]interface{}{})
	}

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/admin/errors/top", handler.GetTopErrors)

	// Create request
	req, _ := http.NewRequest("GET", "/api/admin/errors/top?limit=2&sort=count", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response status
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert response structure
	assert.Contains(t, response, "success")
	assert.Equal(t, true, response["success"])
	assert.Contains(t, response, "data")

	data := response["data"].(map[string]interface{})

	// Assert expected fields exist
	assert.Contains(t, data, "top_errors")
	assert.Contains(t, data, "limit")
	assert.Contains(t, data, "sort_by")
	assert.Contains(t, data, "returned_count")
	assert.Contains(t, data, "count_in_top_list")
	assert.Contains(t, data, "total_errors")

	// Assert field values
	assert.Equal(t, float64(2), data["limit"])
	assert.Equal(t, "count", data["sort_by"])
	assert.Equal(t, float64(2), data["returned_count"])
	assert.Equal(t, float64(8), data["count_in_top_list"]) // ERROR_A(5) + ERROR_B(3) = 8
	assert.Equal(t, float64(10), data["total_errors"])     // ERROR_A(5) + ERROR_B(3) + ERROR_C(2) = 10

	// Assert top_errors structure
	topErrors := data["top_errors"].([]interface{})
	assert.Len(t, topErrors, 2)

	// Verify first error (should be ERROR_A with count 5)
	firstError := topErrors[0].(map[string]interface{})
	assert.Equal(t, "ERROR_A", firstError["error_code"])
	assert.Equal(t, float64(5), firstError["count"])

	// Verify second error (should be ERROR_B with count 3)
	secondError := topErrors[1].(map[string]interface{})
	assert.Equal(t, "ERROR_B", secondError["error_code"])
	assert.Equal(t, float64(3), secondError["count"])

	// Assert count_in_top_list < total_errors
	countInTopList := data["count_in_top_list"].(float64)
	totalErrors := data["total_errors"].(float64)
	assert.Less(t, countInTopList, totalErrors)
}

func TestMonitoringHandler_GetTopErrors_SortParameters(t *testing.T) {
	// Setup
	logger := common.NewLogger(common.LogLevelDebug, "text")
	monitor := common.NewErrorMonitor(logger, 10, 5*time.Minute)
	handler := NewMonitoringHandler(logger, monitor)

	// Record test errors with time differences
	ctx := context.Background()
	monitor.RecordError(ctx, "ERROR_A", errors.New("error a"), map[string]interface{}{})

	time.Sleep(10 * time.Millisecond) // Ensure time difference
	monitor.RecordError(ctx, "ERROR_B", errors.New("error b"), map[string]interface{}{})

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/admin/errors/top", handler.GetTopErrors)

	tests := []struct {
		name           string
		sortParam      string
		expectedSortBy string
	}{
		{"Sort by count", "count", "count"},
		{"Sort by rate", "rate", "rate"},
		{"Sort by time", "time", "time"},
		{"Sort by code", "code", "code"},
		{"Default sort", "", "count"},
		{"Invalid sort", "invalid", "count"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/admin/errors/top"
			if tt.sortParam != "" {
				url += "?sort=" + tt.sortParam
			}

			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			data := response["data"].(map[string]interface{})
			assert.Equal(t, tt.expectedSortBy, data["sort_by"])
		})
	}
}

func TestMonitoringHandler_GetTopErrors_LimitParameter(t *testing.T) {
	// Setup
	logger := common.NewLogger(common.LogLevelDebug, "text")
	monitor := common.NewErrorMonitor(logger, 10, 5*time.Minute)
	handler := NewMonitoringHandler(logger, monitor)

	// Record multiple errors
	ctx := context.Background()
	errorCodes := []string{"ERROR_1", "ERROR_2", "ERROR_3", "ERROR_4", "ERROR_5"}
	for _, code := range errorCodes {
		monitor.RecordError(ctx, code, errors.New("test error"), map[string]interface{}{})
	}

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/admin/errors/top", handler.GetTopErrors)

	tests := []struct {
		name          string
		limitParam    string
		expectedLimit float64
		expectedCount int
	}{
		{"Limit 3", "3", 3, 3},
		{"Limit 10", "10", 10, 5}, // Only 5 errors exist
		{"Default limit", "", 10, 5},
		{"Invalid limit", "invalid", 10, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/admin/errors/top"
			if tt.limitParam != "" {
				url += "?limit=" + tt.limitParam
			}

			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			data := response["data"].(map[string]interface{})
			assert.Equal(t, tt.expectedLimit, data["limit"])

			topErrors := data["top_errors"].([]interface{})
			assert.Len(t, topErrors, tt.expectedCount)
		})
	}
}
