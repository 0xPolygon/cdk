-- +migrate Down
DROP TABLE IF EXISTS certificate_info;
DROP TABLE IF EXISTS certificate_info_history;
DROP TABLE IF EXISTS certificate_info_history;

-- +migrate Up
CREATE TABLE certificate_info (
    height                      INTEGER NOT NULL,
    retry_count                 INTEGER DEFAULT 0, 
    certificate_id              VARCHAR NOT NULL,
    status                      INTEGER NOT NULL,
    previous_local_exit_root    VARCHAR,
    new_local_exit_root         VARCHAR NOT NULL,
    from_block                  INTEGER NOT NULL,
    to_block                    INTEGER NOT NULL,
    created_at                  INTEGER NOT NULL,
    updated_at                  INTEGER NOT NULL,
    signed_certificate          TEXT,
    PRIMARY KEY (height)
);

CREATE TABLE certificate_info_history (
    height                      INTEGER NOT NULL ,
    retry_count                 INTEGER DEFAULT 0,  
    certificate_id              VARCHAR NOT NULL,
    status                      INTEGER NOT NULL,
    previous_local_exit_root    VARCHAR,
    new_local_exit_root         VARCHAR NOT NULL,
    from_block                  INTEGER NOT NULL,
    to_block                    INTEGER NOT NULL,
    created_at                  INTEGER NOT NULL,
    updated_at                  INTEGER NOT NULL,
    signed_certificate          TEXT,
    PRIMARY KEY (height, retry_count)
);
