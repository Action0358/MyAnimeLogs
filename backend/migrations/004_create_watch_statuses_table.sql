-- Create watch_statuses table
-- Drop existing table if it exists to avoid conflicts
DROP TABLE IF EXISTS watch_statuses CASCADE;

CREATE TABLE watch_statuses (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    episode_id INTEGER NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    viewed BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure one watch status per user per episode
    CONSTRAINT unique_user_episode_watch UNIQUE(user_id, episode_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_watch_statuses_user_id ON watch_statuses(user_id);
CREATE INDEX IF NOT EXISTS idx_watch_statuses_episode_id ON watch_statuses(episode_id);
CREATE INDEX IF NOT EXISTS idx_watch_statuses_viewed ON watch_statuses(viewed);

-- Create trigger for updated_at (reuse function from animes table)
CREATE TRIGGER update_watch_statuses_updated_at BEFORE UPDATE ON watch_statuses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();