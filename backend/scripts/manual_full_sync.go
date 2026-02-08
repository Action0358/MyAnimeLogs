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
	fmt.Println("This will re-sync all animes from Annict, episodes from Jikan, and descriptions from TMDB.")
	fmt.Println("Estimated time: 90-120 minutes")
	fmt.Println()

	startTime := time.Now()

	// Step 1: Re-sync all animes from Annict (updates media, official_site_url, etc.)
	logger.Info("Step 1/5: Re-syncing all animes from Annict API...")
	annictSuccess, annictFailure, err := annictClient.SyncAllExistingAnimes()
	if err != nil {
		logger.Warn("Failed to re-sync animes from Annict", "error", err)
	} else {
		logger.Info("Annict re-sync completed",
			"success", annictSuccess,
			"failure", annictFailure)
	}

	// Step 2: Sync episodes from Jikan API
	logger.Info("Step 2/5: Syncing episodes from Jikan API...")
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

	// Step 3: Sync episode details (air_date and duration)
	logger.Info("Step 3/5: Syncing episode details from Jikan API...")
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

	// Step 4: Sync descriptions from TMDB
	logger.Info("Step 4/5: Syncing descriptions from TMDB...")
	tmdbAPIKey := os.Getenv("TMDB_API_KEY")
	if tmdbAPIKey == "" {
		logger.Warn("TMDB_API_KEY not set, skipping description sync")
	} else {
		descSuccess, descNotFound, descFailure, err := animeService.SyncDescriptionsFromTMDB()
		if err != nil {
			logger.Warn("Failed to sync descriptions from TMDB", "error", err)
		} else {
			logger.Info("TMDB description sync completed",
				"success", descSuccess,
				"not_found", descNotFound,
				"failure", descFailure)
		}
	}

	// Step 5: Sync released_on from Jikan API
	logger.Info("Step 5/5: Syncing release dates from Jikan API...")
	releasedSuccess, releasedNotFound, releasedFailure, err := animeService.SyncReleasedOnFromJikan()
	if err != nil {
		logger.Warn("Failed to sync released_on from Jikan", "error", err)
	} else {
		logger.Info("Jikan released_on sync completed",
			"success", releasedSuccess,
			"not_found", releasedNotFound,
			"failure", releasedFailure)
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

	// Display anime description statistics
	var totalAnimes, withDescription, withReleasedOn, withMedia, withOfficialSite int
	err = db.QueryRow(`
		SELECT 
			COUNT(*),
			COUNT(CASE WHEN description IS NOT NULL AND description != '' THEN 1 END),
			COUNT(CASE WHEN released_on IS NOT NULL AND released_on != '' THEN 1 END),
			COUNT(CASE WHEN media IS NOT NULL AND media != '' THEN 1 END),
			COUNT(CASE WHEN official_site_url IS NOT NULL AND official_site_url != '' THEN 1 END)
		FROM animes
	`).Scan(&totalAnimes, &withDescription, &withReleasedOn, &withMedia, &withOfficialSite)
	if err == nil && totalAnimes > 0 {
		logger.Info("Anime statistics",
			"total_animes", totalAnimes,
			"with_description", withDescription,
			"description_percent", fmt.Sprintf("%.1f%%", float64(withDescription)/float64(totalAnimes)*100),
			"with_released_on", withReleasedOn,
			"released_on_percent", fmt.Sprintf("%.1f%%", float64(withReleasedOn)/float64(totalAnimes)*100),
			"with_media", withMedia,
			"media_percent", fmt.Sprintf("%.1f%%", float64(withMedia)/float64(totalAnimes)*100),
			"with_official_site", withOfficialSite,
			"official_site_percent", fmt.Sprintf("%.1f%%", float64(withOfficialSite)/float64(totalAnimes)*100))
	}

	logger.Info("✓ Manual full sync completed successfully!")
	logger.Info("Future syncs will only update new/changed data.")

	return nil
}
