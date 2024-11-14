-- +migrate Down
DROP SCHEMA IF EXISTS aggregator CASCADE;

-- +migrate Up
CREATE SCHEMA aggregator;

CREATE TABLE IF NOT EXISTS aggregator.proof (
	batch_num BIGINT NOT NULL,
	batch_num_final BIGINT NOT NULL,
	proof varchar NULL,
	proof_id varchar NULL,
	input_prover varchar NULL,
	prover varchar NULL,
	prover_id varchar NULL,
	created_at TIMESTAMP,
	updated_at TIMESTAMP,
	generating_since TIMESTAMP,
    PRIMARY KEY (batch_num, batch_num_final)
);

CREATE TABLE  IF NOT EXISTS aggregator.sequence (
    from_batch_num BIGINT NOT NULL,
    to_batch_num   BIGINT NOT NULL,
	PRIMARY KEY (from_batch_num)
);
