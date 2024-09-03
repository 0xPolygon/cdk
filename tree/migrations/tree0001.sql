-- +migrate Down
DROP TABLE IF EXISTS root;
DROP TABLE IF EXISTS rht;

-- +migrate Up
CREATE TABLE root (
    hash        VARCHAR PRIMARY KEY,
    position    INTEGER NOT NULL UNIQUE,
    block_num   BIGINT NOT NULL,
    block_position   BIGINT NOT NULL
);

CREATE TABLE rht (
    hash    VARCHAR PRIMARY KEY,
    left    VARCHAR NOT NULL,
    right   VARCHAR NOT NULL
);
