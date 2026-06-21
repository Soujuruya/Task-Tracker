# Task Tracker

[![Go Version](https://img.shields.io/github/go-mod/go-version/Soujuruya/Task-Tracker)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/Soujuruya/Task-Tracker)](https://goreportcard.com/report/github.com/Soujuruya/Task-Tracker)

Task Tracker - HTTP API на Go для регистрации пользователей, выдачи JWT/refresh-токенов и управления задачами. Проект сделан как учебный backend-сервис с разделением на слои, доменной логикой, in-memory и PostgreSQL-хранилищем, graceful shutdown и базовыми security-практиками.

## Что показывает проект

- Проектирование HTTP API на стандартной библиотеке Go.
- Разделение кода на `transport`, `service`, `repository`, `domain`.
- Авторизация через access JWT и refresh-token rotation.
- Безопасное хранение паролей через Argon2id с настраиваемыми параметрами.
- Хранение refresh-токенов в виде SHA-256 хэшей.
- Работа с PostgreSQL через `database/sql` и `pgx` driver.
- Параметризованные SQL-запросы без конкатенации пользовательского ввода.
- In-memory репозитории с `sync.RWMutex` для локального запуска и тестирования идей.
- Graceful shutdown HTTP-сервера.
- Rate limiting для публичных auth-эндпоинтов.
- Пагинация, фильтрация и сортировка задач.
- Доменная state machine для статусов задач.
- Audit log изменений задач на уровне сервиса.
- Конфигурация через переменные окружения.

## Быстрый старт

### Требования

- Go 1.25+
- Make
- PostgreSQL, если используется `STORAGE_TYPE=postgres`
- `psql`, если запускаются миграции через Makefile

### Настройка окружения

Создайте `.env` на основе `.env.example`:

```bash
cp .env.example .env
```

Минимальный вариант для запуска без PostgreSQL:

```bash
export ENVIRONMENT=development
export AUTH_SERVICE_ADDR=:8080
export JWT_SECRET=dev_secret
export ACCESS_TOKEN_TTL=30m
export REFRESH_TOKEN_TTL=48h

export MEMORY=64
export ITERATIONS=3
export PARALLELISM=2
export SALT_LENGTH=16
export KEY_LENGTH=32
export MAX_CONCURRENCY=12

export RATE_LIMIT_MAX_REQUESTS=10
export RATE_LIMIT_WINDOW_SIZE=1m

export STORAGE_TYPE=memory
```

Для PostgreSQL:

```bash
export STORAGE_TYPE=postgres
export POSTGRES_DSN=postgres://postgres:postgres@127.0.0.1:5432/task_tracker?sslmode=disable
```

### Запуск

```bash
make run
```

Или напрямую:

```bash
source .env
go run ./cmd
```

### Миграции PostgreSQL

```bash
make migrate-up
```

Откат:

```bash
make migrate-down
```

## Архитектура

```text
cmd/
  main.go                         # точка входа

internal/
  app/                            # сборка зависимостей и запуск приложения
  config/                         # чтение env-конфигурации
  domain/                         # доменные модели, статусы, ошибки
  service/                        # бизнес-логика auth, tokens, tasks
  repository/                     # интерфейсы и реализации хранилищ
    user/
    task/
    refresh_token/
    auditlog/
  transport/                      # HTTP server, handlers, middleware, DTO
  pkg/                            # вспомогательные пакеты: hasher, logger, ctx keys

migrations/                       # SQL-миграции PostgreSQL
```

Основная идея: HTTP-слой отвечает за запросы, ответы и middleware; сервисный слой содержит правила приложения; репозитории скрывают детали хранения; доменный слой содержит модели и ошибки.

## API

По умолчанию сервис слушает `:8080`.

### Auth

| Method | Path | Auth | Описание |
| --- | --- | --- | --- |
| `POST` | `/register` | no | Регистрация пользователя |
| `POST` | `/login` | no | Получение access и refresh токена |
| `POST` | `/refresh` | no | Ротация refresh-токена и выдача новой пары токенов |
| `POST` | `/logout` | yes | Отзыв активной refresh-сессии пользователя |
| `GET` | `/validate` | bearer | Проверка access-токена |

### Tasks

| Method | Path | Auth | Описание |
| --- | --- | --- | --- |
| `POST` | `/tasks` | yes | Создать задачу |
| `GET` | `/tasks` | yes | Получить список задач |
| `PUT` | `/tasks/{id}` | yes | Частично обновить задачу |
| `DELETE` | `/tasks/{id}` | yes | Удалить задачу |
| `GET` | `/tasks/{id}/history` | yes | Получить историю изменений задачи |

Авторизация передается через заголовок:

```http
Authorization: Bearer <access_token>
```

## Примеры запросов

### Регистрация

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demo",
    "password": "password123"
  }'
```

Ответ:

```json
{
  "user_id": "019..."
}
```

### Логин

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demo",
    "password": "password123"
  }'
```

Ответ:

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<refresh_token>"
}
```

### Создание задачи

```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "title": "Read Go docs",
    "description": "Review net/http and context packages",
    "progress_status": "todo"
  }'
```

Ответ:

```json
{
  "id": "019...",
  "title": "Read Go docs",
  "description": "Review net/http and context packages",
  "progress_status": "todo",
  "created_at": "2026-06-22T10:00:00Z"
}
```

### Получение списка задач

```bash
curl "http://localhost:8080/tasks?status=todo&page=1&page_size=10" \
  -H "Authorization: Bearer <access_token>"
```

Поддерживаемые query-параметры:

| Parameter | Описание |
| --- | --- |
| `status` | `todo`, `in_progress`, `done`, `blocked` |
| `created_from` | дата в формате RFC3339 |
| `created_to` | дата в формате RFC3339 |
| `page` | номер страницы, начиная с `1` |
| `page_size` | размер страницы, от `1` до `100` |

### Обновление задачи

`PUT /tasks/{id}` работает как частичное обновление: меняются только переданные поля.

```bash
curl -X PUT http://localhost:8080/tasks/<task_id> \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "progress_status": "in_progress"
  }'
```

### Ротация refresh-токена

```bash
curl -X POST http://localhost:8080/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<refresh_token>"
  }'
```

## Бизнес-правила задач

Статусы задач:

- `todo`
- `in_progress`
- `blocked`
- `done`

Разрешенные переходы:

| From | To |
| --- | --- |
| `todo` | `in_progress`, `blocked` |
| `in_progress` | `todo`, `blocked`, `done` |
| `blocked` | `todo`, `in_progress` |
| `done` | нет переходов |

Дополнительные правила:

- `title` обязателен при создании задачи.
- Если статус при создании не передан, используется `todo`.
- Задачи сортируются по `created_at` от новых к старым.
- Задачу со статусом `done` нельзя удалить.
- История изменений доступна только владельцу задачи.

## Конфигурация

| Variable | Описание | Пример |
| --- | --- | --- |
| `ENVIRONMENT` | окружение: `development` или `production` | `development` |
| `AUTH_SERVICE_ADDR` | адрес HTTP-сервера | `:8080` |
| `JWT_SECRET` | секрет подписи JWT | `change_me` |
| `ACCESS_TOKEN_TTL` | срок жизни access-токена | `30m` |
| `REFRESH_TOKEN_TTL` | срок жизни refresh-токена | `48h` |
| `MEMORY` | память Argon2id в мегабайтах | `64` |
| `ITERATIONS` | количество итераций Argon2id | `3` |
| `PARALLELISM` | параллелизм Argon2id | `2` |
| `SALT_LENGTH` | длина соли в байтах | `16` |
| `KEY_LENGTH` | длина хэша в байтах | `32` |
| `MAX_CONCURRENCY` | лимит одновременных хэширований | `12` |
| `RATE_LIMIT_MAX_REQUESTS` | максимум запросов в окно | `10` |
| `RATE_LIMIT_WINDOW_SIZE` | размер окна rate limiter | `1m` |
| `STORAGE_TYPE` | тип хранения: `memory` или `postgres` | `postgres` |
| `POSTGRES_DSN` | DSN подключения к PostgreSQL | `postgres://...` |

В `production` обязательные переменные должны быть заданы явно. В `development` часть значений имеет безопасные для локальной разработки defaults.

## Хранение данных

Проект поддерживает два режима хранения:

- `memory` - пользователи, задачи, refresh-токены и audit log живут в памяти процесса.
- `postgres` - пользователи и задачи хранятся в PostgreSQL, refresh-токены и audit log пока остаются in-memory.

Это осознанное текущее ограничение проекта: PostgreSQL-репозитории для refresh-токенов и audit log можно добавить следующими шагами.

## Безопасность

В проекте реализованы следующие практики:

- Пароли не хранятся в открытом виде.
- Для паролей используется Argon2id.
- Сравнение password hash выполняется через constant-time compare.
- Access token подписывается через HMAC SHA-256.
- Refresh token хранится как SHA-256 hash.
- Refresh token rotation инвалидирует предыдущий refresh-токен.
- SQL-запросы используют placeholders.
- Публичные auth-эндпоинты защищены rate limiter.
- Сервис не возвращает внутренние ошибки клиенту напрямую.

Что стоит улучшить перед production:

- Хранить refresh-токены и audit log в PostgreSQL.
- Добавить HTTPS/TLS termination на уровне reverse proxy или сервера.
- Добавить CORS/security headers при появлении frontend-клиента.
- Добавить `golangci-lint`, `govulncheck` и CI.

## Проверка проекта

```bash
go test ./...
go vet ./...
```

Дополнительно рекомендуется:

```bash
golangci-lint run ./...
govulncheck ./...
```

## Текущий статус

Реализовано:

- регистрация и логин;
- access JWT;
- refresh-token rotation;
- logout;
- middleware авторизации;
- CRUD задач;
- фильтрация и пагинация задач;
- история изменений задач;
- memory и PostgreSQL репозитории для пользователей и задач;
- SQL-миграции для пользователей и задач.

В планах:

- PostgreSQL-хранилище для refresh-токенов;
- PostgreSQL-хранилище для audit log;
- GitHub Actions;
- OpenAPI-спецификация.

