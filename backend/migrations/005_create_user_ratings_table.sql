-- Create user_ratings table
CREATE TABLE IF NOT EXISTS user_ratings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    anime_id INTEGER NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
    rating DECIMAL(3,2) NOT NULL CHECK (rating >= 1.0 AND rating <= 5.0),
    comment TEXT CHECK (LENGTH(comment) <= 1000),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure one rating per user per anime
    CONSTRAINT unique_user_anime_rating UNIQUE(user_id, anime_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_ratings_user_id ON user_ratings(user_id);
CREATE INDEX IF NOT EXISTS idx_user_ratings_anime_id ON user_ratings(anime_id);
CREATE INDEX IF NOT EXISTS idx_user_ratings_rating ON user_ratings(rating);

-- Create trigger for updated_at (reuse function from users table)
DROP TRIGGER IF EXISTS update_user_ratings_updated_at ON user_ratings;
CREATE TRIGGER update_user_ratings_updated_at BEFORE UPDATE ON user_ratings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();