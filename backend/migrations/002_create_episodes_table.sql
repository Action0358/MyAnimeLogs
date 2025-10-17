-- Create episodes table
CREATE TABLE IF NOT EXISTS episodes (
    id SERIAL PRIMARY KEY,
    anime_id INTEGER NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    episode_number INTEGER NOT NULL CHECK (episode_number > 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add missing columns if they don't exist
DO $$ 
BEGIN 
    -- Add air_date column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='episodes' AND column_name='air_date') THEN
        ALTER TABLE episodes ADD COLUMN air_date DATE;
    END IF;
    
    -- Add duration_minutes column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='episodes' AND column_name='duration_minutes') THEN
        ALTER TABLE episodes ADD COLUMN duration_minutes INTEGER;
    END IF;
END $$;

-- Add unique constraint if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'unique_anime_episode') THEN
        ALTER TABLE episodes ADD CONSTRAINT unique_anime_episode UNIQUE(anime_id, episode_number);
    END IF;
END $$;

-- Create indexes for performance (after ensuring columns exist)
DO $$ 
BEGIN 
    CREATE INDEX IF NOT EXISTS idx_episodes_anime_id ON episodes(anime_id);
    CREATE INDEX IF NOT EXISTS idx_episodes_episode_number ON episodes(episode_number);
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='episodes' AND column_name='air_date') THEN
        CREATE INDEX IF NOT EXISTS idx_episodes_air_date ON episodes(air_date);
    END IF;
END $$;

-- Create trigger for updated_at (reuse function from users table)
DROP TRIGGER IF EXISTS update_episodes_updated_at ON episodes;
CREATE TRIGGER update_episodes_updated_at BEFORE UPDATE ON episodes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
