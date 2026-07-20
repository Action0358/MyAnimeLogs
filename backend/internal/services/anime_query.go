package services

import (
	"database/sql"
	"fmt"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
	"myanimelogs/internal/models"
)

// queryAnimes is the shared row-iteration core for anime list queries.
// Scan failures on individual rows are logged and skipped; iteration errors fail the query.
// The scan function determines which columns each caller expects.
// db/logger are nil-guarded because QueryAnimeSyncTargets is exported and may be
// called with either unset (mirrors SaveGenresToDatabase's defensive stance).
func queryAnimes(db *database.DB, logger *common.Logger, queryName, query string, args []interface{}, scan func(*sql.Rows) (models.Anime, error)) ([]models.Anime, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %w", queryName, err)
	}
	defer rows.Close()

	animes := []models.Anime{}
	for rows.Next() {
		anime, err := scan(rows)
		if err != nil {
			if logger != nil {
				logger.Warn("Failed to scan anime row", "query", queryName, "error", err)
			}
			continue
		}
		animes = append(animes, anime)
	}

	if err := rows.Err(); err != nil {
		if logger != nil {
			logger.Error("Error iterating anime rows", "query", queryName, "error", err)
		}
		return nil, fmt.Errorf("error iterating %s: %w", queryName, err)
	}

	return animes, nil
}

// scanAnimeSyncTarget scans the minimal column set (id, mal_anime_id, title, annict_id)
// used by sync-target queries
func scanAnimeSyncTarget(rows *sql.Rows) (models.Anime, error) {
	var anime models.Anime
	err := rows.Scan(&anime.ID, &anime.MalAnimeID, &anime.Title, &anime.AnnictID)
	return anime, err
}

// QueryAnimeSyncTargets executes a query that selects sync-target animes
// (id, mal_anime_id, title, annict_id). Exported so the scheduler's metadata
// sync queries share the same row handling as AnimeService.
func QueryAnimeSyncTargets(db *database.DB, logger *common.Logger, queryName, query string, args ...interface{}) ([]models.Anime, error) {
	return queryAnimes(db, logger, queryName, query, args, scanAnimeSyncTarget)
}
