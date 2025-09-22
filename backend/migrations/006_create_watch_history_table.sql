-- Create watch_history table
-- Drop existing table if it exists to avoid conflicts
DROP TABLE IF EXISTS watch_history CASCADE;

CREATE TABLE watch_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    episode_id INTEGER NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    watched_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_watch_history_user_id ON watch_history(user_id);
CREATE INDEX IF NOT EXISTS idx_watch_history_episode_id ON watch_history(episode_id);
CREATE INDEX IF NOT EXISTS idx_watch_history_watched_at ON watch_history(watched_at);
CREATE INDEX IF NOT EXISTS idx_watch_history_user_watched_at ON watch_history(user_id, watched_at DESC);