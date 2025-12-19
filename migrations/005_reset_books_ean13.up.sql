-- Внимание: удаляет все данные
DROP TRIGGER IF EXISTS books_search_vector_trigger ON books;
DROP FUNCTION IF EXISTS books_search_vector_update();
DROP INDEX IF EXISTS books_search_vector_gin;

DROP TABLE IF EXISTS books;

CREATE TABLE books (
                       id TEXT PRIMARY KEY,              -- EAN-13 как строка
                       barcode TEXT NOT NULL UNIQUE,     -- генерируемый штрихкод (обычно совпадает с id)
                       factory_barcode TEXT,             -- заводской, может быть NULL и не уникален
                       title TEXT NOT NULL,
                       author TEXT NOT NULL,
                       publisher TEXT,
                       year INT,
                       location TEXT,
                       extra JSONB NOT NULL DEFAULT '{}',
                       search_vector tsvector
);

CREATE OR REPLACE FUNCTION books_search_vector_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('russian', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('russian', coalesce(NEW.author, '')), 'B') ||
    setweight(to_tsvector('russian', coalesce(NEW.publisher, '')), 'C') ||
    setweight(to_tsvector('russian', coalesce(NEW.location, '')), 'D');
RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER books_search_vector_trigger
    BEFORE INSERT OR UPDATE ON books
                         FOR EACH ROW EXECUTE FUNCTION books_search_vector_update();

CREATE INDEX books_search_vector_gin ON books USING GIN (search_vector);
CREATE INDEX books_factory_barcode_idx ON books(factory_barcode);
