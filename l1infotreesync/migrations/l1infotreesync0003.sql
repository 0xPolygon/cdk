-- +migrate Down
ALTER TABLE block DROP COLUMN hash;

-- +migrate Up
ALTER TABLE block ADD COLUMN hash VARCHAR;