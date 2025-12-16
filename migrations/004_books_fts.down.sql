DROP INDEX IF EXISTS books_search_vector_gin;
DROP TRIGGER IF EXISTS books_search_vector_trigger ON books;
DROP FUNCTION IF EXISTS books_search_vector_update();
ALTER TABLE books DROP COLUMN IF EXISTS search_vector;
