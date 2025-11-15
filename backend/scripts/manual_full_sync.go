//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
	"myanimelogs/internal/services"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env.local"); err != nil {
		log.Printf("Warning: .env.local file not found: %v", err)
	}

	// Initialize logger
	logger := common.NewLogger(common.LogLevelInfo, "text")

	// Initialize database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := database.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize services
	annictToken := os.Getenv("ANNICT_API_KEY")
	if annictToken == "" {
		log.Fatal("ANNICT_API_KEY is not set")
	}

	annictClient := services.NewAnnictClientWithDB(annictToken, db, logger)
	animeService := services.NewAnimeService(db, logger, annictClient)

	fmt.Println("\n=== Manual Full Sync ===")
	fmt.Println("This will sync all episode titles and details.")
	fmt.Println("Estimated time: 40-60 minutes")
	fmt.Println()

	startTime := time.Now()

	// Step 1: Enrich episode titles
	fmt.Println("Step 1/2: Enriching episode titles from Annict API...")
	titleSuccess, titleFailure, titleErrors, err := animeService.EnrichEpisodesWithTitles(nil)
	if err != nil {
		log.Fatalf("Failed to enrich episode titles: %v", err)
	}

	fmt.Printf("Episode titles: Success=%d, Failure=%d\n", titleSuccess, titleFailure)
	if len(titleErrors) > 0 {
		fmt.Printf("Errors (first 5): %v\n", titleErrors[:min(5, len(titleErrors))])
	}

	// Step 2: Sync episode details (air_date and duration)
	fmt.Println("\nStep 2/2: Syncing episode details from Jikan API...")
	detailSuccess, detailFailure, detailErrors, err := animeService.SyncEpisodeDetailsForAllAnimes()
	if err != nil {
		log.Fatalf("Failed to sync episode details: %v", err)
	}

	fmt.Printf("Episode details: Success=%d, Failure=%d\n", detailSuccess, detailFailure)
	if len(detailErrors) > 0 {
		fmt.Printf("Errors (first 5): %v\n", detailErrors[:min(5, len(detailErrors))])
	}

	duration := time.Since(startTime)
	fmt.Printf("\n=== Sync completed in %s ===\n", duration)

	// Display statistics
	fmt.Println("\nFinal statistics:")
	rows, err := db.Query(`
		SELECT 
			COUNT(*) as total_episodes,
			COUNT(CASE WHEN title != '第' || episode_number || '話' AND title IS NOT NULL AND title != '' THEN 1 END) as with_custom_title,
			COUNT(duration_minutes) as with_duration,
			COUNT(air_date) as with_air_date
		FROM episodes
	`)
	if err != nil {
		log.Printf("Failed to get statistics: %v", err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		var total, customTitle, duration, airDate int
		if err := rows.Scan(&total, &customTitle, &duration, &airDate); err != nil {
			log.Printf("Failed to scan statistics: %v", err)
			return
		}

		fmt.Printf("  Total episodes: %d\n", total)
		fmt.Printf("  With custom title: %d (%.1f%%)\n", customTitle, float64(customTitle)/float64(total)*100)
		fmt.Printf("  With duration: %d (%.1f%%)\n", duration, float64(duration)/float64(total)*100)
		fmt.Printf("  With air_date: %d (%.1f%%)\n", airDate, float64(airDate)/float64(total)*100)
	}

	fmt.Println("\n✓ Manual full sync completed successfully!")
	fmt.Println("Future syncs will only update new/changed data.")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
