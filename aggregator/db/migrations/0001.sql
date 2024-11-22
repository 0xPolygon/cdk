-- +migrate Down
DROP TABLE IF EXISTS proof;
DROP TABLE IF EXISTS sequence;

-- +migrate Up
CREATE TABLE IF NOT EXISTS proof (
	batch_num BIGINT NOT NULL,
	batch_num_final BIGINT NOT NULL,
	proof TEXT NULL,
	proof_id TEXT NULL,
	input_prover TEXT NULL,
	prover TEXT NULL,
	prover_id TEXT NULL,
	created_at BIGINT NOT NULL,
	updated_at BIGINT NOT NULL,
	generating_since BIGINT DEFAULT NULL,
    PRIMARY KEY (batch_num, batch_num_final)
);

CREATE TABLE  IF NOT EXISTS sequence (
    from_batch_num BIGINT NOT NULL,
    to_batch_num   BIGINT NOT NULL,
	PRIMARY KEY (from_batch_num)
);
