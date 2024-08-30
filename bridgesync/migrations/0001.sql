-- +migrate Down
DROP SCHEMA IF EXISTS bridge CASCADE;

-- +migrate Up
CREATE SCHEMA bridge;

CREATE TABLE bridge.index
(
    index_num   BIGINT PRIMARY KEY,
);

CREATE TABLE bridge.claim
(
    index_num           INTEGER NOT NULL REFERENCES bridge.index (index_num) ON DELETE CASCADE,
    index_pos           INTEGER NOT NULL,

    leaf_type           INTEGER NOT NULL,
	origin_network      INTEGER NOT NULL,
	origin_address      VARCHAR NOT NULL,
	destination_address INTEGER NOT NULL,
	destination_address VARCHAR NOT NULL,
	amount              DECIMAL(78, 0) NOT NULL,
	metadata            BLOB,
	deposit_count       INTEGER NOT NULL,

    PRIMARY KEY (index_num, index_pos)
);

CREATE TABLE bridge.bridge
(
    index_num               BIGINT NOT NULL REFERENCES bridge.index (index_num) ON DELETE CASCADE,
    index_pos               BIGINT NOT NULL,

    global_index            DECIMAL(78, 0) NOT NULL,
	origin_network          INTEGER NOT NULL,
	origin_address          VARCHAR NOT NULL,
	destination_address     VARCHAR NOT NULL,
	amount                  DECIMAL(78, 0) NOT NULL,
	proof_local_exit_root   VARCHAR,
	proof_rollup_exit_root  VARCHAR,
	mainnet_exit_root       VARCHAR,
	rollup_exit_root        VARCHAR,
	global_exit_root        VARCHAR,
	destination_network     INTEGER NOT NULL,
	metadata                BLOB,
	is_message              BOOLEAN,

    PRIMARY KEY (index_num, index_pos)
);