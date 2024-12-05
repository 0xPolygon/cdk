-- +migrate Down
DROP TABLE IF EXISTS block;
DROP TABLE IF EXISTS claim;
DROP TABLE IF EXISTS bridge;

-- +migrate Up
CREATE TABLE tracked_block (
	subscriber_id VARCHAR NOT NULL,
	num           BIGINT NOT NULL,
	hash          VARCHAR NOT NULL
);