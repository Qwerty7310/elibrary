-- 1. Удаляем старый триггер
DROP TRIGGER IF EXISTS books_search_vector_trigger ON books;

-- 2. Переименовываем id в uuid
ALTER TABLE books RENAME COLUMN id TO uuid;

-- 3. Переименовываем barcode в ean13 (наш штрихкод)
ALTER TABLE books RENAME COLUMN barcode TO ean13;

-- 4. Добавляем timestamps если их нет
ALTER TABLE books
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

-- 5. Обновляем constraints
ALTER TABLE books
DROP CONSTRAINT books_pkey,
ADD PRIMARY KEY (uuid);

ALTER TABLE books
DROP CONSTRAINT books_barcode_key,
ADD CONSTRAINT books_ean13_key UNIQUE (ean13);

-- 6. Обновляем комментарии
COMMENT ON COLUMN books.uuid IS 'Уникальный идентификатор книги (UUID)';
COMMENT ON COLUMN books.ean13 IS 'Библиотечный штрихкод EAN-13 (уникальный)';
COMMENT ON COLUMN books.factory_barcode IS 'Заводской штрихкод/ISBN (может повторяться)';

-- 7. Обновляем триггер для полнотекстового поиска
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

-- 8. Триггер для updated_at
CREATE OR REPLACE FUNCTION update_books_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_books_updated_at
    BEFORE UPDATE ON books
    FOR EACH ROW
    EXECUTE FUNCTION update_books_updated_at();

-- 9. Обновляем индексы
DROP INDEX IF EXISTS books_factory_barcode_idx;
CREATE INDEX books_ean13_idx ON books(ean13);
CREATE INDEX books_factory_barcode_idx ON books(factory_barcode);
CREATE INDEX books_created_at_idx ON books(created_at);

-- 10. (Опционально) Создаем таблицу для последовательностей штрихкодов
CREATE TABLE IF NOT EXISTS barcode_sequences (
                                                 prefix INT PRIMARY KEY CHECK (prefix BETWEEN 200 AND 299),
    last_value BIGINT NOT NULL DEFAULT 0,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
    );

CREATE TRIGGER update_barcode_sequences_updated_at
    BEFORE UPDATE ON barcode_sequences
    FOR EACH ROW
    EXECUTE FUNCTION update_books_updated_at;

-- Инициализируем дефолтный префикс
INSERT INTO barcode_sequences (prefix, description)
VALUES (200, 'Основная библиотека')
    ON CONFLICT (prefix) DO NOTHING;