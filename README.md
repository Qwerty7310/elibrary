# elibrary

E-library — сервис каталога и учета библиотечного фонда. Предоставляет REST API для управления книгами, произведениями, авторами, издателями, локациями и пользователями, а также JWT-аутентификацию и ролевую модель доступа. Система генерирует штрих-коды EAN-13 и готовит задания на печать штрих-кодов через интеграцию с очередью сообщений.

## Возможности

- Публичные и внутренние API для книг
- Админский CRUD для произведений, авторов, издателей, локаций и пользователей
- JWT-аутентификация и RBAC (admin)
- Генерация и валидация EAN-13
- Миграции PostgreSQL и Docker-окружение

## Стек

- Backend: Go, chi, pgx
- Database: PostgreSQL
- Frontend: React, TypeScript, Vite
- Infrastructure: Docker, docker-compose, migrate

## Архитектура

```
cmd/server          # Точка входа HTTP-сервера
internal/domain     # Сущности и ошибки домена
internal/service    # Бизнес-логика
internal/repository # Слой доступа к данным (PostgreSQL)
internal/http       # Роутинг, хендлеры, middleware
internal/readmodel  # Read-модели для API
migrations          # SQL-миграции
client              # Frontend на React
```

## Запуск локально (Docker)

```
docker compose up --build
```

Сервисы:
- backend: http://localhost:8080
- postgres: доступен внутри Docker-сети

## Конфигурация

Переменные окружения (backend):
- `DB_URL` (обязательная): строка подключения PostgreSQL
- `HTTP_ADDR` (опционально): адрес биндинга, по умолчанию `:8080`
- `JWT_SECRET` (обязательная): секрет для JWT
- `CORS_ALLOWED_ORIGINS` (опционально): список разрешенных origin через запятую

Пример:

```
DB_URL=postgres://elibrary:elibrary@localhost:5432/elibrary?sslmode=disable
HTTP_ADDR=:8080
JWT_SECRET=change-me
CORS_ALLOWED_ORIGINS=http://localhost:5173
```

## Обзор API

Публичные:
- `POST /auth/login`
- `GET /books/public`
- `GET /books/public/{id}`
- `GET /health`

Защищенные (JWT):
- `GET /books/internal`
- `GET /books/internal/{id}`
- `GET /authors/{id}`
- `GET /works/{id}`
- `GET /publishers/{id}`
- `GET /locations/type/{type}`
- `GET /locations/{id}`
- `GET /locations/child/{id}/{type}`

Админские (JWT + роль admin):
- `POST /admin/books`
- `PUT /admin/books/{id}`
- `POST /admin/works`, `PUT /admin/works/{id}`, `DELETE /admin/works/{id}`
- `POST /admin/authors`, `PUT /admin/authors/{id}`, `DELETE /admin/authors/{id}`
- `POST /admin/publishers`, `PUT /admin/publishers/{id}`, `DELETE /admin/publishers/{id}`
- `POST /admin/locations`, `PUT /admin/locations/{id}`, `DELETE /admin/locations/{id}`
- `POST /admin/users`, `PUT /admin/users/{id}`, `DELETE /admin/users/{id}`

## Штрих-коды

Backend генерирует EAN-13 коды для книг и локаций. Изображения штрих-кодов формируются в PNG. Планируется интеграция с очередью сообщений для отправки заданий на печать во внешний сервис.
