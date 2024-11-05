-- +migrate Down
DROP TABLE IF EXISTS claim;

-- +migrate Up
CREATE TABLE claim (
	leaf_type              INT NOT NULL,
	proof_local_exit_root  VARCHAR NOT NULL,
	proof_rollup_exit_root VARCHAR NOT NULL, 
	global_index           VARCHAR NOT NULL,
	mainnet_exit_root      VARCHAR NOT NULL,
	rollup_exit_root       VARCHAR NOT NULL,
	origin_network         INT NOT NULL,
	origin_token_address   VARCHAR NOT NULL,
	destination_network    INT NOT NULL,
	destination_address    VARCHAR NOT NULL,
	amount                 VARCHAR NOT NULL,
	metadata               VARCHAR,
	status                 VARCHAR NOT NULL,
	tx_id                  VARCHAR NOT NULL
);