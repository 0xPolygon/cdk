-- +migrate Down
DROP TABLE IF EXISTS block;
DROP TABLE IF EXISTS claim;
DROP TABLE IF EXISTS bridge;

-- +migrate Up
CREATE TABLE block (
    num   BIGINT PRIMARY KEY
);

CREATE TABLE bridge (
	block_num           INTEGER NOT NULL REFERENCES block(num) ON DELETE CASCADE,
	block_pos           INTEGER NOT NULL,
	leaf_type           INTEGER NOT NULL,
	origin_network      INTEGER NOT NULL,
	origin_address      VARCHAR NOT NULL,
	destination_network INTEGER NOT NULL,
	destination_address VARCHAR NOT NULL,
	amount              TEXT NOT NULL,
	metadata            BLOB,
	deposit_count       INTEGER NOT NULL,
	PRIMARY KEY (block_num, block_pos)
);

CREATE TABLE claim (
	block_num               INTEGER NOT NULL REFERENCES block(num) ON DELETE CASCADE,
	block_pos               INTEGER NOT NULL,
	global_index           	TEXT NOT NULL,
	origin_network          INTEGER NOT NULL,
	origin_address          VARCHAR NOT NULL,
	destination_address     VARCHAR NOT NULL,
	amount                  TEXT NOT NULL,
	proof_local_exit_root   VARCHAR,
	proof_rollup_exit_root  VARCHAR,
	mainnet_exit_root       VARCHAR,
	rollup_exit_root        VARCHAR,
	global_exit_root        VARCHAR,
	destination_network     INTEGER NOT NULL,
	metadata                BLOB,
	is_message              BOOLEAN,
	PRIMARY KEY (block_num, block_pos)
);