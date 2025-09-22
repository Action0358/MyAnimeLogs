-- Create anime_genres table
-- Drop existing table if it exists to avoid conflicts
DROP TABLE IF EXISTS anime_genres CASCADE;

CREATE TABLE anime_genres (
    id SERIAL PRIMARY KEY,
    anime_id INTEGER NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
    genre VARCHAR(100) NOT NULL
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_anime_genres_anime_id ON anime_genres(anime_id);
CREATE INDEX IF NOT EXISTS idx_anime_genres_genre ON anime_genres(genre);

-- Create unique constraint to prevent duplicate genre assignments
CREATE UNIQUE INDEX IF NOT EXISTS idx_anime_genres_unique ON anime_genres(anime_id, genre);