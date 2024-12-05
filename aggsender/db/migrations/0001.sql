-- +migrate Down
DROP TABLE IF EXISTS certificate_info;

-- +migrate Up
CREATE TABLE certificate_info (
    height                      INTEGER NOT NULL,
    certificate_id              VARCHAR NOT NULL PRIMARY KEY,
    status                      INTEGER NOT NULL,
    new_local_exit_root         VARCHAR NOT NULL,
    from_block                  INTEGER NOT NULL,
    to_block                    INTEGER NOT NULL,
    created_at                  INTEGER NOT NULL,
    updated_at                  INTEGER NOT NULL,
    signed_certificate          TEXT
);