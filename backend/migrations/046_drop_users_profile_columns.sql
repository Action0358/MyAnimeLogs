-- プロフィールの自由記述項目を廃止する（#789）
--
-- 背景:
-- 公開プロフィールページは存在せず、入力された内容が他ユーザーの目に触れる
-- 経路が無いため、需要が無いと判断した。表示名(name)はGoogleアカウント由来の
-- 値で維持し、ユーザーによる編集は行わない。
--
-- 注意:
-- 既存の入力内容は失われ、復元できない。
--
-- 冪等性:
-- IF EXISTS のため再実行しても影響はない。
ALTER TABLE users DROP COLUMN IF EXISTS description;
ALTER TABLE users DROP COLUMN IF EXISTS favorite_anime;
ALTER TABLE users DROP COLUMN IF EXISTS sns_link;
