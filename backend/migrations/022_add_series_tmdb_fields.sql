-- +migrate Up
ALTER TABLE animes ADD COLUMN series_tmdb_id INTEGER;

-- +migrate Down
ALTER TABLE animes DROP COLUMN series_tmdb_id;
