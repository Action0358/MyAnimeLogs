-- Create animes table (enhanced version)
CREATE TABLE IF NOT EXISTS animes (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    title_kana VARCHAR(255),
    title_en VARCHAR(255),
    description TEXT DEFAULT '',
    image_url VARCHAR(500),
    season VARCHAR(50), -- e.g., "2024-winter", "2024-spring"
    season_year INTEGER,
    status VARCHAR(50), -- e.g., "airing", "finished", "upcoming"
    media VARCHAR(50), -- e.g., "tv", "movie", "ova"
    episodes_count INTEGER DEFAULT 0,
    official_site_url VARCHAR(500),
    wikipedia_url VARCHAR(500),
    twitter_username VARCHAR(100),
    twitter_hashtag VARCHAR(100),
    mal_anime_id VARCHAR(50),
    annict_id INTEGER UNIQUE, -- External ID from Annict API
    watchers_count INTEGER DEFAULT 0,
    reviews_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_animes_season ON animes(season);
CREATE INDEX IF NOT EXISTS idx_animes_status ON animes(status);
CREATE INDEX IF NOT EXISTS idx_animes_season_year ON animes(season_year);
-- Use simple text search instead of Japanese-specific configuration
CREATE INDEX IF NOT EXISTS idx_animes_title ON animes USING gin(to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS idx_animes_title_kana ON animes USING gin(to_tsvector('simple', COALESCE(title_kana, '')));
CREATE INDEX IF NOT EXISTS idx_animes_title_en ON animes USING gin(to_tsvector('english', COALESCE(title_en, '')));
CREATE INDEX IF NOT EXISTS idx_animes_annict_id ON animes(annict_id);

-- Create trigger for updated_at (function already exists from users table)
-- Drop trigger if exists before creating
DROP TRIGGER IF EXISTS update_animes_updated_at ON animes;
CREATE TRIGGER update_animes_updated_at BEFORE UPDATE ON animes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();