-- +migrate Down

-- +migrate Up
CREATE TABLE IF NOT EXISTS proof (
	batch_num BIGINT NOT NULL,
	batch_num_final BIGINT NOT NULL,
	proof varchar NULL,
	proof_id varchar NULL,
	input_prover varchar NULL,
	prover varchar NULL,
	prover_id varchar NULL,
	created_at BIGINT NOT NULL,
	updated_at BIGINT NOT NULL,
	generating_since BIGINT,
    PRIMARY KEY (batch_num, batch_num_final)
);

CREATE TABLE  IF NOT EXISTS sequence (
    from_batch_num BIGINT NOT NULL,
    to_batch_num   BIGINT NOT NULL,
	PRIMARY KEY (from_batch_num)
);
