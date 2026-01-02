BEGIN;

DROP TRIGGER IF EXISTS publishers_touch_books_trg ON publishers;
DROP FUNCTION IF EXISTS publishers_touch_books();

DROP TRIGGER IF EXISTS authors_touch_books_trg ON authors;
DROP FUNCTION IF EXISTS authors_touch_books();

DROP TRIGGER IF EXISTS works_touch_books_trg ON works;
DROP FUNCTION IF EXISTS works_touch_books();

DROP TRIGGER IF EXISTS work_authors_touch_books_trg ON work_authors;
DROP FUNCTION IF EXISTS work_authors_touch_books();

DROP TRIGGER IF EXISTS book_works_touch_book_trg ON book_works;
DROP FUNCTION IF EXISTS book_works_touch_book();

DROP FUNCTION IF EXISTS touch_book(uuid);

DROP TABLE IF EXISTS work_authors;
DROP TABLE IF EXISTS book_works;
DROP TABLE IF EXISTS works;

ALTER TABLE books
    ADD COLUMN content jsonb NOT NULL DEFAULT '[]'::jsonb;

COMMIT;
