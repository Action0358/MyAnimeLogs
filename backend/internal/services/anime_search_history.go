package services

import (
	"encoding/json"
	"fmt"

	"myanimelogs/internal/constants"
	"myanimelogs/internal/models"
)

// logSearch - 検索ログを記録（非同期）
func (s *AnimeService) logSearch(query string, filters map[string]interface{}, userID *int, resultCount int) {
	// 安全にJSONに変換
	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		// JSON変換エラーは無視（メイン機能に影響しない）
		s.logger.Warn("Failed to marshal search filters", "error", err, "filters", filters)
		return
	}

	_, err = s.db.Exec(`
		INSERT INTO search_logs (user_id, query, filters, result_count, searched_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	`, userID, query, string(filtersJSON), resultCount)
	if err != nil {
		// ログ記録エラーは無視（メイン機能に影響しない）
		s.logger.Warn("Failed to log search", "error", err, "query", query, "result_count", resultCount)
	}
}

// logAdvancedSearch - 高度検索のログを記録（非同期）
func (s *AnimeService) logAdvancedSearch(params models.AdvancedSearchParams, userID *int, resultCount int) {
	filters := map[string]interface{}{
		"genres":      params.Genres,
		"year":        params.Year,
		"season":      params.Season,
		"status":      params.Status,
		"user_status": params.UserStatus,
		"sort":        params.Sort,
		"order":       params.Order,
	}
	go s.logSearch(params.Query, filters, userID, resultCount)
}

// logBasicSearch - 基本検索のログを記録（非同期）
func (s *AnimeService) logBasicSearch(params models.AnimeSearchParams, userID *int, resultCount int) {
	filters := map[string]interface{}{
		"genre":  params.Genre,
		"season": params.Season,
		"status": params.Status,
		"sort":   params.Sort,
	}
	go s.logSearch(params.Keyword, filters, userID, resultCount)
}

// GetSearchHistory - ユーザーの検索履歴を取得
func (s *AnimeService) GetSearchHistory(userID int, limit int) ([]models.SearchLog, error) {
	if limit <= 0 {
		limit = constants.SearchHistoryDefaultLimit
	}
	if limit > constants.SearchHistoryMaxItems {
		limit = constants.SearchHistoryMaxItems
	}

	rows, err := s.db.Query(`
		SELECT id, user_id, query, COALESCE(filters, '{}') as filters, result_count, clicked_anime_id, searched_at
		FROM search_logs
		WHERE user_id = $1
		ORDER BY searched_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get search history: %w", err)
	}
	defer rows.Close()

	history := []models.SearchLog{}
	for rows.Next() {
		var log models.SearchLog
		err := rows.Scan(&log.ID, &log.UserID, &log.Query, &log.Filters, &log.ResultCount, &log.ClickedID, &log.SearchedAt)
		if err != nil {
			s.logger.Warn("Failed to scan search log row", "error", err)
			continue
		}
		history = append(history, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search history: %w", err)
	}

	return history, nil
}

// DeleteSearchHistory - ユーザーの検索履歴を削除
func (s *AnimeService) DeleteSearchHistory(userID int) error {
	_, err := s.db.Exec(`
		DELETE FROM search_logs
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete search history: %w", err)
	}

	return nil
}

// DeleteSearchHistoryItem - 特定の検索履歴を削除
func (s *AnimeService) DeleteSearchHistoryItem(userID int, logID int) error {
	result, err := s.db.Exec(`
		DELETE FROM search_logs
		WHERE id = $1 AND user_id = $2
	`, logID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete search history item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrSearchHistoryNotFound
	}

	return nil
}

// CleanupOldSearchHistory - 古い検索履歴を削除（最新N件を残す）
// ROW_NUMBER()ウィンドウ関数を使用して効率的に削除
func (s *AnimeService) CleanupOldSearchHistory(userID int) error {
	_, err := s.db.Exec(`
		DELETE FROM search_logs
		WHERE id IN (
			SELECT id FROM (
				SELECT id, ROW_NUMBER() OVER (ORDER BY searched_at DESC) as rn
				FROM search_logs
				WHERE user_id = $1
			) ranked
			WHERE rn > $2
		)
	`, userID, constants.SearchHistoryMaxItems)
	if err != nil {
		return fmt.Errorf("failed to cleanup old search history: %w", err)
	}

	return nil
}
