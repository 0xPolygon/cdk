-- +migrate Down
DROP TABLE IF EXISTS block;
DROP TABLE IF EXISTS claim;
DROP TABLE IF EXISTS bridge;

-- +migrate Up
CREATE TABLE block (
    num   BIGINT PRIMARY KEY
);

CREATE TABLE injected_ger (
	block_num           INTEGER NOT NULL REFERENCES block(num) ON DELETE CASCADE,
	block_pos           INTEGER NOT NULL,
	l1_info_tree_index  INTEGER NOT NULL,
	global_exit_root    VARCHAR NOT NULL,
	PRIMARY KEY (block_num, block_pos)
);
