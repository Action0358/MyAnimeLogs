-- Create watch_statuses table
CREATE TABLE IF NOT EXISTS watch_statuses (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    episode_id INTEGER NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add missing columns if they don't exist
DO $$ 
BEGIN 
    -- Add viewed column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='watch_statuses' AND column_name='viewed') THEN
        ALTER TABLE watch_statuses ADD COLUMN viewed BOOLEAN NOT NULL DEFAULT FALSE;
    END IF;
END $$;

-- Add unique constraint if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'unique_user_episode_watch') THEN
        ALTER TABLE watch_statuses ADD CONSTRAINT unique_user_episode_watch UNIQUE(user_id, episode_id);
    END IF;
END $$;

-- Create indexes for performance
DO $$ 
BEGIN 
    CREATE INDEX IF NOT EXISTS idx_watch_statuses_user_id ON watch_statuses(user_id);
    CREATE INDEX IF NOT EXISTS idx_watch_statuses_episode_id ON watch_statuses(episode_id);
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='watch_statuses' AND column_name='viewed') THEN
        CREATE INDEX IF NOT EXISTS idx_watch_statuses_viewed ON watch_statuses(viewed);
    END IF;
END $$;

-- Create trigger for updated_at (reuse function from users table)
DROP TRIGGER IF EXISTS update_watch_statuses_updated_at ON watch_statuses;
CREATE TRIGGER update_watch_statuses_updated_at BEFORE UPDATE ON watch_statuses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();