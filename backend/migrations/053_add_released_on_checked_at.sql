-- +migrate Up
-- 放送開始日（released_on）が空の作品について、外部API（Annictの放送枠・AniList）へ
-- 補完を問い合わせた日時を記録する。
--
-- 今期・前期以外の作品は日次同期でreleased_onを補っておらず、過去シーズンの作品の
-- 約6割が空のままだった（例: 新劇場版 銀魂 -吉原大炎上-。AniListには公開日がある）。
-- 全シーズンを補完対象にするにあたり、どの情報源にも日付が無い作品へ毎日問い合わせ
-- 直さないよう、問い合わせ済みの日時を持たせる。NULLは未確認（補完対象）。
ALTER TABLE animes ADD COLUMN IF NOT EXISTS released_on_checked_at TIMESTAMPTZ;

-- +migrate Down
ALTER TABLE animes DROP COLUMN IF EXISTS released_on_checked_at;
