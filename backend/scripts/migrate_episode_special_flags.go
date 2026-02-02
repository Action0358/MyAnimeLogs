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
		logger.Error("Episode special flags migration failed", "error", err)
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

	fmt.Println("\n=== Episode Special Flags Migration ===")
	fmt.Println("This will update episode special flags (filler, recap, is_special)")
	fmt.Println("using data from Jikan API.")
	fmt.Println()

	startTime := time.Now()

	// Get animes with MAL ID that have episodes
	rows, err := db.Query(`
		SELECT DISTINCT a.id, a.mal_anime_id
		FROM animes a
		INNER JOIN episodes e ON a.id = e.anime_id
		WHERE a.mal_anime_id IS NOT NULL AND a.mal_anime_id != ''
		ORDER BY a.id
	`)
	if err != nil {
		return fmt.Errorf("failed to query animes: %w", err)
	}
	defer rows.Close()

	type animeInfo struct {
		ID         int
		MalAnimeID string
	}
	var animes []animeInfo

	for rows.Next() {
		var anime animeInfo
		if err := rows.Scan(&anime.ID, &anime.MalAnimeID); err != nil {
			logger.Warn("Failed to scan anime", "error", err)
			continue
		}
		animes = append(animes, anime)
	}

	logger.Info("Starting episode special flags migration", "total_animes", len(animes))

	// Initialize Jikan client
	jikanClient := services.NewJikanClient(logger)

	successCount := 0
	errorCount := 0
	updatedEpisodes := 0

	for i, anime := range animes {
		logger.Info("Processing anime",
			"progress", fmt.Sprintf("%d/%d", i+1, len(animes)),
			"anime_id", anime.ID,
			"mal_anime_id", anime.MalAnimeID)

		// Get episodes from Jikan API
		jikanEpisodes, err := jikanClient.GetAnimeEpisodes(anime.MalAnimeID)
		if err != nil {
			logger.Warn("Failed to get episodes from Jikan",
				"anime_id", anime.ID,
				"mal_anime_id", anime.MalAnimeID,
				"error", err)
			errorCount++
			continue
		}

		if len(jikanEpisodes) == 0 {
			logger.Debug("No episodes found in Jikan API",
				"anime_id", anime.ID,
				"mal_anime_id", anime.MalAnimeID)
			continue
		}

		// Update episode flags
		for _, jikanEp := range jikanEpisodes {
			isSpecial := jikanEp.Filler || jikanEp.Recap

			result, err := db.Exec(`
				UPDATE episodes
				SET filler = $1, recap = $2, is_special = $3, updated_at = CURRENT_TIMESTAMP
				WHERE anime_id = $4 AND episode_number = $5
			`, jikanEp.Filler, jikanEp.Recap, isSpecial, anime.ID, jikanEp.MalID)

			if err != nil {
				logger.Warn("Failed to update episode",
					"anime_id", anime.ID,
					"episode_number", jikanEp.MalID,
					"error", err)
				continue
			}

			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				updatedEpisodes++
			}
		}

		successCount++

		// Rate limiting: Jikan API allows 3 requests/second
		time.Sleep(400 * time.Millisecond)
	}

	duration := time.Since(startTime)

	logger.Info("Episode special flags migration completed",
		"duration", duration.String(),
		"total_animes", len(animes),
		"success", successCount,
		"errors", errorCount,
		"updated_episodes", updatedEpisodes)

	// Display statistics
	logger.Info("Fetching final statistics...")
	statsRows, err := db.Query(`
		SELECT 
			COUNT(*) as total_episodes,
			COUNT(CASE WHEN is_special = true THEN 1 END) as special_episodes,
			COUNT(CASE WHEN filler = true THEN 1 END) as filler_episodes,
			COUNT(CASE WHEN recap = true THEN 1 END) as recap_episodes
		FROM episodes
	`)
	if err != nil {
		return fmt.Errorf("failed to get statistics: %w", err)
	}
	defer statsRows.Close()

	if statsRows.Next() {
		var total, special, filler, recap int
		if err := statsRows.Scan(&total, &special, &filler, &recap); err != nil {
			return fmt.Errorf("failed to scan statistics: %w", err)
		}

		logger.Info("Final statistics",
			"total_episodes", total,
			"special_episodes", special,
			"special_percent", fmt.Sprintf("%.2f%%", float64(special)/float64(total)*100),
			"filler_episodes", filler,
			"recap_episodes", recap)
	}

	logger.Info("✓ Episode special flags migration completed successfully!")

	return nil
}
