-- +migrate Up
ALTER TABLE animes ADD COLUMN released_on VARCHAR(255);

-- +migrate Down
ALTER TABLE animes DROP COLUMN released_on;
