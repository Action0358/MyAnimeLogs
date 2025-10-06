-- Add Anilist image columns to animes table
-- これにより、Anilist APIから取得した画像URLをデータベースに保存できます

DO $$ 
BEGIN 
    -- Add cover_image_large column (Anilist ExtraLarge/Large)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='cover_image_large') THEN
        ALTER TABLE animes ADD COLUMN cover_image_large VARCHAR(500);
    END IF;
    
    -- Add cover_image_medium column (Anilist Medium)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='cover_image_medium') THEN
        ALTER TABLE animes ADD COLUMN cover_image_medium VARCHAR(500);
    END IF;
    
    -- Add cover_image_updated_at column (画像の更新日時を記録)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='cover_image_updated_at') THEN
        ALTER TABLE animes ADD COLUMN cover_image_updated_at TIMESTAMP;
    END IF;
END $$;

-- Create index on mal_anime_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_animes_mal_anime_id ON animes(mal_anime_id) WHERE mal_anime_id IS NOT NULL;

-- Create index on cover_image_updated_at for TTL queries
CREATE INDEX IF NOT EXISTS idx_animes_cover_image_updated_at ON animes(cover_image_updated_at) WHERE cover_image_updated_at IS NOT NULL;

-- Add comment to table
COMMENT ON TABLE animes IS 'アニメ情報テーブル。Annict APIの基本情報とAnilist APIの画像URLを保存';
