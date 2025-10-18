-- Create animes table (enhanced version)
CREATE TABLE IF NOT EXISTS animes (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ensure id column has proper sequence (idempotent)
DO $$
BEGIN
    -- Check if sequence exists, if not create it
    IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'animes_id_seq') THEN
        CREATE SEQUENCE animes_id_seq;
    END IF;
    
    -- Ensure id column uses the sequence
    ALTER TABLE animes ALTER COLUMN id SET DEFAULT nextval('animes_id_seq');
    ALTER SEQUENCE animes_id_seq OWNED BY animes.id;
    
    -- Set sequence to current max id
    PERFORM setval('animes_id_seq', COALESCE((SELECT MAX(id) FROM animes), 0) + 1, false);
END $$;

-- Add all missing columns if they don't exist
DO $$ 
BEGIN 
    -- Add title_kana column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='title_kana') THEN
        ALTER TABLE animes ADD COLUMN title_kana VARCHAR(255);
    END IF;
    
    -- Add title_en column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='title_en') THEN
        ALTER TABLE animes ADD COLUMN title_en VARCHAR(255);
    END IF;
    
    -- Add image_url column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='image_url') THEN
        ALTER TABLE animes ADD COLUMN image_url VARCHAR(500);
    END IF;
    
    -- Add season column (format: YYYY-season, e.g., "2025-summer")
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='season') THEN
        ALTER TABLE animes ADD COLUMN season VARCHAR(50);
    END IF;
    
    -- Add status column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='status') THEN
        ALTER TABLE animes ADD COLUMN status VARCHAR(50);
    END IF;
    
    -- Add media column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='media') THEN
        ALTER TABLE animes ADD COLUMN media VARCHAR(50);
    END IF;
    
    -- Add episodes_count column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='episodes_count') THEN
        ALTER TABLE animes ADD COLUMN episodes_count INTEGER DEFAULT 0;
    END IF;
    
    -- Add official_site_url column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='official_site_url') THEN
        ALTER TABLE animes ADD COLUMN official_site_url VARCHAR(500);
    END IF;
    
    -- Add wikipedia_url column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='wikipedia_url') THEN
        ALTER TABLE animes ADD COLUMN wikipedia_url VARCHAR(500);
    END IF;
    
    -- Add twitter_username column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='twitter_username') THEN
        ALTER TABLE animes ADD COLUMN twitter_username VARCHAR(100);
    END IF;
    
    -- Add twitter_hashtag column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='twitter_hashtag') THEN
        ALTER TABLE animes ADD COLUMN twitter_hashtag VARCHAR(100);
    END IF;
    
    -- Add mal_anime_id column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='mal_anime_id') THEN
        ALTER TABLE animes ADD COLUMN mal_anime_id VARCHAR(50);
    END IF;
    
    -- Add annict_id column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='annict_id') THEN
        ALTER TABLE animes ADD COLUMN annict_id INTEGER;
    END IF;
    
    -- Add watchers_count column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='watchers_count') THEN
        ALTER TABLE animes ADD COLUMN watchers_count INTEGER DEFAULT 0;
    END IF;
    
    -- Add reviews_count column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='reviews_count') THEN
        ALTER TABLE animes ADD COLUMN reviews_count INTEGER DEFAULT 0;
    END IF;
END $$;

-- Add unique constraint to annict_id if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'animes_annict_id_key') THEN
        ALTER TABLE animes ADD CONSTRAINT animes_annict_id_key UNIQUE (annict_id);
    END IF;
END $$;

-- Create indexes for performance (after ensuring columns exist)
DO $$ 
BEGIN 
    -- Create basic indexes
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='season') THEN
        CREATE INDEX IF NOT EXISTS idx_animes_season ON animes(season);
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='status') THEN
        CREATE INDEX IF NOT EXISTS idx_animes_status ON animes(status);
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='annict_id') THEN
        CREATE INDEX IF NOT EXISTS idx_animes_annict_id ON animes(annict_id);
    END IF;
    
    -- Create composite index for status and season
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='status') 
       AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='season') THEN
        CREATE INDEX IF NOT EXISTS idx_animes_status_season ON animes(status, season);
    END IF;
    
    -- Create text search indexes
    CREATE INDEX IF NOT EXISTS idx_animes_title ON animes USING gin(to_tsvector('simple', title));
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='title_kana') THEN
        CREATE INDEX IF NOT EXISTS idx_animes_title_kana ON animes USING gin(to_tsvector('simple', COALESCE(title_kana, '')));
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='title_en') THEN
        CREATE INDEX IF NOT EXISTS idx_animes_title_en ON animes USING gin(to_tsvector('english', COALESCE(title_en, '')));
    END IF;
    
    -- Create index for MAL anime ID (used for Jikan API lookups)
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='mal_anime_id') THEN
        CREATE INDEX IF NOT EXISTS idx_animes_mal_anime_id ON animes(mal_anime_id) WHERE mal_anime_id IS NOT NULL;
    END IF;
END $$;

-- Create trigger for updated_at (function already exists from users table)
-- Drop trigger if exists before creating
DROP TRIGGER IF EXISTS update_animes_updated_at ON animes;
CREATE TRIGGER update_animes_updated_at BEFORE UPDATE ON animes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
