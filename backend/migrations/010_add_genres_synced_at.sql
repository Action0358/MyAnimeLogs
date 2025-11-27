-- Add genres_synced_at column to animes table if it doesn't exist
-- This migration ensures the column exists for genre sync functionality

DO $$ 
BEGIN 
    -- Add genres_synced_at column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='genres_synced_at') THEN
        ALTER TABLE animes ADD COLUMN genres_synced_at TIMESTAMP;
        RAISE NOTICE 'Added genres_synced_at column to animes table';
    ELSE
        RAISE NOTICE 'genres_synced_at column already exists in animes table';
    END IF;
END $$;

-- Create index for genres_synced_at (used for efficient genre sync queries)
-- Only create if the column exists and index doesn't exist
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='genres_synced_at') THEN
        CREATE INDEX IF NOT EXISTS idx_animes_genres_synced_at ON animes(genres_synced_at) 
            WHERE mal_anime_id IS NOT NULL AND mal_anime_id != '';
        RAISE NOTICE 'Created index idx_animes_genres_synced_at';
    END IF;
END $$;
