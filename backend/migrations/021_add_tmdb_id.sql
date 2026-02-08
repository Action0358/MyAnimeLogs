-- Add TMDB ID column to animes table
ALTER TABLE animes ADD COLUMN IF NOT EXISTS tmdb_id INTEGER;

-- Create index for TMDB ID lookups
CREATE INDEX IF NOT EXISTS idx_animes_tmdb_id ON animes(tmdb_id);
