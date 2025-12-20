-- 1. Удаляем таблицу последовательностей (если создавали)
DROP TABLE IF EXISTS barcode_sequences;

-- 2. Удаляем триггеры
DROP TRIGGER IF EXISTS update_books_updated_at ON books;
DROP TRIGGER IF EXISTS books_search_vector_trigger ON books;
DROP FUNCTION IF EXISTS update_books_updated_at();

-- 3. Удаляем индексы
DROP INDEX IF EXISTS books_ean13_idx;
DROP INDEX IF EXISTS books_created_at_idx;

-- 4. Восстанавливаем старые названия столбцов
ALTER TABLE books RENAME COLUMN uuid TO id;
ALTER TABLE books RENAME COLUMN ean13 TO barcode;

-- 5. Удаляем timestamps
ALTER TABLE books
DROP COLUMN IF EXISTS created_at,
DROP COLUMN IF EXISTS updated_at;

-- 6. Восстанавливаем constraints
ALTER TABLE books
DROP CONSTRAINT books_pkey,
ADD PRIMARY KEY (id);

ALTER TABLE books
DROP CONSTRAINT books_ean13_key,
ADD CONSTRAINT books_barcode_key UNIQUE (barcode);

-- 7. Восстанавливаем старый триггер поиска
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

-- 8. Восстанавливаем индекс factory_barcode
CREATE INDEX IF NOT EXISTS books_factory_barcode_idx ON books(factory_barcode);