-- +migrate Up
ALTER TABLE animes ADD COLUMN IF NOT EXISTS released_on VARCHAR(255);

-- +migrate Down
ALTER TABLE animes DROP COLUMN IF EXISTS released_on;
