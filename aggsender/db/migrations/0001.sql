-- +migrate Down
DROP TABLE IF EXISTS certificate_info;

CREATE TABLE certificate_info (
	height                      INTEGER NOT NULL,
	certificate_id              VARCHAR NOT NULL,
    status                      INTEGER NOT NULL,
    new_local_exit_root         VARCHAR NOT NULL,
    from_block                  INTEGER NOT NULL,
    to_block                    INTEGER NOT NULL,

    PRIMARY KEY (height, certificate_id)
);