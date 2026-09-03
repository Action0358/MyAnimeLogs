-- 詳細ページの「同シリーズの他シーズン」表示（#755）で series_title を検索条件に使うため、
-- インデックスを追加する。NULLの行は対象外（WHERE句で除外）にしてサイズを抑える。
--
-- 冪等性:
-- IF NOT EXISTS のため再実行しても影響はない。
CREATE INDEX IF NOT EXISTS idx_animes_series_title
  ON animes (series_title)
  WHERE series_title IS NOT NULL;
