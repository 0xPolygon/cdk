-- +migrate Down
DROP SCHEMA IF EXISTS aggregator CASCADE;

-- +migrate Up
CREATE SCHEMA aggregator;

CREATE TABLE IF NOT EXISTS aggregator.batch (
	batch_num BIGINT NOT NULL,
	batch jsonb NOT NULL,
	datastream varchar NOT NULL,
	PRIMARY KEY (batch_num)
);

CREATE TABLE IF NOT EXISTS aggregator.proof (
	batch_num BIGINT NOT NULL REFERENCES aggregator.batch (batch_num) ON DELETE CASCADE,
	batch_num_final BIGINT NOT NULL,
	proof varchar NULL,
	proof_id varchar NULL,
	input_prover varchar NULL,
	prover varchar NULL,
	prover_id varchar NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	generating_since timestamptz NULL,
    PRIMARY KEY (batch_num, batch_num_final)
);

CREATE TABLE  IF NOT EXISTS aggregator.sequence (
    from_batch_num BIGINT NOT NULL REFERENCES aggregator.batch (batch_num) ON DELETE CASCADE,
    to_batch_num   BIGINT NOT NULL,
	PRIMARY KEY (from_batch_num)
);
