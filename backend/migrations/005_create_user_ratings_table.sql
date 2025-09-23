-- Create user_ratings table
CREATE TABLE IF NOT EXISTS user_ratings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    anime_id INTEGER NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add missing columns if they don't exist
DO $$ 
BEGIN 
    -- Add rating column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='user_ratings' AND column_name='rating') THEN
        ALTER TABLE user_ratings ADD COLUMN rating DECIMAL(3,2) NOT NULL DEFAULT 3.0;
    END IF;
    
    -- Add comment column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='user_ratings' AND column_name='comment') THEN
        ALTER TABLE user_ratings ADD COLUMN comment TEXT;
    END IF;
END $$;

-- Add check constraints if they don't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'user_ratings_rating_check') THEN
        ALTER TABLE user_ratings ADD CONSTRAINT user_ratings_rating_check CHECK (rating >= 1.0 AND rating <= 5.0);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'user_ratings_comment_check') THEN
        ALTER TABLE user_ratings ADD CONSTRAINT user_ratings_comment_check CHECK (LENGTH(comment) <= 1000);
    END IF;
END $$;

-- Add unique constraint if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'unique_user_anime_rating') THEN
        ALTER TABLE user_ratings ADD CONSTRAINT unique_user_anime_rating UNIQUE(user_id, anime_id);
    END IF;
END $$;

-- Create indexes for performance
DO $$ 
BEGIN 
    CREATE INDEX IF NOT EXISTS idx_user_ratings_user_id ON user_ratings(user_id);
    CREATE INDEX IF NOT EXISTS idx_user_ratings_anime_id ON user_ratings(anime_id);
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='user_ratings' AND column_name='rating') THEN
        CREATE INDEX IF NOT EXISTS idx_user_ratings_rating ON user_ratings(rating);
    END IF;
END $$;

-- Create trigger for updated_at (reuse function from users table)
DROP TRIGGER IF EXISTS update_user_ratings_updated_at ON user_ratings;
CREATE TRIGGER update_user_ratings_updated_at BEFORE UPDATE ON user_ratings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();