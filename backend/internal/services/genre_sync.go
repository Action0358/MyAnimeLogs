package services

import (
	"fmt"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
)

// saveGenresToDatabase saves genre information to the database
// This is a shared function used by both AnnictClient and AnimeService
func saveGenresToDatabase(db *database.DB, logger *common.Logger, animeID int, genres []string) error {
	if db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Begin transaction to ensure atomicity
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Always delete existing genres to ensure data consistency
	// Even if the new list is empty, we want to clear stale data
	_, err = tx.Exec(`DELETE FROM anime_genres WHERE anime_id = $1`, animeID)
	if err != nil {
		return fmt.Errorf("failed to delete existing genres: %w", err)
	}

	insertedCount := 0

	// Insert new genres if any exist
	if len(genres) > 0 {
		// Prepare insert statement for better performance
		stmt, err := tx.Prepare(`
			INSERT INTO anime_genres (anime_id, genre)
			VALUES ($1, $2)
			ON CONFLICT (anime_id, genre) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement: %w", err)
		}
		defer stmt.Close()

		// Insert new genres using prepared statement
		for _, genre := range genres {
			_, err := stmt.Exec(animeID, genre)
			if err != nil {
				if logger != nil {
					logger.Warn("Failed to insert genre",
						"anime_id", animeID,
						"genre", genre,
						"error", err)
				}
				continue
			}
			insertedCount++
		}
	}

	// Update genres_synced_at timestamp
	_, err = tx.Exec(`UPDATE animes SET genres_synced_at = CURRENT_TIMESTAMP WHERE id = $1`, animeID)
	if err != nil {
		return fmt.Errorf("failed to update genres_synced_at: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if logger != nil {
		logger.Debug("Saved genres to database",
			"anime_id", animeID,
			"genre_count", insertedCount)
	}

	return nil
}
