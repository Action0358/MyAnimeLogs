package common

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"
)

func TestErrorStatistics_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		stats    ErrorStatistics
		expected bool
	}{
		{
			name: "Valid statistics",
			stats: ErrorStatistics{
				TotalErrors:     100,
				ErrorTypeCount:  5,
				TotalRate:       10.5,
				AverageRate:     2.1,
				MostRecentError: time.Now(),
			},
			expected: true,
		},
		{
			name: "Zero values (valid)",
			stats: ErrorStatistics{
				TotalErrors:     0,
				ErrorTypeCount:  0,
				TotalRate:       0.0,
				AverageRate:     0.0,
				MostRecentError: time.Time{},
			},
			expected: true,
		},
		{
			name: "Negative total errors (invalid)",
			stats: ErrorStatistics{
				TotalErrors:     -1,
				ErrorTypeCount:  5,
				TotalRate:       10.5,
				AverageRate:     2.1,
				MostRecentError: time.Now(),
			},
			expected: false,
		},
		{
			name: "NaN average rate (invalid)",
			stats: ErrorStatistics{
				TotalErrors:     100,
				ErrorTypeCount:  5,
				TotalRate:       10.5,
				AverageRate:     math.NaN(),
				MostRecentError: time.Now(),
			},
			expected: false,
		},
		{
			name: "Infinite average rate (invalid)",
			stats: ErrorStatistics{
				TotalErrors:     100,
				ErrorTypeCount:  5,
				TotalRate:       10.5,
				AverageRate:     math.Inf(1),
				MostRecentError: time.Now(),
			},
			expected: false,
		},
		{
			name: "Inconsistent zero error types with non-zero values (invalid)",
			stats: ErrorStatistics{
				TotalErrors:     100, // Should be 0 if ErrorTypeCount is 0
				ErrorTypeCount:  0,
				TotalRate:       10.5,       // Should be 0 if ErrorTypeCount is 0
				AverageRate:     2.1,        // Should be 0 if ErrorTypeCount is 0
				MostRecentError: time.Now(), // Should be zero time if ErrorTypeCount is 0
			},
			expected: false,
		},
		{
			name: "Zero error types with non-zero MostRecentError (invalid)",
			stats: ErrorStatistics{
				TotalErrors:     0,
				ErrorTypeCount:  0,
				TotalRate:       0.0,
				AverageRate:     0.0,
				MostRecentError: time.Now(), // ❌ Should be zero time
			},
			expected: false,
		},
		{
			name: "Non-zero errors with zero error types (invalid)",
			stats: ErrorStatistics{
				TotalErrors:     50, // ❌ Inconsistent with ErrorTypeCount = 0
				ErrorTypeCount:  0,
				TotalRate:       0.0,
				AverageRate:     0.0,
				MostRecentError: time.Time{},
			},
			expected: false,
		},
		{
			name: "Inconsistent average rate calculation (invalid)",
			stats: ErrorStatistics{
				TotalErrors:     100,
				ErrorTypeCount:  5,
				TotalRate:       10.0,
				AverageRate:     5.0, // ❌ Should be 2.0 (10.0/5)
				MostRecentError: time.Now(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.stats.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestErrorStatistics_Sanitize(t *testing.T) {
	tests := []struct {
		name     string
		input    ErrorStatistics
		expected ErrorStatistics
	}{
		{
			name: "Clean data (no changes)",
			input: ErrorStatistics{
				TotalErrors:     100,
				ErrorTypeCount:  5,
				TotalRate:       10.5,
				AverageRate:     2.1,
				MostRecentError: time.Now(),
			},
			expected: ErrorStatistics{
				TotalErrors:     100,
				ErrorTypeCount:  5,
				TotalRate:       10.5,
				AverageRate:     2.1,
				MostRecentError: time.Now(),
			},
		},
		{
			name: "NaN values sanitized",
			input: ErrorStatistics{
				TotalErrors:     100,
				ErrorTypeCount:  5,
				TotalRate:       math.NaN(),
				AverageRate:     math.NaN(),
				MostRecentError: time.Now(),
			},
			expected: ErrorStatistics{
				TotalErrors:     100,
				ErrorTypeCount:  5,
				TotalRate:       0.0,
				AverageRate:     0.0,
				MostRecentError: time.Now(),
			},
		},
		{
			name: "Negative values sanitized",
			input: ErrorStatistics{
				TotalErrors:     -10,
				ErrorTypeCount:  -5, // Will be sanitized to 0
				TotalRate:       -10.5,
				AverageRate:     2.1,
				MostRecentError: time.Now(),
			},
			expected: ErrorStatistics{
				TotalErrors:     0,           // Sanitized due to ErrorTypeCount = 0
				ErrorTypeCount:  0,           // Sanitized from negative
				TotalRate:       0.0,         // Sanitized due to ErrorTypeCount = 0
				AverageRate:     0.0,         // Sanitized due to ErrorTypeCount = 0
				MostRecentError: time.Time{}, // Sanitized due to ErrorTypeCount = 0
			},
		},
		{
			name: "Infinite values sanitized",
			input: ErrorStatistics{
				TotalErrors:     100,
				ErrorTypeCount:  5,
				TotalRate:       math.Inf(1),
				AverageRate:     math.Inf(-1),
				MostRecentError: time.Now(),
			},
			expected: ErrorStatistics{
				TotalErrors:     100,
				ErrorTypeCount:  5,
				TotalRate:       0.0,
				AverageRate:     0.0,
				MostRecentError: time.Now(),
			},
		},
		{
			name: "Inconsistent zero error types sanitized",
			input: ErrorStatistics{
				TotalErrors:     100, // Will be sanitized to 0
				ErrorTypeCount:  0,
				TotalRate:       10.5,       // Will be sanitized to 0.0
				AverageRate:     2.1,        // Will be sanitized to 0.0
				MostRecentError: time.Now(), // Will be sanitized to zero time
			},
			expected: ErrorStatistics{
				TotalErrors:     0,
				ErrorTypeCount:  0,
				TotalRate:       0.0,
				AverageRate:     0.0,
				MostRecentError: time.Time{}, // Zero time
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Sanitize()

			if result.TotalErrors != tt.expected.TotalErrors {
				t.Errorf("TotalErrors = %v, expected %v", result.TotalErrors, tt.expected.TotalErrors)
			}
			if result.ErrorTypeCount != tt.expected.ErrorTypeCount {
				t.Errorf("ErrorTypeCount = %v, expected %v", result.ErrorTypeCount, tt.expected.ErrorTypeCount)
			}
			if result.TotalRate != tt.expected.TotalRate {
				t.Errorf("TotalRate = %v, expected %v", result.TotalRate, tt.expected.TotalRate)
			}
			if result.AverageRate != tt.expected.AverageRate {
				t.Errorf("AverageRate = %v, expected %v", result.AverageRate, tt.expected.AverageRate)
			}
		})
	}
}

func TestGetErrorStatistics_ZeroDivisionPrevention(t *testing.T) {
	logger := NewLogger(LogLevelDebug, "text")
	monitor := NewErrorMonitor(logger, 10, 5*time.Minute)

	// エラーが記録されていない状態でテスト
	statistics := monitor.GetErrorStatistics()

	// ゼロ除算が発生していないことを確認
	if math.IsNaN(statistics.AverageRate) {
		t.Error("AverageRate should not be NaN")
	}

	if math.IsInf(statistics.AverageRate, 0) {
		t.Error("AverageRate should not be Inf")
	}

	// 期待される値の確認
	if statistics.TotalErrors != 0 {
		t.Errorf("TotalErrors = %v, expected 0", statistics.TotalErrors)
	}

	if statistics.ErrorTypeCount != 0 {
		t.Errorf("ErrorTypeCount = %v, expected 0", statistics.ErrorTypeCount)
	}

	if statistics.AverageRate != 0.0 {
		t.Errorf("AverageRate = %v, expected 0.0", statistics.AverageRate)
	}

	if !statistics.IsValid() {
		t.Error("Statistics should be valid")
	}
}
func TestGetTopErrorsWithSort(t *testing.T) {
	logger := NewLogger(LogLevelDebug, "text")
	monitor := NewErrorMonitor(logger, 10, 5*time.Minute)

	// テスト用のエラーを記録
	ctx := context.Background()

	// 異なる特性を持つエラーを記録
	monitor.RecordError(ctx, "ERROR_A", errors.New("error a"), map[string]interface{}{})
	monitor.RecordError(ctx, "ERROR_A", errors.New("error a"), map[string]interface{}{})
	monitor.RecordError(ctx, "ERROR_A", errors.New("error a"), map[string]interface{}{}) // Count: 3

	time.Sleep(10 * time.Millisecond) // 時間差を作る

	monitor.RecordError(ctx, "ERROR_B", errors.New("error b"), map[string]interface{}{})
	monitor.RecordError(ctx, "ERROR_B", errors.New("error b"), map[string]interface{}{}) // Count: 2, より新しい

	monitor.RecordError(ctx, "ERROR_C", errors.New("error c"), map[string]interface{}{}) // Count: 1, 最新

	tests := []struct {
		name     string
		sortBy   SortBy
		expected string // 最初のエラーコード
	}{
		{
			name:     "Sort by count (descending)",
			sortBy:   SortByCount,
			expected: "ERROR_A", // 最多の3回
		},
		{
			name:     "Sort by time (most recent first)",
			sortBy:   SortByTime,
			expected: "ERROR_C", // 最新
		},
		{
			name:     "Sort by error code (ascending)",
			sortBy:   SortByErrorCode,
			expected: "ERROR_A", // アルファベット順で最初
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := monitor.GetTopErrorsWithSort(10, tt.sortBy)

			if len(result) == 0 {
				t.Fatal("Expected at least one error metric")
			}

			if result[0].ErrorCode != tt.expected {
				t.Errorf("Expected first error to be %s, got %s", tt.expected, result[0].ErrorCode)
			}
		})
	}
}

func TestGetTopErrorsWithSort_Limit(t *testing.T) {
	logger := NewLogger(LogLevelDebug, "text")
	monitor := NewErrorMonitor(logger, 10, 5*time.Minute)

	// 複数のエラーを記録
	ctx := context.Background()
	errorCodes := []string{"ERROR_1", "ERROR_2", "ERROR_3", "ERROR_4", "ERROR_5"}

	for _, code := range errorCodes {
		monitor.RecordError(ctx, code, errors.New("test error"), map[string]interface{}{})
	}

	// 制限をテスト
	result := monitor.GetTopErrorsWithSort(3, SortByCount)

	if len(result) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result))
	}
}

func TestGetTopErrorsWithSort_InvalidSort(t *testing.T) {
	logger := NewLogger(LogLevelDebug, "text")
	monitor := NewErrorMonitor(logger, 10, 5*time.Minute)

	// エラーを記録
	ctx := context.Background()
	monitor.RecordError(ctx, "ERROR_A", errors.New("error a"), map[string]interface{}{})
	monitor.RecordError(ctx, "ERROR_B", errors.New("error b"), map[string]interface{}{})

	// 無効なソート条件でデフォルト（カウント順）になることを確認
	result := monitor.GetTopErrorsWithSort(10, SortBy("invalid"))

	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}

	// デフォルトでカウント順になっているかは、この場合両方とも1回なので順序は不定
	// ただし、結果が返されることを確認
}
func TestGetTotalErrorCount(t *testing.T) {
	logger := NewLogger(LogLevelDebug, "text")
	monitor := NewErrorMonitor(logger, 10, 5*time.Minute)

	// 初期状態では0
	if count := monitor.GetTotalErrorCount(); count != 0 {
		t.Errorf("Expected 0 total errors initially, got %d", count)
	}

	// エラーを記録
	ctx := context.Background()
	monitor.RecordError(ctx, "ERROR_A", errors.New("error a"), map[string]interface{}{})
	monitor.RecordError(ctx, "ERROR_A", errors.New("error a"), map[string]interface{}{})
	monitor.RecordError(ctx, "ERROR_B", errors.New("error b"), map[string]interface{}{})

	// 総数が正しいことを確認
	expectedTotal := 3
	if count := monitor.GetTotalErrorCount(); count != expectedTotal {
		t.Errorf("Expected %d total errors, got %d", expectedTotal, count)
	}
}

func TestGetTopErrors_ResponseFields(t *testing.T) {
	logger := NewLogger(LogLevelDebug, "text")
	monitor := NewErrorMonitor(logger, 10, 5*time.Minute)

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

	// 上位2件を取得
	topErrors := monitor.GetTopErrorsWithSort(2, SortByCount)

	// 上位2件の合計を計算
	countInTopList := 0
	for _, errorMetric := range topErrors {
		countInTopList += errorMetric.Count
	}

	// システム全体の総数を取得
	totalErrors := monitor.GetTotalErrorCount()

	// 検証
	if len(topErrors) != 2 {
		t.Errorf("Expected 2 top errors, got %d", len(topErrors))
	}

	if countInTopList != 8 { // ERROR_A(5) + ERROR_B(3) = 8
		t.Errorf("Expected count_in_top_list to be 8, got %d", countInTopList)
	}

	if totalErrors != 10 { // ERROR_A(5) + ERROR_B(3) + ERROR_C(2) = 10
		t.Errorf("Expected total_errors to be 10, got %d", totalErrors)
	}

	// 上位リストの合計 < システム全体の総数 であることを確認
	if countInTopList >= totalErrors {
		t.Errorf("count_in_top_list (%d) should be less than total_errors (%d)", countInTopList, totalErrors)
	}
}
