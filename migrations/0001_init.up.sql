CREATE TABLE IF NOT EXISTS users (
    username    VARCHAR PRIMARY KEY,
    space       BIGINT NOT NULL
);
CREATE TABLE IF NOT EXISTS files (
    ownerid     VARCHAR NOT NULL,
    objkey      VARCHAR NOT NULL,
    filename    VARCHAR NOT NULL,
    id          VARCHAR NOT NULL,
    size        BIGINT NOT NULL,
    expiresat   TIMESTAMPTZ NOT NULL
);
