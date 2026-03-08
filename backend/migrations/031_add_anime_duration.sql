-- 映画作品の上映時間フィールドを追加
ALTER TABLE animes
ADD COLUMN duration_minutes INTEGER;

COMMENT ON COLUMN animes.duration_minutes IS '映画作品の上映時間（分）。TVアニメの場合はNULL。Jikan APIから取得。';
