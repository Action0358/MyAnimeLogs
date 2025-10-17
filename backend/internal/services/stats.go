package services

import (
	"fmt"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
	"myanimelogs/internal/models"
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

// GetUserStats retrieves comprehensive user statistics
func (s *StatsService) GetUserStats(userID int) (*models.UserStats, error) {
	stats := &models.UserStats{}

	// Get total episodes watched
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM watch_statuses 
		WHERE user_id = $1 AND viewed = true
	`, userID).Scan(&stats.TotalEpisodesWatched)
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes watched: %w", err)
	}

	// Get total animes completed
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM user_favorites 
		WHERE user_id = $1 AND status = 'completed'
	`, userID).Scan(&stats.TotalAnimesCompleted)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed animes: %w", err)
	}

	// Get actual total watch time from episode durations
	var totalMinutes int
	err = s.db.QueryRow(`
		SELECT COALESCE(SUM(e.duration_minutes), 0)
		FROM watch_statuses ws
		JOIN episodes e ON ws.episode_id = e.id
		WHERE ws.user_id = $1 AND ws.viewed = true AND e.duration_minutes IS NOT NULL
	`, userID).Scan(&totalMinutes)
	if err != nil {
		// Fallback to estimate if query fails
		stats.TotalWatchTimeHours = float64(stats.TotalEpisodesWatched) * 0.4
	} else if totalMinutes > 0 {
		// Convert minutes to hours
		stats.TotalWatchTimeHours = float64(totalMinutes) / 60.0
	} else {
		// Fallback to estimate if no duration data available
		stats.TotalWatchTimeHours = float64(stats.TotalEpisodesWatched) * 0.4
	}

	// Get favorite genres (top 3)
	rows, err := s.db.Query(`
		SELECT ag.genre, COUNT(*) as count
		FROM anime_genres ag
		JOIN user_favorites uf ON ag.anime_id = uf.anime_id
		WHERE uf.user_id = $1
		GROUP BY ag.genre
		ORDER BY count DESC
		LIMIT 3
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get favorite genres: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var genre string
		var count int
		if err := rows.Scan(&genre, &count); err != nil {
			continue
		}
		stats.FavoriteGenres = append(stats.FavoriteGenres, genre)
	}

	// Get this month's episodes
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM watch_history 
		WHERE user_id = $1 AND watched_at >= date_trunc('month', CURRENT_DATE)
	`, userID).Scan(&stats.ThisMonthEpisodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get this month episodes: %w", err)
	}

	return stats, nil
}

// GetTotalWatchTime calculates total watch time for a user in minutes
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

// GetMonthlyWatchTime calculates watch time for a specific month in minutes
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
