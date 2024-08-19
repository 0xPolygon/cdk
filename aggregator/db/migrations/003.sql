-- +migrate Up
ALTER TABLE aggregator.batch
    ALTER COLUMN witness DROP NOT NULL;

-- +migrate Down
ALTER TABLE aggregator.batch
    ALTER COLUMN witness SET NOT NULL;
