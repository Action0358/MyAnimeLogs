-- Add series_title column that was lost due to migration 022 executing both Up and Down sections
-- series_title: タイトル正規化後のシリーズ名（内部識別用）

ALTER TABLE animes ADD COLUMN IF NOT EXISTS series_title VARCHAR(255);
CREATE INDEX IF NOT EXISTS idx_animes_series_title ON animes(series_title);

-- Also ensure released_on is VARCHAR(255) type for flexible date format support
-- (migration 020 created it as DATE, migration 023 should have changed it to VARCHAR)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'animes' AND column_name = 'released_on' AND data_type = 'date'
    ) THEN
        ALTER TABLE animes ALTER COLUMN released_on TYPE VARCHAR(255);
    END IF;
END $$;
