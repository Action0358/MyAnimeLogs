//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
	"myanimelogs/internal/services"

	"github.com/joho/godotenv"
)

func main() {
	logger := common.NewLogger(common.LogLevelInfo, "text")

	if err := run(logger); err != nil {
		logger.Error("Manual full sync failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *common.Logger) error {
	// Load environment variables
	if err := godotenv.Load(".env.local"); err != nil {
		logger.Warn(".env.local file not found", "error", err)
	}

	// Initialize database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	db, err := database.New(dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize services
	annictToken := os.Getenv("ANNICT_API_KEY")
	if annictToken == "" {
		return fmt.Errorf("ANNICT_API_KEY is not set")
	}

	annictClient := services.NewAnnictClientWithDB(annictToken, db, logger)
	animeService := services.NewAnimeService(db, logger, annictClient)

	fmt.Println("\n=== Manual Full Sync ===")
	fmt.Println("This will sync all episodes from Jikan API.")
	fmt.Println("Estimated time: 40-60 minutes")
	fmt.Println()

	startTime := time.Now()

	// Step 1: Sync episodes from Jikan API
	logger.Info("Step 1/2: Syncing episodes from Jikan API...")
	episodeSuccess, episodeFailure, episodeErrors, err := animeService.SyncEpisodesFromJikan(nil)
	if err != nil {
		return fmt.Errorf("failed to sync episodes from Jikan: %w", err)
	}

	logger.Info("Episode sync completed",
		"success", episodeSuccess,
		"failure", episodeFailure)
	if len(episodeErrors) > 0 {
		sampleSize := 5
		if len(episodeErrors) < sampleSize {
			sampleSize = len(episodeErrors)
		}
		logger.Warn("Episode sync errors occurred",
			"error_count", len(episodeErrors),
			"sample_errors", episodeErrors[:sampleSize])
	}

	// Step 2: Sync episode details (air_date and duration)
	logger.Info("Step 2/2: Syncing episode details from Jikan API...")
	detailSuccess, detailFailure, detailErrors, err := animeService.SyncEpisodeDetailsForAllAnimes()
	if err != nil {
		return fmt.Errorf("failed to sync episode details: %w", err)
	}

	logger.Info("Episode details completed",
		"success", detailSuccess,
		"failure", detailFailure)
	if len(detailErrors) > 0 {
		sampleSize := 5
		if len(detailErrors) < sampleSize {
			sampleSize = len(detailErrors)
		}
		logger.Warn("Episode detail errors occurred",
			"error_count", len(detailErrors),
			"sample_errors", detailErrors[:sampleSize])
	}

	duration := time.Since(startTime)
	logger.Info("Sync completed", "duration", duration.String())

	// Display statistics
	logger.Info("Fetching final statistics...")
	rows, err := db.Query(`
		SELECT 
			COUNT(*) as total_episodes,
			COUNT(CASE WHEN title != '第' || episode_number || '話' AND title IS NOT NULL AND title != '' THEN 1 END) as with_custom_title,
			COUNT(duration_minutes) as with_duration,
			COUNT(air_date) as with_air_date
		FROM episodes
	`)
	if err != nil {
		return fmt.Errorf("failed to get statistics: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var total, customTitle, duration, airDate int
		if err := rows.Scan(&total, &customTitle, &duration, &airDate); err != nil {
			return fmt.Errorf("failed to scan statistics: %w", err)
		}

		logger.Info("Final statistics",
			"total_episodes", total,
			"with_custom_title", customTitle,
			"custom_title_percent", fmt.Sprintf("%.1f%%", float64(customTitle)/float64(total)*100),
			"with_duration", duration,
			"duration_percent", fmt.Sprintf("%.1f%%", float64(duration)/float64(total)*100),
			"with_air_date", airDate,
			"air_date_percent", fmt.Sprintf("%.1f%%", float64(airDate)/float64(total)*100))
	}

	logger.Info("✓ Manual full sync completed successfully!")
	logger.Info("Future syncs will only update new/changed data.")

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
