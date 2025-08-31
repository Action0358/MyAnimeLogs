package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func New(databaseURL string) (*DB, error) {
	var db *sql.DB
	var err error

	// Retry connection for a few times
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", databaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}

		err = db.Ping()
		if err == nil {
			// Connection successful
			log.Println("Successfully connected to the database.")
			return &DB{db}, nil
		}

		log.Printf("Failed to ping database (attempt %d/%d): %v. Retrying in 5 seconds...", i+1, maxRetries, err)
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

func Migrate(db *DB) error {
	// Run database migrations
	migrations := []string{
		createUsersTable,
		createAnimesTable,
		createEpisodesTable,
		createUserFavoritesTable,
		createWatchStatusesTable,
		createUserRatingsTable,
		createWatchHistoryTable,
		createAnimeGenresTable,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	return nil
}

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	email VARCHAR(255) NOT NULL UNIQUE,
	password VARCHAR(255) NOT NULL,
	description TEXT DEFAULT '',
	icon VARCHAR(500),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const createAnimesTable = `
CREATE TABLE IF NOT EXISTS animes (
	id INTEGER PRIMARY KEY,
	title VARCHAR(500) NOT NULL,
	description TEXT DEFAULT '',
	image_url VARCHAR(500),
	season VARCHAR(50),
	status VARCHAR(50),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_animes_season ON animes(season);
CREATE INDEX IF NOT EXISTS idx_animes_status ON animes(status);`

const createEpisodesTable = `
CREATE TABLE IF NOT EXISTS episodes (
	id INTEGER PRIMARY KEY,
	anime_id INTEGER NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
	title VARCHAR(500) NOT NULL,
	episode_number INTEGER NOT NULL CHECK (episode_number > 0),
	UNIQUE(anime_id, episode_number)
);
CREATE INDEX IF NOT EXISTS idx_episodes_anime_id ON episodes(anime_id);`

const createUserFavoritesTable = `
CREATE TABLE IF NOT EXISTS user_favorites (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	anime_id INTEGER NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
	status VARCHAR(50) NOT NULL DEFAULT 'planned',
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(user_id, anime_id)
);
CREATE INDEX IF NOT EXISTS idx_user_favorites_user_id ON user_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_user_favorites_anime_id ON user_favorites(anime_id);
CREATE INDEX IF NOT EXISTS idx_user_favorites_status ON user_favorites(status);`

const createWatchStatusesTable = `
CREATE TABLE IF NOT EXISTS watch_statuses (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	episode_id INTEGER NOT NULL REFERENCES episodes(id) ON DELETE NO ACTION,
	viewed BOOLEAN NOT NULL DEFAULT FALSE,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(user_id, episode_id)
);
CREATE INDEX IF NOT EXISTS idx_watch_statuses_user_id ON watch_statuses(user_id);
CREATE INDEX IF NOT EXISTS idx_watch_statuses_episode_id ON watch_statuses(episode_id);`

const createUserRatingsTable = `
CREATE TABLE IF NOT EXISTS user_ratings (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	anime_id INTEGER NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
	rating DECIMAL(2,1) NOT NULL CHECK (rating >= 1.0 AND rating <= 5.0),
	comment TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(user_id, anime_id)
);
CREATE INDEX IF NOT EXISTS idx_user_ratings_user_id ON user_ratings(user_id);
CREATE INDEX IF NOT EXISTS idx_user_ratings_anime_id ON user_ratings(anime_id);`

const createWatchHistoryTable = `
CREATE TABLE IF NOT EXISTS watch_history (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	episode_id INTEGER NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
	watched_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_watch_history_user_id ON watch_history(user_id);
CREATE INDEX IF NOT EXISTS idx_watch_history_episode_id ON watch_history(episode_id);
CREATE INDEX IF NOT EXISTS idx_watch_history_watched_at ON watch_history(watched_at);`

const createAnimeGenresTable = `
CREATE TABLE IF NOT EXISTS anime_genres (
	id SERIAL PRIMARY KEY,
	anime_id INTEGER NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
	genre VARCHAR(50) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_anime_genres_anime_id ON anime_genres(anime_id);
CREATE INDEX IF NOT EXISTS idx_anime_genres_genre ON anime_genres(genre);`
