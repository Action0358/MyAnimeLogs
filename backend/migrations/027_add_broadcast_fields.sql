-- Add broadcast information fields from Jikan API
-- broadcast_day: 放送曜日（例: "Fridays"）
-- broadcast_time: 放送時間（例: "23:30"）
-- broadcast_timezone: タイムゾーン（例: "Asia/Tokyo"）

ALTER TABLE animes
ADD COLUMN IF NOT EXISTS broadcast_day VARCHAR(20),
ADD COLUMN IF NOT EXISTS broadcast_time VARCHAR(10),
ADD COLUMN IF NOT EXISTS broadcast_timezone VARCHAR(50);
