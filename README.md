# elibrary

`elibrary` — сервис учета библиотечного фонда с Go backend и React frontend. Проект предоставляет REST API для каталога книг, авторов, произведений, издателей, пользователей и локаций хранения, поддерживает JWT-аутентификацию, ролевую модель доступа, генерацию штрих-кодов EAN-13 и подготовку заданий на печать.

## Что умеет проект

- публичный и внутренний API для книг;
- CRUD для произведений, авторов, издателей, локаций и пользователей;
- JWT-аутентификация;
- RBAC с ролью `admin`;
- генерация и валидация EAN-13;
- хранение изображений сущностей на локальном диске;
- SQL-миграции PostgreSQL;
- запуск через Docker Compose.

## Стек

- Backend: Go, `chi`, `pgx`, `jwt`
- Frontend: React, TypeScript, Vite
- Database: PostgreSQL 15
- Messaging: RabbitMQ
- Infra: Docker, Docker Compose, `migrate`

## Структура репозитория

```text
cmd/server              точка входа backend
internal/config         загрузка конфигурации
internal/domain         доменные сущности и ошибки
internal/service        бизнес-логика
internal/repository     интерфейсы репозиториев
internal/repository/postgres реализации репозиториев для PostgreSQL
internal/http           router, handlers, middleware
internal/readmodel      структуры ответов API
internal/storage        интерфейсы для хранения файлов
internal/storage/local  локальное файловое хранилище изображений
migrations              SQL-миграции
client                  frontend на React/Vite
```

## Требования

- Go 1.24+
- Docker и Docker Compose
- Node.js 20+ и `npm` для локального запуска frontend
- PostgreSQL 15+ для запуска backend без Docker
- утилита `migrate` для ручного прогона миграций

## Быстрый старт через Docker

Основной способ поднять проект:

```bash
cp .env.example .env
# отредактируй .env и задай свои реальные значения
docker compose up --build
```

`docker compose` прочитает переменные из локального файла `.env`, поэтому секреты больше не хранятся в `docker-compose.yml`.

Что поднимется:

- backend: `http://localhost:8080`
- frontend: доступен через контейнер `frontend` в сети `web`
- postgres: контейнер `elibrary-postgres`
- rabbitmq: `amqp://localhost:5672`, management UI `http://localhost:15672`

Замечание: в `docker-compose.yml` frontend подключен к внешней Docker-сети `web`. Если такой сети нет, создай ее заранее:

```bash
docker network create web
```

Также backend монтирует изображения в `/home/root/elibrary-data/images` на хосте. Эта директория должна существовать и быть доступной Docker.

## Локальный запуск backend без Docker

1. Подними PostgreSQL и RabbitMQ.
2. Прогони миграции.
3. Установи переменные окружения.
4. Запусти сервер.

Пример:

```bash
export DB_URL='postgres://elibrary:NEW_DB_PASSWORD@localhost:5432/elibrary?sslmode=disable'
export HTTP_ADDR=':8080'
export JWT_SECRET='NEW_LONG_RANDOM_JWT_SECRET'
export CORS_ALLOWED_ORIGINS='http://localhost:5173'
export IMAGES_PATH='./data/images'
export IMAGES_URL='/static/images'
export RABBIT_URL='amqp://printer_user:NEW_RABBIT_PASSWORD@localhost:5672/'
export RABBIT_QUEUE='print_queue'

go run ./cmd/server
```

## Миграции

В репозитории есть `Makefile` для ручного запуска миграций:

```bash
make migrate-up
make migrate-down
```

Теперь `DB_URL` нужно передавать явно:

```bash
make migrate-up DB_URL='postgres://user:pass@localhost:5432/db?sslmode=disable'
```

## Переменные окружения backend

Обязательные для реального запуска:

- `DB_URL` — строка подключения к PostgreSQL
- `JWT_SECRET` — секрет для подписи JWT
- `RABBIT_URL` — адрес RabbitMQ с актуальными учетными данными

Поддерживаемые настройки:

- `HTTP_ADDR` — адрес HTTP-сервера, по умолчанию `:8080`
- `CORS_ALLOWED_ORIGINS` — список origin через запятую, по умолчанию `http://localhost:5173,http://localhost:3000`
- `IMAGES_PATH` — локальный путь для хранения изображений, по умолчанию `./data/images`
- `IMAGES_URL` — URL-префикс для раздачи изображений, по умолчанию `/static/images`
- `RABBIT_QUEUE` — имя очереди печати, по умолчанию `print_queue`

Если не задать `DB_URL`, `JWT_SECRET` или `RABBIT_URL`, backend завершится на старте с явной ошибкой.

## Локальный запуск frontend

```bash
cd client
npm install
npm run dev
```

Дополнительные команды:

```bash
cd client
npm run build
npm run lint
```

## Тесты

Запуск всех Go-тестов:

```bash
go test ./...
```

Если в окружении есть проблема с правами на системный Go cache, используй:

```bash
env GOCACHE=/tmp/go-build-test go test ./...
```

С подробным выводом:

```bash
env GOCACHE=/tmp/go-build-test go test -v ./...
```

## Основные маршруты API

### Публичные

- `GET /health`
- `POST /auth/login`

### Доступные после JWT

- `GET /books/public`
- `GET /books/public/{id}`
- `GET /books/internal`
- `GET /books/internal/{id}`
- `GET /works/{id}`
- `GET /authors/{id}`
- `GET /publishers/{id}`
- `GET /locations/type/{type}`
- `GET /locations/{id}`
- `GET /locations/child/{id}/{type}`
- `GET /reference/authors`
- `GET /reference/works`
- `GET /reference/publishers`

### Только для `admin`

- `POST /admin/{entity}/{id}/image`
- `POST /admin/print`
- `GET|POST /admin/roles`
- `GET /admin/permissions`
- `GET|POST|PUT|DELETE /admin/users`
- `POST|PUT /admin/books`
- `GET|POST|PUT|DELETE /admin/works`
- `GET|POST|PUT|DELETE /admin/authors`
- `GET|POST|PUT|DELETE /admin/publishers`
- `GET|POST|PUT|DELETE /admin/locations`

## Аутентификация

Для защищенных маршрутов нужен заголовок:

```text
Authorization: Bearer <jwt-token>
```

JWT выдается через `POST /auth/login`.

## Изображения

Изображения сущностей сохраняются через локальное файловое хранилище:

- файлы кладутся в `IMAGES_PATH/<entity>/<uuid>/photo.jpg`
- backend раздает их через `IMAGES_URL`

Поддерживаемые типы сущностей:

- `author`
- `book`
- `publisher`

## Штрих-коды

Backend генерирует EAN-13 для книг и локаций. Для валидного кода можно получить PNG-изображение штрих-кода, которое затем может быть отправлено в очередь печати.

## Очередь сообщений и печать

Для асинхронной печати используется RabbitMQ. Backend не отправляет штрих-коды напрямую на принтер: он формирует задание на печать и кладет его в очередь, откуда его должен забрать отдельный consumer или внешний сервис печати.

Это нужно для того, чтобы:

- не блокировать HTTP-запрос во время печати;
- отделить API от конкретного принтера или драйвера;
- обрабатывать задания независимо и масштабировать печать отдельно от backend;
- переживать временную недоступность сервиса печати без остановки API.

Ключевые настройки:

- `RABBIT_URL` — адрес подключения к RabbitMQ;
- `RABBIT_QUEUE` — имя очереди, по умолчанию `print_queue`.

## Как запускать на сервере с новыми секретами

1. Создай файл `.env` рядом с `docker-compose.yml` на основе [.env.example](/home/qwerty/elibrary/.env.example).
2. Заполни в нем новые значения как минимум для `POSTGRES_PASSWORD`, `RABBITMQ_DEFAULT_PASS`, `JWT_SECRET`, `DB_URL` и `RABBIT_URL`.
3. Проверь, что пароли внутри `DB_URL` и `RABBIT_URL` совпадают с `POSTGRES_PASSWORD` и `RABBITMQ_DEFAULT_PASS`.
4. Запусти `docker compose up -d --build`.

Для запуска backend без Docker задай те же `DB_URL`, `JWT_SECRET` и `RABBIT_URL` через переменные окружения сервиса (`systemd EnvironmentFile`, секреты CI/CD, Ansible vault, и т.д.), затем запускай `go run ./cmd/server` или свой собранный бинарник.

Где это в проекте:

- [internal/service/print_queue.go](/home/qwerty/elibrary/internal/service/print_queue.go) — отправка сообщений в RabbitMQ;
- `POST /admin/print` — admin-эндпоинт, который ставит задачу в очередь.

Как это связано со штрих-кодами:

1. Backend генерирует EAN-13 код для книги или локации.
2. По этому коду строится PNG-изображение штрих-кода.
3. Данные для печати передаются в очередь сообщений.
4. Внешний worker или сервис печати читает сообщение из очереди и выполняет фактическую печать.

Именно здесь очередь сообщений встроена в доменный поток: генерация штрих-кода происходит внутри backend, а доставка задания на печать вынесена в RabbitMQ.
