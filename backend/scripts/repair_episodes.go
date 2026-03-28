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
	env := flag.String("env", "local", "Environment to use: local, staging, production")
	dryRun := flag.Bool("dry-run", false, "Dry run mode: only show what would be repaired")
	dbURLOverride := flag.String("db-url", "", "Override DATABASE_URL (use direct connection instead of pooler)")
	flag.Parse()

	logger := common.NewLogger(common.LogLevelInfo, "text")

	if err := repairEpisodes(logger, *env, *dryRun, *dbURLOverride); err != nil {
		logger.Error("Episode repair failed", "error", err)
		os.Exit(1)
	}
}

func repairEpisodes(logger *common.Logger, env string, dryRun bool, dbURLOverride string) error {
	var envFile string
	switch env {
	case "local":
		envFile = ".env.local"
	case "staging":
		envFile = ".env.staging"
	case "production":
		envFile = ".env.production"
	default:
		return fmt.Errorf("invalid environment: %s", env)
	}

	logger.Info("Loading environment", "env", env, "file", envFile)
	if err := godotenv.Load(envFile); err != nil {
		logger.Warn("Environment file not found", "file", envFile, "error", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURLOverride != "" {
		dbURL = dbURLOverride
		logger.Info("Using overridden DATABASE_URL (direct connection)")
	}
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	db, err := database.New(dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	animeService := services.NewAnimeService(db, logger)

	fmt.Printf("\n=== Episode Repair Script (%s) ===\n", env)
	if dryRun {
		fmt.Println("Mode: DRY RUN (no changes will be made)")
	} else {
		fmt.Println("Mode: LIVE (changes will be applied)")
	}
	fmt.Println()

	startTime := time.Now()

	// Step 1: エピソード数が不足している作品を検出
	logger.Info("Step 1: Detecting animes with missing episodes...")
	rows, err := db.Query(`
		SELECT a.id, a.mal_anime_id, a.title, a.episodes_count,
		       (SELECT COUNT(*) FROM episodes e WHERE e.anime_id = a.id) as actual_count
		FROM animes a
		WHERE a.mal_anime_id IS NOT NULL AND a.mal_anime_id != ''
		AND a.episodes_count IS NOT NULL AND a.episodes_count > 0
		AND a.episodes_count > (SELECT COUNT(*) FROM episodes e WHERE e.anime_id = a.id)
		ORDER BY a.id
	`)
	if err != nil {
		return fmt.Errorf("failed to query animes with missing episodes: %w", err)
	}

	type repairTarget struct {
		ID            int
		MalAnimeID    string
		Title         string
		ExpectedCount int
		ActualCount   int
	}
	var targets []repairTarget

	for rows.Next() {
		var t repairTarget
		if err := rows.Scan(&t.ID, &t.MalAnimeID, &t.Title, &t.ExpectedCount, &t.ActualCount); err != nil {
			logger.Warn("Failed to scan repair target", "error", err)
			continue
		}
		targets = append(targets, t)
	}
	rows.Close()

	logger.Info("Found animes with missing episodes", "count", len(targets))
	for _, t := range targets {
		logger.Info("Missing episodes",
			"anime_id", t.ID,
			"title", t.Title,
			"expected", t.ExpectedCount,
			"actual", t.ActualCount,
			"missing", t.ExpectedCount-t.ActualCount)
	}

	if dryRun {
		logger.Info("Dry run mode - skipping episode repair")
	} else {
		// Step 2: Jikan APIからエピソードを再取得（UPSERT方式）
		logger.Info("Step 2: Repairing episodes from Jikan API (UPSERT mode)...")
		successCount := 0
		failureCount := 0

		for i, t := range targets {
			if i > 0 && i%10 == 0 {
				logger.Info("Repair progress", "processed", i, "total", len(targets))
			}

			count, err := animeService.CreateEpisodesFromJikan(t.ID, t.MalAnimeID)
			if err != nil {
				logger.Warn("Failed to repair episodes",
					"anime_id", t.ID, "title", t.Title, "error", err)
				failureCount++
			} else {
				logger.Info("Repaired episodes",
					"anime_id", t.ID, "title", t.Title, "upserted", count)
				successCount++
			}

			// Jikan APIレート制限: 3req/sec → 400ms wait
			time.Sleep(400 * time.Millisecond)
		}

		logger.Info("Episode repair completed",
			"success", successCount, "failure", failureCount)
	}

	// Step 3: デフォルトタイトルのエピソードを検出
	logger.Info("Step 3: Detecting episodes with default titles...")
	var defaultTitleCount int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM episodes
		WHERE title LIKE '第%話'
	`).Scan(&defaultTitleCount)
	if err != nil {
		logger.Warn("Failed to count default title episodes", "error", err)
	} else {
		logger.Info("Episodes with default titles", "count", defaultTitleCount)
	}

	// デフォルトタイトルのエピソードがある作品を検出
	defaultRows, err := db.Query(`
		SELECT DISTINCT a.id, a.annict_id, a.title,
		       COUNT(e.id) as default_title_count
		FROM animes a
		INNER JOIN episodes e ON a.id = e.anime_id
		WHERE a.annict_id IS NOT NULL
		AND e.title LIKE '第%話'
		GROUP BY a.id, a.annict_id, a.title
		ORDER BY default_title_count DESC
		LIMIT 50
	`)
	if err != nil {
		logger.Warn("Failed to query animes with default titles", "error", err)
	} else {
		for defaultRows.Next() {
			var animeID int
			var annictID int
			var title string
			var count int
			if err := defaultRows.Scan(&animeID, &annictID, &title, &count); err != nil {
				continue
			}
			logger.Info("Anime with default episode titles",
				"anime_id", animeID,
				"annict_id", annictID,
				"title", title,
				"default_title_episodes", count)
		}
		defaultRows.Close()
	}

	duration := time.Since(startTime)
	logger.Info("Episode repair script completed", "duration", duration.String())

	return nil
}
