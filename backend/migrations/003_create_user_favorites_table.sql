-- Create user_favorites table
CREATE TABLE IF NOT EXISTS user_favorites (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    anime_id INTEGER NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add missing columns if they don't exist
DO $$ 
BEGIN 
    -- Add status column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='user_favorites' AND column_name='status') THEN
        ALTER TABLE user_favorites ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'planned';
    END IF;
END $$;

-- Add check constraint if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'user_favorites_status_check') THEN
        ALTER TABLE user_favorites ADD CONSTRAINT user_favorites_status_check CHECK (status IN ('watching', 'completed', 'planned', 'dropped', 'on_hold'));
    END IF;
END $$;

-- Add unique constraint if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'unique_user_anime_favorite') THEN
        ALTER TABLE user_favorites ADD CONSTRAINT unique_user_anime_favorite UNIQUE(user_id, anime_id);
    END IF;
END $$;

-- Create indexes for performance
DO $$ 
BEGIN 
    CREATE INDEX IF NOT EXISTS idx_user_favorites_user_id ON user_favorites(user_id);
    CREATE INDEX IF NOT EXISTS idx_user_favorites_anime_id ON user_favorites(anime_id);
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='user_favorites' AND column_name='status') THEN
        CREATE INDEX IF NOT EXISTS idx_user_favorites_status ON user_favorites(status);
    END IF;
END $$;

-- Create trigger for updated_at (reuse function from users table)
DROP TRIGGER IF EXISTS update_user_favorites_updated_at ON user_favorites;
CREATE TRIGGER update_user_favorites_updated_at BEFORE UPDATE ON user_favorites
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();