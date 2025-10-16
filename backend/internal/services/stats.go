package services

import (
	"fmt"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
)

// StatsService handles user statistics
type StatsService struct {
	db     *database.DB
	logger *common.Logger
}

// NewStatsService creates a new stats service
func NewStatsService(db *database.DB, logger *common.Logger) *StatsService {
	return &StatsService{
		db:     db,
		logger: logger,
	}
}

// UserStats represents user statistics
type UserStats struct {
	TotalWatchedEpisodes  int `json:"total_watched_episodes"`
	TotalWatchTimeMinutes int `json:"total_watch_time_minutes"`
	CompletedAnimes       int `json:"completed_animes"`
	MonthlyWatchTime      int `json:"monthly_watch_time"`
}

// GetUserStats retrieves comprehensive user statistics
func (s *StatsService) GetUserStats(userID int) (*UserStats, error) {
	stats := &UserStats{}

	// Get total watched episodes
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM watch_statuses
		WHERE user_id = $1 AND viewed = true
	`, userID).Scan(&stats.TotalWatchedEpisodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get total watched episodes: %w", err)
	}

	// Get total watch time (sum of episode durations)
	err = s.db.QueryRow(`
		SELECT COALESCE(SUM(e.duration_minutes), 0)
		FROM watch_statuses ws
		JOIN episodes e ON ws.episode_id = e.id
		WHERE ws.user_id = $1 AND ws.viewed = true AND e.duration_minutes IS NOT NULL
	`, userID).Scan(&stats.TotalWatchTimeMinutes)
	if err != nil {
		return nil, fmt.Errorf("failed to get total watch time: %w", err)
	}

	// Get completed animes count
	err = s.db.QueryRow(`
		SELECT COUNT(*)
		FROM user_favorites
		WHERE user_id = $1 AND status = 'completed'
	`, userID).Scan(&stats.CompletedAnimes)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed animes: %w", err)
	}

	// Get monthly watch time (current month)
	err = s.db.QueryRow(`
		SELECT COALESCE(SUM(e.duration_minutes), 0)
		FROM watch_statuses ws
		JOIN episodes e ON ws.episode_id = e.id
		WHERE ws.user_id = $1 
		  AND ws.viewed = true 
		  AND e.duration_minutes IS NOT NULL
		  AND ws.updated_at >= DATE_TRUNC('month', CURRENT_DATE)
	`, userID).Scan(&stats.MonthlyWatchTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly watch time: %w", err)
	}

	return stats, nil
}

// GetTotalWatchTime calculates total watch time for a user
func (s *StatsService) GetTotalWatchTime(userID int) (int, error) {
	var totalMinutes int
	err := s.db.QueryRow(`
		SELECT COALESCE(SUM(e.duration_minutes), 0) as total_minutes
		FROM watch_statuses ws
		JOIN episodes e ON ws.episode_id = e.id
		WHERE ws.user_id = $1 AND ws.viewed = true AND e.duration_minutes IS NOT NULL
	`, userID).Scan(&totalMinutes)
	if err != nil {
		return 0, fmt.Errorf("failed to get total watch time: %w", err)
	}

	return totalMinutes, nil
}

// GetMonthlyWatchTime calculates watch time for a specific month
func (s *StatsService) GetMonthlyWatchTime(userID int, year int, month int) (int, error) {
	var totalMinutes int
	err := s.db.QueryRow(`
		SELECT COALESCE(SUM(e.duration_minutes), 0) as total_minutes
		FROM watch_statuses ws
		JOIN episodes e ON ws.episode_id = e.id
		WHERE ws.user_id = $1 
		  AND ws.viewed = true 
		  AND e.duration_minutes IS NOT NULL
		  AND EXTRACT(YEAR FROM ws.updated_at) = $2
		  AND EXTRACT(MONTH FROM ws.updated_at) = $3
	`, userID, year, month).Scan(&totalMinutes)
	if err != nil {
		return 0, fmt.Errorf("failed to get monthly watch time: %w", err)
	}

	return totalMinutes, nil
}
