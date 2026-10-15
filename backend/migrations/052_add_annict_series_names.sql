-- +migrate Up
-- Annictの作品シリーズ（GraphQL Work.seriesList）を保存し、関連作品の判定に使う。
--
-- 関連作品はこれまでタイトルからの推測（series_title・前方一致・先頭語一致）のみで、
-- 「銀魂オンシアター2D」のように空白なしで副題が続く作品は「銀魂」と繋がらず、
-- 逆に「亜人」と「亜人ちゃんは語りたい」のような別作品を区別できなかった。
-- Annictのシリーズは人手で整理された所属情報のため、これを主な手がかりにする。
--
-- annict_series_names: シリーズ名の配列。所属シリーズが無い作品は空配列
-- annict_series_synced_at: シリーズを取得した日時。NULLは未取得（補完対象）
ALTER TABLE animes ADD COLUMN IF NOT EXISTS annict_series_names TEXT[];
ALTER TABLE animes ADD COLUMN IF NOT EXISTS annict_series_synced_at TIMESTAMPTZ;

-- 関連作品検索の配列重なり判定（&&）用
CREATE INDEX IF NOT EXISTS idx_animes_annict_series_names ON animes USING GIN (annict_series_names);

-- +migrate Down
DROP INDEX IF EXISTS idx_animes_annict_series_names;
ALTER TABLE animes DROP COLUMN IF EXISTS annict_series_synced_at;
ALTER TABLE animes DROP COLUMN IF EXISTS annict_series_names;
