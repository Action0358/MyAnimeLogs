package common

import (
	"context"
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

// GetTopErrors returns the most frequent errors
func (em *ErrorMonitor) GetTopErrors(limit int) []ErrorMetrics {
	metrics := em.GetErrorMetrics()

	// Sort by count in descending order using standard library
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Count > metrics[j].Count
	})

	if limit > 0 && limit < len(metrics) {
		return metrics[:limit]
	}

	return metrics
}

// GetErrorStatistics returns comprehensive error statistics
func (em *ErrorMonitor) GetErrorStatistics() map[string]interface{} {
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

	return map[string]interface{}{
		"total_errors":      totalErrors,
		"error_type_count":  len(metrics),
		"total_rate":        totalRate,
		"average_rate":      totalRate / float64(len(metrics)),
		"most_recent_error": mostRecentError,
	}
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
