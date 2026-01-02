BEGIN;

ALTER TABLE books
DROP COLUMN content;


CREATE TABLE works (
                       id          uuid PRIMARY KEY,
                       title       text NOT NULL,
                       description text,
                       year        int,

                       created_at  timestamptz NOT NULL DEFAULT now(),
                       updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER update_works_updated_at
    BEFORE UPDATE ON works
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();


CREATE TABLE book_works (
                            book_id uuid NOT NULL REFERENCES books(id) ON DELETE CASCADE,
                            work_id uuid NOT NULL REFERENCES works(id) ON DELETE RESTRICT,

                            position int,
                            note     text,

                            PRIMARY KEY (book_id, work_id)
);

CREATE INDEX book_works_book_id_idx ON book_works (book_id);
CREATE INDEX book_works_work_id_idx ON book_works (work_id);


CREATE TABLE work_authors (
                              work_id   uuid NOT NULL REFERENCES works(id) ON DELETE CASCADE,
                              author_id uuid NOT NULL REFERENCES authors(id) ON DELETE RESTRICT,

                              PRIMARY KEY (work_id, author_id)
);

CREATE INDEX work_authors_work_id_idx   ON work_authors (work_id);
CREATE INDEX work_authors_author_id_idx ON work_authors (author_id);


CREATE OR REPLACE FUNCTION books_search_vector_update()
RETURNS trigger AS $$
DECLARE
pub_name     text := '';
    works_text   text := '';
    authors_text text := '';
BEGIN
    -- издательство
    IF NEW.publisher_id IS NOT NULL THEN
SELECT p.name
INTO pub_name
FROM publishers p
WHERE p.id = NEW.publisher_id;
END IF;

    -- произведения книги
SELECT COALESCE(string_agg(w.title, ' ' ORDER BY COALESCE(bw.position, 2147483647)), '')
INTO works_text
FROM book_works bw
         JOIN works w ON w.id = bw.work_id
WHERE bw.book_id = NEW.id;

-- авторы произведений книги
SELECT COALESCE(string_agg(
                        concat_ws(' ', a.last_name, a.first_name, a.middle_name),
                        ' '
                ), '')
INTO authors_text
FROM book_works bw
         JOIN work_authors wa ON wa.work_id = bw.work_id
         JOIN authors a       ON a.id = wa.author_id
WHERE bw.book_id = NEW.id;

NEW.search_vector :=
          setweight(to_tsvector('russian', works_text),   'A')
        || setweight(to_tsvector('russian', authors_text), 'A')
        || setweight(to_tsvector('russian', pub_name),     'B')
        || setweight(to_tsvector('simple',  COALESCE(NEW.barcode, '')),         'C')
        || setweight(to_tsvector('simple',  COALESCE(NEW.factory_barcode, '')), 'C');

RETURN NEW;
END;
$$ LANGUAGE plpgsql;


DROP TRIGGER IF EXISTS books_search_vector_trigger ON books;

CREATE TRIGGER books_search_vector_trigger
    BEFORE INSERT OR UPDATE ON books
                         FOR EACH ROW
                         EXECUTE FUNCTION books_search_vector_update();



CREATE OR REPLACE FUNCTION touch_book(p_book_id uuid)
RETURNS void AS $$
BEGIN
UPDATE books
SET updated_at = now()
WHERE id = p_book_id;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION book_works_touch_book()
RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        PERFORM touch_book(NEW.book_id);
ELSE
        PERFORM touch_book(OLD.book_id);
END IF;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER book_works_touch_book_trg
    AFTER INSERT OR DELETE ON book_works
    FOR EACH ROW
    EXECUTE FUNCTION book_works_touch_book();


CREATE OR REPLACE FUNCTION work_authors_touch_books()
RETURNS trigger AS $$
DECLARE
v_work_id uuid;
BEGIN
    v_work_id := CASE WHEN TG_OP = 'INSERT' THEN NEW.work_id ELSE OLD.work_id END;

UPDATE books
SET updated_at = now()
WHERE id IN (
    SELECT bw.book_id FROM book_works bw WHERE bw.work_id = v_work_id
);

RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER work_authors_touch_books_trg
    AFTER INSERT OR DELETE ON work_authors
    FOR EACH ROW
    EXECUTE FUNCTION work_authors_touch_books();


CREATE OR REPLACE FUNCTION works_touch_books()
RETURNS trigger AS $$
BEGIN
UPDATE books
SET updated_at = now()
WHERE id IN (
    SELECT bw.book_id FROM book_works bw WHERE bw.work_id = NEW.id
);
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER works_touch_books_trg
    AFTER UPDATE OF title, description, year ON works
    FOR EACH ROW
    EXECUTE FUNCTION works_touch_books();


CREATE OR REPLACE FUNCTION authors_touch_books()
RETURNS trigger AS $$
BEGIN
UPDATE books
SET updated_at = now()
WHERE id IN (
    SELECT bw.book_id
    FROM book_works bw
             JOIN work_authors wa ON wa.work_id = bw.work_id
    WHERE wa.author_id = NEW.id
);
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER authors_touch_books_trg
    AFTER UPDATE OF last_name, first_name, middle_name, bio ON authors
    FOR EACH ROW
    EXECUTE FUNCTION authors_touch_books();


CREATE OR REPLACE FUNCTION publishers_touch_books()
RETURNS trigger AS $$
BEGIN
UPDATE books
SET updated_at = now()
WHERE publisher_id = NEW.id;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER publishers_touch_books_trg
    AFTER UPDATE OF name ON publishers
    FOR EACH ROW
    EXECUTE FUNCTION publishers_touch_books();


COMMIT;
