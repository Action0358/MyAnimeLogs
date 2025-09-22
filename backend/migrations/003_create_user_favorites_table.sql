-- Create user_favorites table
-- Drop existing table if it exists to avoid conflicts
DROP TABLE IF EXISTS user_favorites CASCADE;

CREATE TABLE user_favorites (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    anime_id INTEGER NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'planned' CHECK (status IN ('watching', 'completed', 'planned', 'dropped', 'on_hold')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure one favorite per user per anime
    CONSTRAINT unique_user_anime_favorite UNIQUE(user_id, anime_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_favorites_user_id ON user_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_user_favorites_anime_id ON user_favorites(anime_id);
CREATE INDEX IF NOT EXISTS idx_user_favorites_status ON user_favorites(status);

-- Create trigger for updated_at (reuse function from animes table)
CREATE TRIGGER update_user_favorites_updated_at BEFORE UPDATE ON user_favorites
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();