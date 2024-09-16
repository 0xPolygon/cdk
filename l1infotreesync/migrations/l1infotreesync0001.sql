-- +migrate Down
DROP TABLE IF EXISTS block;
DROP TABLE IF EXISTS claim;
DROP TABLE IF EXISTS bridge;

-- +migrate Up
CREATE TABLE block (
    num   BIGINT PRIMARY KEY
);

CREATE TABLE l1info_leaf (
    block_num           INTEGER NOT NULL REFERENCES block(num) ON DELETE CASCADE,
    block_pos           INTEGER NOT NULL,
    position            INTEGER NOT NULL,
    previous_block_hash VARCHAR NOT NULL,
    timestamp           INTEGER NOT NULL,
    mainnet_exit_root   VARCHAR NOT NULL,
    rollup_exit_root    VARCHAR NOT NULL,
    global_exit_root    VARCHAR NOT NULL,
    hash                VARCHAR NOT NULL,
    PRIMARY KEY (block_num, block_pos)
);
