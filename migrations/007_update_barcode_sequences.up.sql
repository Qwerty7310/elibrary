-- 007_update_barcode_sequences.sql

BEGIN;

DROP TABLE IF EXISTS barcode_sequences CASCADE;

CREATE TABLE barcode_sequences
(
    type        text PRIMARY KEY,
    prefix      integer     NOT NULL UNIQUE,
    last_value  bigint      NOT NULL DEFAULT 0,
    description text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT barcode_sequences_prefix_check
        CHECK (
            (type = 'book' AND prefix BETWEEN 200 AND 299) OR
            (type = 'location' AND prefix BETWEEN 300 AND 399)
            )
);

CREATE TRIGGER update_barcode_sequences_updated_at
    BEFORE UPDATE
    ON barcode_sequences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

INSERT INTO barcode_sequences (type, prefix, description)
VALUES ('book', 200, 'Books barcode prefix'),
       ('location', 300, 'Locations barcode prefix');

COMMIT;
