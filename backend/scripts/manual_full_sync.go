//go:build ignore
// +build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
	"myanimelogs/internal/services"

	"github.com/joho/godotenv"
)

func main() {
	// Parse command line arguments
	env := flag.String("env", "local", "Environment to use: local, staging, production")
	priority := flag.Bool("priority", false, "Priority mode: sync current season + popular animes only")
	flag.Parse()

	logger := common.NewLogger(common.LogLevelInfo, "text")

	if *priority {
		if err := runPriority(logger, *env); err != nil {
			logger.Error("Priority sync failed", "error", err)
			os.Exit(1)
		}
	} else {
		if err := run(logger, *env); err != nil {
			logger.Error("Manual full sync failed", "error", err)
			os.Exit(1)
		}
	}
}

func initServices(logger *common.Logger, env string) (*database.DB, *services.AnnictClient, *services.AnimeService, error) {
	var envFile string
	switch env {
	case "local":
		envFile = ".env.local"
	case "staging":
		envFile = ".env.staging"
	case "production":
		envFile = ".env.production"
	default:
		return nil, nil, nil, fmt.Errorf("invalid environment: %s (use: local, staging, production)", env)
	}

	logger.Info("Loading environment", "env", env, "file", envFile)
	if err := godotenv.Load(envFile); err != nil {
		logger.Warn("Environment file not found", "file", envFile, "error", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, nil, nil, fmt.Errorf("DATABASE_URL is not set")
	}

	db, err := database.New(dbURL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	annictToken := os.Getenv("ANNICT_API_KEY")
	if annictToken == "" {
		db.Close()
		return nil, nil, nil, fmt.Errorf("ANNICT_API_KEY is not set")
	}

	annictClient := services.NewAnnictClientWithDB(annictToken, db, logger)
	animeService := services.NewAnimeService(db, logger, annictClient)

	return db, annictClient, animeService, nil
}

func runPriority(logger *common.Logger, env string) error {
	db, annictClient, animeService, err := initServices(logger, env)
	if err != nil {
		return err
	}
	defer db.Close()

	// Determine current season
	now := time.Now()
	year := now.Year()
	var seasonName string
	switch {
	case now.Month() >= 1 && now.Month() <= 3:
		seasonName = "winter"
	case now.Month() >= 4 && now.Month() <= 6:
		seasonName = "spring"
	case now.Month() >= 7 && now.Month() <= 9:
		seasonName = "summer"
	default:
		seasonName = "autumn"
	}
	currentSeason := fmt.Sprintf("%d-%s", year, seasonName)

	fmt.Printf("\n=== Priority Sync (%s) ===\n", env)
	fmt.Printf("Current season: %s\n", currentSeason)
	fmt.Println("Syncing current season + popular animes (top 100)")
	fmt.Println()

	startTime := time.Now()
	animeIDSet := make(map[int]bool)

	// Step 1: Fetch current season from Annict
	logger.Info("Step 1/3: Fetching current season animes from Annict...", "season", currentSeason)
	seasonAnimes, err := annictClient.FetchAndSyncWorksBySeason(currentSeason)
	if err != nil {
		logger.Warn("Failed to fetch current season", "error", err)
	} else {
		for _, a := range seasonAnimes {
			animeIDSet[a.ID] = true
		}
		logger.Info("Current season animes synced", "count", len(seasonAnimes))
	}

	// Step 2: Re-sync popular animes (top 100 by watchers_count)
	logger.Info("Step 2/3: Re-syncing popular animes from Annict...")
	popularAnimes, err := annictClient.FetchAndSyncPopularWorks()
	if err != nil {
		logger.Warn("Failed to fetch popular works", "error", err)
	} else {
		for _, a := range popularAnimes {
			animeIDSet[a.ID] = true
		}
		logger.Info("Popular animes synced", "count", len(popularAnimes))
	}

	// Collect unique anime IDs
	animeIDs := make([]int, 0, len(animeIDSet))
	for id := range animeIDSet {
		animeIDs = append(animeIDs, id)
	}
	logger.Info("Total unique animes to enrich", "count", len(animeIDs))

	// Step 3: Enrich with episodes, descriptions, released_on
	logger.Info("Step 3/3: Enriching anime data (episodes, descriptions, released_on)...")
	if err := animeService.SyncPriorityAnimeData(animeIDs); err != nil {
		logger.Warn("Failed to enrich priority anime data", "error", err)
	}

	duration := time.Since(startTime)
	logger.Info("Priority sync completed", "duration", duration.String(), "total_animes", len(animeIDs))

	return nil
}

func run(logger *common.Logger, env string) error {
	db, annictClient, animeService, err := initServices(logger, env)
	if err != nil {
		return err
	}
	defer db.Close()

	fmt.Printf("\n=== Manual Full Sync (%s) ===\n", env)
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
			COUNT(CASE WHEN released_on IS NOT NULL THEN 1 END),
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
