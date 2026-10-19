package handlers

import (
	"net/http"
	"strconv"

	"myanimelogs/internal/common"

	"github.com/gin-gonic/gin"
)

// MonitoringHandler 監視とヘルスチェックエンドポイントを処理
type MonitoringHandler struct {
	*BaseHandler
	errorMonitor *common.ErrorMonitor
}

// NewMonitoringHandler 新しい監視ハンドラーを作成
func NewMonitoringHandler(logger *common.Logger, errorMonitor *common.ErrorMonitor) *MonitoringHandler {
	return &MonitoringHandler{
		BaseHandler:  NewBaseHandlerWithLogger(logger, errorMonitor),
		errorMonitor: errorMonitor,
	}
}

// GetErrorMetrics 現在のエラーメトリクスを返す
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

// GetTopErrors 最も頻繁に発生するエラーを返す
func (h *MonitoringHandler) GetTopErrors(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	// ソートパラメータを取得
	sortBy := c.DefaultQuery("sort", "count")

	// ヘルパー関数を使用して文字列をSortBy型に変換
	sortCriteria := common.ToSortBy(sortBy)

	topErrors := h.errorMonitor.GetTopErrorsWithSort(limit, sortCriteria)

	// 返されたトップリストの件数を計算
	countInTopList := 0
	for _, errorMetric := range topErrors {
		countInTopList += errorMetric.Count
	}

	// システム全体のエラー総数を取得
	systemTotalErrors := h.errorMonitor.GetTotalErrorCount()

	// 元のパラメータではなく、実際に使用されたソート基準を返す
	actualSortBy := string(sortCriteria)

	h.response.Success(c, gin.H{
		"top_errors":        topErrors,
		"limit":             limit,
		"sort_by":           actualSortBy,
		"returned_count":    len(topErrors),
		"count_in_top_list": countInTopList,
		"total_errors":      systemTotalErrors,
	})
}

// ResetErrorMetrics すべてのエラーメトリクスをリセット（管理者のみ）
func (h *MonitoringHandler) ResetErrorMetrics(c *gin.Context) {
	h.errorMonitor.ResetMetrics()
	h.response.Success(c, gin.H{
		"message": "Error metrics have been reset",
	})
}

// HealthCheck アプリケーションのヘルスステータスを返す
func (h *MonitoringHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "MyAnimeLogs API",
		"version": "1.0.0",
	})
}
