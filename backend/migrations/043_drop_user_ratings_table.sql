-- 評価(rating)機能の撤去に伴い、未使用となった user_ratings テーブルを削除する。
-- 関連: 005_create_user_ratings_table.sql（テーブル作成）
--
-- CASCADE により、user_ratings に紐づくインデックス
-- (idx_user_ratings_user_id / idx_user_ratings_anime_id / idx_user_ratings_rating) と
-- RLSポリシー(038で付与)も同時に削除される。
-- このテーブルを参照するVIEW・外部キーは存在しないため、他テーブルへの影響はない。
DROP TABLE IF EXISTS user_ratings CASCADE;
