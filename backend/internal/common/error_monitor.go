package common

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"
)

// ErrorMonitor tracks and monitors application errors
type ErrorMonitor struct {
	logger         *Logger
	errorCounts    map[string]int
	errorRates     map[string][]time.Time
	mutex          sync.RWMutex
	alertThreshold int
	timeWindow     time.Duration
}

// ErrorMetrics represents error statistics
type ErrorMetrics struct {
	ErrorCode    string    `json:"error_code"`
	Count        int       `json:"count"`
	Rate         float64   `json:"rate_per_minute"`
	LastOccurred time.Time `json:"last_occurred"`
}

// ErrorStatistics represents comprehensive error statistics
type ErrorStatistics struct {
	TotalErrors     int       `json:"total_errors"`
	ErrorTypeCount  int       `json:"error_type_count"`
	TotalRate       float64   `json:"total_rate"`
	AverageRate     float64   `json:"average_rate"`
	MostRecentError time.Time `json:"most_recent_error"`
}

// SortBy represents different sorting options for error metrics
type SortBy string

const (
	SortByCount     SortBy = "count"
	SortByRate      SortBy = "rate"
	SortByTime      SortBy = "time"
	SortByErrorCode SortBy = "code"
)

// ToSortBy converts a string to SortBy type with fallback to default
func ToSortBy(s string) SortBy {
	switch s {
	case "rate":
		return SortByRate
	case "time":
		return SortByTime
	case "code":
		return SortByErrorCode
	case "count":
		return SortByCount
	default:
		return SortByCount // Default fallback
	}
}

// NewErrorMonitor creates a new error monitor
func NewErrorMonitor(logger *Logger, alertThreshold int, timeWindow time.Duration) *ErrorMonitor {
	return &ErrorMonitor{
		logger:         logger,
		errorCounts:    make(map[string]int),
		errorRates:     make(map[string][]time.Time),
		alertThreshold: alertThreshold,
		timeWindow:     timeWindow,
	}
}

// RecordError records an error occurrence
func (em *ErrorMonitor) RecordError(ctx context.Context, errorCode string, err error, metadata map[string]interface{}) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	now := time.Now()

	// Increment error count
	em.errorCounts[errorCode]++

	// Add to rate tracking
	if em.errorRates[errorCode] == nil {
		em.errorRates[errorCode] = make([]time.Time, 0)
	}
	em.errorRates[errorCode] = append(em.errorRates[errorCode], now)

	// Clean old entries (outside time window)
	em.cleanOldEntries(errorCode, now)

	// Log the error with metadata
	fields := []interface{}{
		"error_code", errorCode,
		"error_count", em.errorCounts[errorCode],
		"rate_count", len(em.errorRates[errorCode]),
	}

	// Add metadata fields
	for key, value := range metadata {
		fields = append(fields, key, value)
	}

	em.logger.LogError(ctx, "Error Recorded", err, fields...)

	// Check if alert threshold is exceeded
	if len(em.errorRates[errorCode]) >= em.alertThreshold {
		em.triggerAlert(ctx, errorCode, len(em.errorRates[errorCode]))
	}
}

// GetErrorMetrics returns current error metrics
func (em *ErrorMonitor) GetErrorMetrics() []ErrorMetrics {
	em.mutex.Lock() // 書き込みロックに変更（cleanOldEntriesでスライスを変更するため）
	defer em.mutex.Unlock()

	now := time.Now()
	metrics := make([]ErrorMetrics, 0, len(em.errorCounts))

	for errorCode, count := range em.errorCounts {
		// Clean old entries before calculating rate
		em.cleanOldEntries(errorCode, now)

		rateCount := len(em.errorRates[errorCode])
		rate := float64(rateCount) / em.timeWindow.Minutes()

		var lastOccurred time.Time
		if rateCount > 0 {
			lastOccurred = em.errorRates[errorCode][rateCount-1]
		}

		metrics = append(metrics, ErrorMetrics{
			ErrorCode:    errorCode,
			Count:        count,
			Rate:         rate,
			LastOccurred: lastOccurred,
		})
	}

	return metrics
}

// ResetMetrics resets all error metrics
func (em *ErrorMonitor) ResetMetrics() {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.errorCounts = make(map[string]int)
	em.errorRates = make(map[string][]time.Time)

	em.logger.LogInfo(context.Background(), "Error metrics reset")
}

// cleanOldEntries removes entries outside the time window
func (em *ErrorMonitor) cleanOldEntries(errorCode string, now time.Time) {
	cutoff := now.Add(-em.timeWindow)
	rates := em.errorRates[errorCode]

	// Find first entry within time window
	start := 0
	for i, t := range rates {
		if t.After(cutoff) {
			start = i
			break
		}
		if i == len(rates)-1 {
			start = len(rates) // All entries are old
		}
	}

	// Keep only recent entries
	em.errorRates[errorCode] = rates[start:]
}

// triggerAlert triggers an alert when error threshold is exceeded
func (em *ErrorMonitor) triggerAlert(ctx context.Context, errorCode string, currentRate int) {
	em.logger.Logger.Error("Error Rate Alert Triggered",
		"error_code", errorCode,
		"current_rate", currentRate,
		"threshold", em.alertThreshold,
		"time_window_minutes", em.timeWindow.Minutes(),
		"alert_level", "HIGH",
	)

	// Here you could integrate with external alerting systems
	// like Slack, PagerDuty, email notifications, etc.
}

// GetTopErrors returns the most frequent errors (sorted by count)
func (em *ErrorMonitor) GetTopErrors(limit int) []ErrorMetrics {
	return em.GetTopErrorsWithSort(limit, SortByCount)
}

// GetTopErrorsWithSort returns errors sorted by the specified criteria
func (em *ErrorMonitor) GetTopErrorsWithSort(limit int, sortBy SortBy) []ErrorMetrics {
	metrics := em.GetErrorMetrics()

	// Sort based on the specified criteria
	switch sortBy {
	case SortByRate:
		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].Rate > metrics[j].Rate
		})
	case SortByTime:
		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].LastOccurred.After(metrics[j].LastOccurred)
		})
	case SortByErrorCode:
		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].ErrorCode < metrics[j].ErrorCode
		})
	default:
		// Default to count-based sorting (includes SortByCount case)
		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].Count > metrics[j].Count
		})
	}

	if limit > 0 && limit < len(metrics) {
		return metrics[:limit]
	}

	return metrics
}

// GetErrorStatistics returns comprehensive error statistics
func (em *ErrorMonitor) GetErrorStatistics() ErrorStatistics {
	metrics := em.GetErrorMetrics()

	totalErrors := 0
	totalRate := 0.0
	var mostRecentError time.Time

	for _, metric := range metrics {
		totalErrors += metric.Count
		totalRate += metric.Rate
		if metric.LastOccurred.After(mostRecentError) {
			mostRecentError = metric.LastOccurred
		}
	}

	// 安全な平均レート計算（ゼロ除算防止）
	averageRate := 0.0
	metricsCount := len(metrics)
	if metricsCount > 0 {
		averageRate = totalRate / float64(metricsCount)
	}

	statistics := ErrorStatistics{
		TotalErrors:     totalErrors,
		ErrorTypeCount:  metricsCount,
		TotalRate:       totalRate,
		AverageRate:     averageRate,
		MostRecentError: mostRecentError,
	}

	// 安全性のためサニタイズを実行
	return statistics.Sanitize()
}

// GetTotalErrorCount returns the total number of errors across all types
func (em *ErrorMonitor) GetTotalErrorCount() int {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	total := 0
	for _, count := range em.errorCounts {
		total += count
	}
	return total
}

// GetErrorTypeCount returns the number of different error types
func (em *ErrorMonitor) GetErrorTypeCount() int {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	return len(em.errorCounts)
}

// IsValid validates the ErrorStatistics for consistency and safety
func (es ErrorStatistics) IsValid() bool {
	// 基本的な妥当性チェック
	if es.TotalErrors < 0 || es.ErrorTypeCount < 0 || es.TotalRate < 0.0 {
		return false
	}

	// NaN や Inf のチェック
	if math.IsNaN(es.AverageRate) || math.IsInf(es.AverageRate, 0) {
		return false
	}

	if math.IsNaN(es.TotalRate) || math.IsInf(es.TotalRate, 0) {
		return false
	}

	// 論理的整合性のチェック
	if es.ErrorTypeCount == 0 {
		// エラータイプが0の場合、他の値も全て0またはゼロ値であるべき
		return es.TotalErrors == 0 &&
			es.TotalRate == 0.0 &&
			es.AverageRate == 0.0 &&
			es.MostRecentError.IsZero()
	}

	// 追加の論理的整合性チェック
	return es.isLogicallyConsistent()
}

// isLogicallyConsistent performs additional logical consistency checks
func (es ErrorStatistics) isLogicallyConsistent() bool {
	// エラータイプが存在する場合の整合性チェック
	if es.ErrorTypeCount > 0 {
		// エラータイプが存在するのに総エラー数が0は不整合
		if es.TotalErrors == 0 {
			return false
		}

		// 平均レートの妥当性チェック
		expectedAverageRate := es.TotalRate / float64(es.ErrorTypeCount)
		if math.Abs(es.AverageRate-expectedAverageRate) > 0.001 { // 浮動小数点の誤差を考慮
			return false
		}
	}

	// 総エラー数が存在するのにエラータイプが0は不整合
	if es.TotalErrors > 0 && es.ErrorTypeCount == 0 {
		return false
	}

	// レートが存在するのにエラータイプが0は不整合
	if es.TotalRate > 0.0 && es.ErrorTypeCount == 0 {
		return false
	}

	// 最新エラー時刻が設定されているのにエラータイプが0は不整合
	if !es.MostRecentError.IsZero() && es.ErrorTypeCount == 0 {
		return false
	}

	return true
}

// Sanitize ensures all values are safe for JSON serialization
func (es ErrorStatistics) Sanitize() ErrorStatistics {
	sanitized := es

	// NaN や Inf を 0.0 に置換
	if math.IsNaN(sanitized.AverageRate) || math.IsInf(sanitized.AverageRate, 0) {
		sanitized.AverageRate = 0.0
	}

	if math.IsNaN(sanitized.TotalRate) || math.IsInf(sanitized.TotalRate, 0) {
		sanitized.TotalRate = 0.0
	}

	// 負の値を0に修正
	if sanitized.TotalErrors < 0 {
		sanitized.TotalErrors = 0
	}

	if sanitized.ErrorTypeCount < 0 {
		sanitized.ErrorTypeCount = 0
	}

	if sanitized.TotalRate < 0.0 {
		sanitized.TotalRate = 0.0
	}

	// 論理的整合性の修正
	if sanitized.ErrorTypeCount == 0 {
		// エラータイプが0の場合、他の値も0またはゼロ値に修正
		sanitized.TotalErrors = 0
		sanitized.TotalRate = 0.0
		sanitized.AverageRate = 0.0
		sanitized.MostRecentError = time.Time{} // ゼロ値に設定
	}

	return sanitized
}
