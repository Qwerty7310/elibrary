-- 002_rebuild_schema.up.sql
BEGIN;

DROP TRIGGER IF EXISTS books_search_vector_trigger ON books;
DROP TRIGGER IF EXISTS update_books_updated_at ON books;
DROP TRIGGER IF EXISTS update_barcode_sequences_updated_at ON barcode_sequences;

DROP FUNCTION IF EXISTS books_search_vector_update();
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS books_search_vector_gin;
DROP INDEX IF EXISTS books_barcode_idx;
DROP INDEX IF EXISTS books_factory_barcode_idx;
DROP INDEX IF EXISTS books_created_at_idx;
DROP INDEX IF EXISTS books_title_idx;
DROP INDEX IF EXISTS books_author_idx;

DROP TABLE IF EXISTS book_events CASCADE;
DROP TABLE IF EXISTS book_movements CASCADE;
DROP TABLE IF EXISTS book_loans CASCADE;

DROP TABLE IF EXISTS books CASCADE;
DROP TABLE IF EXISTS authors CASCADE;
DROP TABLE IF EXISTS publishers CASCADE;
DROP TABLE IF EXISTS locations CASCADE;
DROP TABLE IF EXISTS users CASCADE;

DROP TABLE IF EXISTS barcode_sequences CASCADE;

CREATE
OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at
= NOW();
RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TABLE users
(
    id          uuid PRIMARY KEY,
    login       text NOT NULL UNIQUE,
    first_name  text,
    last_name   text,
    middle_name text,
    email       text UNIQUE
);


CREATE TABLE authors
(
    id          uuid PRIMARY KEY,
    last_name   text        NOT NULL,
    first_name  text,
    middle_name text,
    birth_date  date,
    death_date  date,
    bio         text,
    photo_url   text,
    created_at  timestamptz NOT NULL DEFAULT NOW(),
    updated_at  timestamptz NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_authors_updated_at
    BEFORE UPDATE
    ON authors
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();


CREATE TABLE publishers
(
    id         uuid PRIMARY KEY,
    name       text        NOT NULL UNIQUE,
    logo_url   text,
    web_url    text,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_publishers_updated_at
    BEFORE UPDATE
    ON publishers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE locations
(
    id          uuid PRIMARY KEY,
    parent_id   uuid NULL REFERENCES locations(id) ON DELETE SET NULL,
    type        text        NOT NULL CHECK (type IN ('building', 'room', 'cabinet', 'shelf')),
    name        text        NOT NULL,

    barcode     text NULL,
    address     text,
    description text,

    created_at  timestamptz NOT NULL DEFAULT NOW(),
    updated_at  timestamptz NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX locations_barcode_uq
    ON locations (barcode) WHERE barcode IS NOT NULL;

CREATE INDEX locations_parent_id_idx ON locations (parent_id);
CREATE INDEX locations_type_idx ON locations (type);

CREATE TRIGGER update_locations_updated_at
    BEFORE UPDATE
    ON locations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE books
(
    id              uuid PRIMARY KEY,
    barcode         text        NOT NULL UNIQUE,
    factory_barcode text,

    title           text        NOT NULL,
    publisher_id    uuid NULL REFERENCES publishers(id) ON DELETE SET NULL,
    year            int,
    description     text,

    content         jsonb       NOT NULL DEFAULT '[]'::jsonb,

    location_id     uuid NULL REFERENCES locations(id) ON DELETE SET NULL,
    extra           jsonb       NOT NULL DEFAULT '{}'::jsonb,

    search_vector   tsvector,

    deleted_at      timestamptz NULL,
    deleted_by      uuid NULL REFERENCES users(id) ON DELETE SET NULL,

    created_at      timestamptz NOT NULL DEFAULT NOW(),
    updated_at      timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX books_barcode_idx ON books (barcode);
CREATE INDEX books_factory_barcode_idx ON books (factory_barcode);
CREATE INDEX books_created_at_idx ON books (created_at);
CREATE INDEX books_publisher_id_idx ON books (publisher_id);
CREATE INDEX books_location_id_idx ON books (location_id);


CREATE
OR REPLACE FUNCTION books_search_vector_update()
RETURNS trigger AS $$
DECLARE
pub_name    text := '';
  loc_name
text := '';
  loc_address
text := '';
  content_text
text := '';
  extra_text
text := '';
BEGIN
  IF
NEW.publisher_id IS NOT NULL THEN
SELECT p.name
INTO pub_name
FROM publishers p
WHERE p.id = NEW.publisher_id;
END IF;

  IF
NEW.location_id IS NOT NULL THEN
SELECT l.name, COALESCE(l.address, '')
INTO loc_name, loc_address
FROM locations l
WHERE l.id = NEW.location_id;
END IF;

  content_text
:= COALESCE(NEW.content::text, '');
  extra_text
:= COALESCE(NEW.extra::text, '');

  NEW.search_vector
:=
      setweight(to_tsvector('russian', COALESCE(NEW.title, '')), 'A')
    || setweight(to_tsvector('russian', COALESCE(NEW.description, '')), 'B')
    || setweight(to_tsvector('russian', COALESCE(pub_name, '')), 'C')
    || setweight(to_tsvector('russian', COALESCE(loc_name || ' ' || loc_address, '')), 'C')
    || setweight(to_tsvector('simple',  COALESCE(NEW.barcode, '')), 'B')
    || setweight(to_tsvector('simple',  COALESCE(NEW.factory_barcode, '')), 'B')
    || setweight(to_tsvector('russian', content_text), 'D')
    || setweight(to_tsvector('russian', extra_text), 'D');

RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER books_search_vector_trigger
    BEFORE INSERT OR
UPDATE ON books
    FOR EACH ROW
    EXECUTE FUNCTION books_search_vector_update();

CREATE TRIGGER update_books_updated_at
    BEFORE UPDATE
    ON books
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX books_search_vector_gin ON books USING GIN(search_vector);

CREATE TABLE book_loans
(
    id          uuid PRIMARY KEY,
    book_id     uuid        NOT NULL REFERENCES books (id) ON DELETE RESTRICT,
    user_id     uuid        NOT NULL REFERENCES users (id) ON DELETE RESTRICT,

    issued_at   timestamptz NOT NULL DEFAULT NOW(),
    due_at      timestamptz NULL,
    returned_at timestamptz NULL,

    issued_by   uuid NULL REFERENCES users(id) ON DELETE SET NULL,
    returned_by uuid NULL REFERENCES users(id) ON DELETE SET NULL,

    note        text
);

CREATE INDEX book_loans_book_id_idx ON book_loans (book_id);
CREATE INDEX book_loans_user_id_idx ON book_loans (user_id);
CREATE INDEX book_loans_issued_at_idx ON book_loans (issued_at);

CREATE UNIQUE INDEX book_loans_one_active_per_book
    ON book_loans (book_id) WHERE returned_at IS NULL;


CREATE TABLE book_movements
(
    id               uuid PRIMARY KEY,
    book_id          uuid        NOT NULL REFERENCES books (id) ON DELETE RESTRICT,

    from_location_id uuid NULL REFERENCES locations(id) ON DELETE SET NULL,
    to_location_id   uuid NULL REFERENCES locations(id) ON DELETE SET NULL,

    moved_at         timestamptz NOT NULL DEFAULT NOW(),
    actor_user_id    uuid NULL REFERENCES users(id) ON DELETE SET NULL,

    reason           text
);

CREATE INDEX book_movements_book_id_idx ON book_movements (book_id);
CREATE INDEX book_movements_moved_at_idx ON book_movements (moved_at);
CREATE INDEX book_movements_to_location_idx ON book_movements (to_location_id);


CREATE TABLE book_events
(
    id            uuid PRIMARY KEY,
    book_id       uuid NULL,
    event_type    text        NOT NULL CHECK (event_type IN ('created', 'updated', 'deleted', 'restored')),
    occurred_at   timestamptz NOT NULL DEFAULT NOW(),
    actor_user_id uuid NULL REFERENCES users(id) ON DELETE SET NULL,

    snapshot      jsonb       NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX book_events_book_id_idx ON book_events (book_id);
CREATE INDEX book_events_occurred_at_idx ON book_events (occurred_at);
CREATE INDEX book_events_type_idx ON book_events (event_type);

CREATE TABLE barcode_sequences
(
    prefix      int PRIMARY KEY CHECK (prefix BETWEEN 200 AND 299),
    last_value  bigint      NOT NULL DEFAULT 0,
    description text,
    created_at  timestamptz NOT NULL DEFAULT NOW(),
    updated_at  timestamptz NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_barcode_sequences_updated_at
    BEFORE UPDATE
    ON barcode_sequences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

INSERT INTO barcode_sequences (prefix, description)
VALUES (200, 'Основная библиотека') ON CONFLICT (prefix) DO NOTHING;

COMMIT;
