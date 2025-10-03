-- Create search_logs table
CREATE TABLE IF NOT EXISTS search_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    query VARCHAR(500) NOT NULL,
    filters JSONB,
    result_count INTEGER NOT NULL DEFAULT 0,
    clicked_anime_id INTEGER REFERENCES animes(id) ON DELETE SET NULL,
    searched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_search_logs_query ON search_logs(query);
CREATE INDEX IF NOT EXISTS idx_search_logs_user_id ON search_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_search_logs_searched_at ON search_logs(searched_at);
CREATE INDEX IF NOT EXISTS idx_search_logs_filters ON search_logs USING gin(filters);

-- Create full-text search indexes for animes table
CREATE INDEX IF NOT EXISTS idx_animes_fulltext_title_desc ON animes USING gin(
    to_tsvector('simple', COALESCE(title, '') || ' ' || COALESCE(description, ''))
);

-- Create composite indexes for advanced search
CREATE INDEX IF NOT EXISTS idx_animes_season_year ON animes(season, season_year);
CREATE INDEX IF NOT EXISTS idx_animes_status_season ON animes(status, season);

-- Create index for anime_genres search optimization
CREATE INDEX IF NOT EXISTS idx_anime_genres_search ON anime_genres(genre, anime_id);