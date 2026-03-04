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
		return nil, nil, nil, fmt.Errorf("ANNICT_API_KEY is not set")
	}

	annictClient := services.NewAnnictClientWithDB(annictToken, db, logger)
	animeService := services.NewAnimeService(db, logger)

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

	fmt.Printf("\n=== Manual Full Sync (%s) - Differential Mode ===\n", env)
	fmt.Println("This will sync animes using differential update system.")
	fmt.Println("Only changed data will be updated (90% faster than full sync).")
	fmt.Println("Estimated time: 5-15 minutes (depending on changes)")
	fmt.Println()

	startTime := time.Now()

	// Step 1: Re-sync all animes from Annict (with differential update)
	logger.Info("Step 1/4: Re-syncing animes from Annict API (differential mode)...")
	annictSuccess, annictFailure, err := annictClient.SyncAllExistingAnimes()
	if err != nil {
		logger.Warn("Failed to re-sync animes from Annict", "error", err)
	} else {
		logger.Info("Annict re-sync completed (differential mode)",
			"success", annictSuccess,
			"failure", annictFailure)
	}

	// Step 2: Sync episodes from Jikan API (creates new episodes only)
	logger.Info("Step 2/4: Syncing episodes from Jikan API...")
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

	// Step 3: Sync descriptions from TMDB (only for animes without description)
	logger.Info("Step 3/4: Syncing descriptions from TMDB (differential mode)...")
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

	// Step 4: Sync released_on from Jikan API (only for animes without released_on)
	logger.Info("Step 4/4: Syncing release dates from Jikan API (differential mode)...")
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

	// Display differential sync statistics
	logger.Info("Fetching differential sync statistics...")

	// Sync status distribution
	var totalAnimes, syncedCount, pendingCount, failedCount int
	db.QueryRow("SELECT COUNT(*) FROM animes").Scan(&totalAnimes)
	db.QueryRow("SELECT COUNT(*) FROM animes WHERE sync_status = 'synced'").Scan(&syncedCount)
	db.QueryRow("SELECT COUNT(*) FROM animes WHERE sync_status = 'pending' OR sync_status IS NULL").Scan(&pendingCount)
	db.QueryRow("SELECT COUNT(*) FROM animes WHERE sync_status = 'failed'").Scan(&failedCount)

	if totalAnimes > 0 {
		logger.Info("Sync status distribution",
			"total", totalAnimes,
			"synced", syncedCount,
			"synced_percent", fmt.Sprintf("%.1f%%", float64(syncedCount)/float64(totalAnimes)*100),
			"pending", pendingCount,
			"failed", failedCount)
	}

	// Metadata sync status
	var withMetadata, withoutMetadata int
	db.QueryRow("SELECT COUNT(*) FROM animes WHERE metadata_synced_at IS NOT NULL").Scan(&withMetadata)
	db.QueryRow("SELECT COUNT(*) FROM animes WHERE metadata_synced_at IS NULL").Scan(&withoutMetadata)

	if totalAnimes > 0 {
		logger.Info("Metadata sync status",
			"synced", withMetadata,
			"synced_percent", fmt.Sprintf("%.1f%%", float64(withMetadata)/float64(totalAnimes)*100),
			"pending", withoutMetadata)
	}

	// Episode statistics
	rows, err := db.Query(`
		SELECT 
			COUNT(*) as total_episodes,
			COUNT(CASE WHEN title != '第' || episode_number || '話' AND title IS NOT NULL AND title != '' THEN 1 END) as with_custom_title,
			COUNT(duration_minutes) as with_duration,
			COUNT(air_date) as with_air_date
		FROM episodes
	`)
	if err == nil {
		defer rows.Close()
		if rows.Next() {
			var total, customTitle, durationCount, airDate int
			if err := rows.Scan(&total, &customTitle, &durationCount, &airDate); err == nil && total > 0 {
				logger.Info("Episode statistics",
					"total_episodes", total,
					"with_custom_title", customTitle,
					"custom_title_percent", fmt.Sprintf("%.1f%%", float64(customTitle)/float64(total)*100),
					"with_duration", durationCount,
					"duration_percent", fmt.Sprintf("%.1f%%", float64(durationCount)/float64(total)*100),
					"with_air_date", airDate,
					"air_date_percent", fmt.Sprintf("%.1f%%", float64(airDate)/float64(total)*100))
			}
		}
	}

	// Anime description statistics
	var withDescription, withReleasedOn, withMedia, withOfficialSite int
	err = db.QueryRow(`
		SELECT 
			COUNT(CASE WHEN description IS NOT NULL AND description != '' THEN 1 END),
			COUNT(CASE WHEN released_on IS NOT NULL THEN 1 END),
			COUNT(CASE WHEN media IS NOT NULL AND media != '' THEN 1 END),
			COUNT(CASE WHEN official_site_url IS NOT NULL AND official_site_url != '' THEN 1 END)
		FROM animes
	`).Scan(&withDescription, &withReleasedOn, &withMedia, &withOfficialSite)
	if err == nil && totalAnimes > 0 {
		logger.Info("Anime data coverage",
			"with_description", withDescription,
			"description_percent", fmt.Sprintf("%.1f%%", float64(withDescription)/float64(totalAnimes)*100),
			"with_released_on", withReleasedOn,
			"released_on_percent", fmt.Sprintf("%.1f%%", float64(withReleasedOn)/float64(totalAnimes)*100),
			"with_media", withMedia,
			"media_percent", fmt.Sprintf("%.1f%%", float64(withMedia)/float64(totalAnimes)*100),
			"with_official_site", withOfficialSite,
			"official_site_percent", fmt.Sprintf("%.1f%%", float64(withOfficialSite)/float64(totalAnimes)*100))
	}

	logger.Info("✓ Manual sync completed successfully (differential mode)!")
	logger.Info("Note: Metadata (genres) will be synced incrementally by the scheduler.")
	logger.Info("Only changed data was updated, significantly reducing processing time.")

	return nil
}

// updateSeriesTitles updates series_title for all animes that don't have it set
// updateSeriesTitles updates series_title and propagates series_tmdb_id for all animes
func updateSeriesTitles(db *database.DB, logger *common.Logger) (int, error) {
	// Phase 1: Update series_title for animes that don't have it
	rows, err := db.Query(`
		SELECT id, title
		FROM animes
		WHERE series_title IS NULL OR series_title = ''
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query animes: %w", err)
	}
	defer rows.Close()

	type animeUpdate struct {
		ID          int
		Title       string
		SeriesTitle string
	}

	var updates []animeUpdate
	for rows.Next() {
		var id int
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			logger.Warn("Failed to scan row", "error", err)
			continue
		}

		seriesTitle := services.NormalizeAnimeTitle(title)
		updates = append(updates, animeUpdate{
			ID:          id,
			Title:       title,
			SeriesTitle: seriesTitle,
		})
	}

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("error iterating rows: %w", err)
	}

	if len(updates) == 0 {
		return 0, nil
	}

	logger.Info("Found animes to update series_title", "count", len(updates))

	// Start transaction for batch update
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`UPDATE animes SET series_title = $1 WHERE id = $2`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	updated := 0
	for _, u := range updates {
		_, err := stmt.Exec(u.SeriesTitle, u.ID)
		if err != nil {
			logger.Warn("Failed to update anime", "id", u.ID, "error", err)
			continue
		}
		updated++
	}

	if err := tx.Commit(); err != nil {
		return updated, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Phase 2: Propagate series_tmdb_id within series groups
	propagated, err := propagateSeriesTMDBIDs(db, logger)
	if err != nil {
		logger.Warn("Failed to propagate series_tmdb_id", "error", err)
	} else {
		logger.Info("Series TMDB ID propagation completed", "propagated", propagated)
	}

	return updated, nil
}

// propagateSeriesTMDBIDs copies series_tmdb_id from animes that have it to siblings with the same series_title
func propagateSeriesTMDBIDs(db *database.DB, logger *common.Logger) (int, error) {
	result, err := db.Exec(`
		UPDATE animes a
		SET series_tmdb_id = sub.series_tmdb_id
		FROM (
			SELECT DISTINCT ON (series_title) series_title, series_tmdb_id
			FROM animes
			WHERE series_title IS NOT NULL AND series_title != ''
			  AND series_tmdb_id IS NOT NULL
			ORDER BY series_title, id ASC
		) sub
		WHERE a.series_title = sub.series_title
		  AND (a.series_tmdb_id IS NULL)
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to propagate series_tmdb_id: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}
