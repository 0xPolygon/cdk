-- +migrate Down
DROP TABLE IF EXISTS /*dbprefix*/root;
DROP TABLE IF EXISTS /*dbprefix*/rht;

-- +migrate Up
CREATE TABLE /*dbprefix*/root (
    hash        VARCHAR PRIMARY KEY,
    position    INTEGER NOT NULL,
    block_num   BIGINT NOT NULL,
    block_position   BIGINT NOT NULL
);

CREATE TABLE /*dbprefix*/rht (
    hash    VARCHAR PRIMARY KEY,
    left    VARCHAR NOT NULL,
    right   VARCHAR NOT NULL
);
