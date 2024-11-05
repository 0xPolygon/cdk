-- +migrate Down
DROP TABLE IF EXISTS block;
DROP TABLE IF EXISTS global_exit_root;

-- +migrate Up
CREATE TABLE block (
    num   BIGINT PRIMARY KEY
);

CREATE TABLE imported_global_exit_root (
	block_num           INTEGER PRIMARY KEY REFERENCES block(num) ON DELETE CASCADE,
	global_exit_root    VARCHAR NOT NULL,
	l1_info_tree_index  INTEGER NOT NULL
);