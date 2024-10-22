-- +migrate Down
CREATE TABLE aggregator.batch (
    batch_num BIGINT NOT NULL,
    batch jsonb NOT NULL,
    datastream varchar NOT NULL,
    PRIMARY KEY (batch_num)
);

ALTER TABLE aggregator.proof
    ADD CONSTRAINT proof_batch_num_fkey FOREIGN KEY (batch_num) REFERENCES aggregator.batch (batch_num) ON DELETE CASCADE;

ALTER TABLE aggregator.sequence
    ADD CONSTRAINT sequence_from_batch_num_fkey FOREIGN KEY (from_batch_num) REFERENCES aggregator.batch (batch_num) ON DELETE CASCADE;


-- +migrate Up
ALTER TABLE aggregator.proof
    DROP CONSTRAINT IF EXISTS proof_batch_num_fkey;

ALTER TABLE aggregator.sequence
    DROP CONSTRAINT IF EXISTS sequence_from_batch_num_fkey;

DROP TABLE IF EXISTS aggregator.batch;
