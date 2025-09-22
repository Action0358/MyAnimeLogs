-- Create episodes table
-- Drop existing table if it exists to avoid conflicts
DROP TABLE IF EXISTS episodes CASCADE;

CREATE TABLE episodes (
    id SERIAL PRIMARY KEY,
    anime_id INTEGER NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    episode_number INTEGER NOT NULL CHECK (episode_number > 0),
    description TEXT,
    air_date DATE,
    duration_minutes INTEGER,
    annict_id INTEGER, -- External ID from Annict API
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure unique episode numbers per anime
    CONSTRAINT unique_anime_episode UNIQUE(anime_id, episode_number)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_episodes_anime_id ON episodes(anime_id);
CREATE INDEX IF NOT EXISTS idx_episodes_episode_number ON episodes(episode_number);
CREATE INDEX IF NOT EXISTS idx_episodes_air_date ON episodes(air_date);
CREATE INDEX IF NOT EXISTS idx_episodes_annict_id ON episodes(annict_id);

-- Create trigger for updated_at (reuse function from animes table)
CREATE TRIGGER update_episodes_updated_at BEFORE UPDATE ON episodes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();