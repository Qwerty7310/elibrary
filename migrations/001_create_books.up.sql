CREATE
EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE books
(
    id        UUID PRIMARY KEY,
    title     TEXT  NOT NULL,
    author    TEXT  NOT NULL,
    publisher TEXT,
    year      INT,
    location  TEXT,
    extra     JSONB NOT NULL DEFAULT '{}'
);