-- 001_initial_schema.down.sql
BEGIN;
DROP TRIGGER IF EXISTS update_books_updated_at ON books;
DROP TRIGGER IF EXISTS update_barcode_sequences_updated_at ON barcode_sequences;
DROP TRIGGER IF EXISTS books_search_vector_trigger ON books;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS books_search_vector_update();
DROP TABLE IF EXISTS barcode_sequences;
DROP TABLE IF EXISTS books;
COMMIT;