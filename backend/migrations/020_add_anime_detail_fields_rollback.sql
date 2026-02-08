-- Rollback: Remove description and released_on fields

ALTER TABLE animes DROP COLUMN IF EXISTS description;
ALTER TABLE animes DROP COLUMN IF EXISTS released_on;
