-- エピソードのUNIQUE制約を確実に追加（UPSERT用）
-- 重複データがある場合は先にクリーンアップ

-- 重複エピソードを削除（IDが大きい方を残す）
DELETE FROM episodes a
USING episodes b
WHERE a.anime_id = b.anime_id
  AND a.episode_number = b.episode_number
  AND a.id < b.id;

-- UNIQUE制約を追加（既存の制約名に合わせる）
-- unique_anime_episode または unique_anime_episode_type が存在しない場合のみ追加
ALTER TABLE episodes DROP CONSTRAINT IF EXISTS unique_anime_episode;
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'unique_anime_episode_type' AND conrelid = 'episodes'::regclass) THEN
        ALTER TABLE episodes ADD CONSTRAINT unique_anime_episode_type UNIQUE (anime_id, episode_number);
    END IF;
END $$;
