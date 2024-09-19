-- +migrate Down
DROP TABLE IF EXISTS initial_info;

-- +migrate Up

CREATE TABLE initial_info (
    block_num           INTEGER NOT NULL REFERENCES block(num) ON DELETE CASCADE,
    leaf_count          INTEGER NOT NULL,
    l1_info_root        VARCHAR NOT NULL,
)

