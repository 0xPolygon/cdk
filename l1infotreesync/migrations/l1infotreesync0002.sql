-- +migrate Down
DROP TABLE IF EXISTS l1info_initial;

-- +migrate Up

CREATE TABLE l1info_initial (
    -- single_row_id prevent to have more than 1 row in this table
    single_row_id       INTEGER check(single_row_id=1) NOT NULL  DEFAULT 1,
    block_num           INTEGER NOT NULL REFERENCES block(num) ON DELETE CASCADE,
    leaf_count          INTEGER NOT NULL,
    l1_info_root        VARCHAR NOT NULL,
    PRIMARY KEY (single_row_id)
);

