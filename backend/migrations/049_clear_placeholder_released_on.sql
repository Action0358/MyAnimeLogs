-- +migrate Up
-- Annictがreleased_on_aboutで放送時期未確定を示しているにもかかわらず、
-- released_onに"YYYY-01-01"のダミー日付が入ったまま同期されていたレコードを修正する。
--
-- 例: 〈物語〉シリーズ オフ&モンスターシーズン 業物語 かれんオウガ（season=2026-autumn,
-- released_on=2026-01-01）。季節が冬以外なのにreleased_onが1月1日になっているのは、
-- Annict側で年のみ判明・月日未定であることを示すパターン（アプリ側の再発防止は
-- services/annict.goで対応済み。本マイグレーションは既存の誤データのみを是正する）。
--
-- 冬アニメの1月1日公開は実在しうるため、season LIKE '%winter'は対象から除外する。
UPDATE animes
SET released_on = NULL
WHERE released_on ~ '^\d{4}-01-01$'
  AND season ~ '^\d{4}-(spring|summer|autumn)$'
  AND released_on = substring(season FROM '^\d{4}') || '-01-01';

-- +migrate Down
-- 元のダミー日付を復元する情報は失われているため、Downでは何もしない。
