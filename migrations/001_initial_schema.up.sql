-- 001_initial_schema.up.sql
BEGIN;

-- 1. Таблица книг
CREATE TABLE books (
                       id UUID PRIMARY KEY,
                       barcode TEXT NOT NULL UNIQUE,
                       factory_barcode TEXT,
                       title TEXT NOT NULL,
                       author TEXT NOT NULL,
                       publisher TEXT,
                       year INT,
                       location TEXT,
                       extra JSONB NOT NULL DEFAULT '{}',
                       search_vector TSVECTOR,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. УНИВЕРСАЛЬНАЯ функция для полнотекстового поиска ПО ВСЕМ полям
CREATE OR REPLACE FUNCTION books_search_vector_update() RETURNS trigger AS $$
BEGIN
  -- Преобразуем extra JSON в текст для поиска
  DECLARE
extra_text TEXT := '';
    extra_key TEXT;
    extra_value TEXT;
BEGIN
    -- Конвертируем все значения из extra в одну строку
    IF NEW.extra IS NOT NULL THEN
      FOR extra_key, extra_value IN SELECT * FROM jsonb_each_text(NEW.extra)
                                                             LOOP
    extra_text := extra_text || ' ' || extra_key || ' ' || extra_value;
END LOOP;
END IF;

    -- Ищем по ВСЕМ полям с разными весами
    NEW.search_vector :=
      -- Важные поля (высокий вес)
      setweight(to_tsvector('russian', coalesce(NEW.title, '')), 'A') ||
      setweight(to_tsvector('russian', coalesce(NEW.author, '')), 'A') ||

      -- Идентификаторы (средний вес)
      setweight(to_tsvector('simple', coalesce(NEW.barcode, '')), 'B') ||
      setweight(to_tsvector('simple', coalesce(NEW.factory_barcode, '')), 'B') ||

      -- Дополнительные поля (низкий вес)
      setweight(to_tsvector('russian', coalesce(NEW.publisher, '')), 'C') ||
      setweight(to_tsvector('russian', coalesce(NEW.location, '')), 'C') ||

      -- Extra JSON (низкий вес)
      setweight(to_tsvector('russian', extra_text), 'D');
END;

RETURN NEW;
END
$$ LANGUAGE plpgsql;

-- 3. Триггер для поиска
CREATE TRIGGER books_search_vector_trigger
    BEFORE INSERT OR UPDATE ON books
                         FOR EACH ROW
                         EXECUTE FUNCTION books_search_vector_update();

-- 4. Функция для обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 5. Триггер для updated_at в books
CREATE TRIGGER update_books_updated_at
    BEFORE UPDATE ON books
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 6. Индексы
CREATE INDEX books_barcode_idx ON books(barcode);
CREATE INDEX books_factory_barcode_idx ON books(factory_barcode);
CREATE INDEX books_search_vector_gin ON books USING GIN(search_vector);
CREATE INDEX books_created_at_idx ON books(created_at);

-- 7. Дополнительный индекс для полнотекстового поиска по отдельным полям
CREATE INDEX books_title_idx ON books USING GIN(to_tsvector('russian', title));
CREATE INDEX books_author_idx ON books USING GIN(to_tsvector('russian', author));

-- 8. Таблица для последовательностей штрихкодов
CREATE TABLE barcode_sequences (
                                   prefix INT PRIMARY KEY CHECK (prefix BETWEEN 200 AND 299),
                                   last_value BIGINT NOT NULL DEFAULT 0,
                                   description TEXT,
                                   created_at TIMESTAMPTZ DEFAULT NOW(),
                                   updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 9. Триггер для barcode_sequences
CREATE TRIGGER update_barcode_sequences_updated_at
    BEFORE UPDATE ON barcode_sequences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 10. Инициализация префикса
INSERT INTO barcode_sequences (prefix, description)
VALUES (200, 'Основная библиотека')
    ON CONFLICT (prefix) DO NOTHING;

COMMIT;