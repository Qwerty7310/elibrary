-- 1) колонка tsvector (если ещё нет)
ALTER TABLE books
    ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- 2) функция обновления search_vector
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

-- 3) триггер (на insert/update)
DROP TRIGGER IF EXISTS books_search_vector_trigger ON books;

CREATE TRIGGER books_search_vector_trigger
    BEFORE INSERT OR UPDATE ON books
                         FOR EACH ROW
                         EXECUTE FUNCTION books_search_vector_update();

-- 4) индекс для быстрого поиска
CREATE INDEX IF NOT EXISTS books_search_vector_gin
    ON books
    USING GIN (search_vector);

-- 5) заполнить search_vector для уже существующих строк
UPDATE books SET title = title;
