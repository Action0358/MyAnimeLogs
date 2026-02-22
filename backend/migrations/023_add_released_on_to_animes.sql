-- +migrate Up
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='released_on') THEN
        ALTER TABLE animes ADD COLUMN released_on VARCHAR(255);
    ELSE
        -- 既に DATE 型などで存在する場合、VARCHAR に変更して柔軟な日付形式に対応可能にする
        ALTER TABLE animes ALTER COLUMN released_on TYPE VARCHAR(255);
    END IF;
END $$;

-- +migrate Down
-- 注意: Down マイグレーションは元の型(DATE)に戻すのが理想ですが、データの損失を防ぐため DROP のみ記述
ALTER TABLE animes DROP COLUMN IF EXISTS released_on;
