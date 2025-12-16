ALTER TABLE books
    ADD COLUMN search_vector tsvector;

UPDATE books
SET search_vector =
        to_tsvector(
                'russian',
                coalesce(title, '') || ' ' ||
                coalesce(author, '') || ' ' ||
                coalesce(publisher, '') || ' ' ||
                coalesce(extra::text, '')
        );

CREATE INDEX books_search_idx
    ON books
    USING GIN(search_vector);