-- Add cover_image_extra_large column to animes table
-- AniListのextraLargeサイズ（最高画質）を保存するためのカラム

DO $$ 
BEGIN 
    -- Add cover_image_extra_large column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='cover_image_extra_large') THEN
        ALTER TABLE animes ADD COLUMN cover_image_extra_large VARCHAR(500);
    END IF;
END $$;

-- Create index on cover_image_extra_large for faster lookups
CREATE INDEX IF NOT EXISTS idx_animes_cover_image_extra_large ON animes(cover_image_extra_large) WHERE cover_image_extra_large IS NOT NULL;

COMMENT ON COLUMN animes.cover_image_extra_large IS 'AniListから取得した最高画質のカバー画像URL（extraLargeサイズ）';
