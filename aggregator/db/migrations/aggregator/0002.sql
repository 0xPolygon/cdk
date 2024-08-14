-- +migrate Up
DELETE FROM aggregator.batch;
ALTER TABLE aggregator.batch
    ADD COLUMN IF NOT EXISTS witness varchar NOT NULL;

-- +migrate Down
ALTER TABLE aggregator.batch
    DROP COLUMN IF EXISTS witness;
