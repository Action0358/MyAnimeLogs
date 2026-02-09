-- Add description and released_on fields for anime detail page
-- description: あらすじ（AniList APIから取得）
-- released_on: 放送開始日（Annict APIから取得）

DO $$
BEGIN
    -- Add description column (あらすじ)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='description') THEN
        ALTER TABLE animes ADD COLUMN description TEXT;
    END IF;

    -- Add released_on column (放送開始日)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='released_on') THEN
        ALTER TABLE animes ADD COLUMN released_on DATE;
    END IF;
END $$;
