# Спринт 0: Инфраструктурный фундамент и старт проектов

## 1. Введение и цели спринта
Цель Спринта 0 — подготовить полноценную отказоустойчивую среду для локальной разработки, исключив любые «костыли».
Мы закладываем:
1. Мультисервисный стек в Docker Compose:
   - **PostgreSQL 16 с расширением pgvector** (для реляционных данных и RAG AI).
   - **Redis 7** (для кэша сессий, rate limiting и WebSocket Pub/Sub).
   - **RabbitMQ 3.13** (брокер для тяжелых фоновых задач Celery).
   - **MinIO S3** (объектное хранилище файлов документов и медиа).
   - **Apache Kafka 3.8 (KRaft mode)** (распределенный журнал событий и Event-Driven Architecture) + **Kafka-UI**.
2. Каркас **FastAPI (Python 3.14)**:
   - Pydantic Settings для строгой валидации переменных окружения (Fail-Fast принцип).
   - Единый **API Envelope** (`ok`, `data`, `error`).
   - Иерархия доменных исключений (`AppException`).
   - **Correlation ID Middleware** для сквозного отслеживания запросов (`X-Request-ID`).
   - Асинхронное ядро базы данных на базе **SQLAlchemy 2.0 Async + asyncpg** (`pool_pre_ping=True`, `expire_on_commit=False`).
   - Модуль **Healthcheck** (`/api/v1/health/live` и `/api/v1/health/ready` с реальным пингом Postgres и Redis).
3. Каркас **Go 1.26+**:
   - Чистая архитектура (Clean Architecture / Standard Go Layout).
   - Пул соединений `pgxpool`.
   - Зеркальный API Envelope.
4. Каркас **Frontend (Vue 3 + TypeScript + Tailwind + Storybook)**.

---

## 2. Что создано и почему (Шаг 0.1: Инфраструктура)

### 2.1 Файлы переменных окружения
- [infra/.env](../../infra/.env) и [infra/.env.example](../../infra/.env.example):
  - **Принцип 12-Factor App (Config)**: параметры подключения, пароли и порты никогда не хардкодятся в коде или compose-файле.
  - **RabbitMQ Gotcha**: официальный образ RabbitMQ использует переменную `RABBITMQ_DEFAULT_PASS` (без слова `WORD`), иначе пароль игнорируется и создается дефолтный `guest`.

### 2.2 Оркестрация контейнеров ([infra/docker-compose.yml](../../infra/docker-compose.yml))
1. **PostgreSQL 16 + pgvector (`pgvector/pgvector:pg16`)**:
   - Использование официального образа `pgvector` гарантирует наличие расширения для векторного поиска (HNSW-индексы, оператор косинусного расстояния `<=>`).
   - Настроен healthcheck через `pg_isready`.
2. **Redis 7 (`redis:7-alpine`)**:
   - Запущен с обязательной защитой паролем (`--requirepass`).
   - Используется для rate limiting, токен-блэклиста и WebSocket Pub/Sub.
3. **RabbitMQ 3.13 (`rabbitmq:3.13-management-alpine`)**:
   - AMQP-порт `5672` для Celery воркеров и веб-дашборд на порту `15672`.
4. **MinIO S3 (`firstfinger/minio:latest`)**:
   - **Важный DevOps-кейс**: В конце 2025 года MinIO закрыл публичный доступ к образам `minio/minio` и `minio/mc` на Docker Hub. Был выбран проверенный community-образ `firstfinger/minio:latest`, содержащий как сервер, так и утилиту `mc`.
   - **Разница между ENTRYPOINT и CMD**: в образе `firstfinger/minio` поле `Entrypoint` пустое, поэтому в `command` передается бинарник целиком: `["minio", "server", "/data", "--console-address", ":9001"]`.
   - **Авто-инициализация бакетов (`minio-init-buckets`)**: одноразовый контейнер создает бакеты `crm-documents`, `crm-attachments`, `crm-recordings` сразу при старте проекта.
5. **Apache Kafka 3.8 в режиме KRaft (`apache/kafka:3.8.0`)**:
   - KRaft режим устраняет необходимость в тяжелом Apache ZooKeeper.
   - **Концепция Advertised Listeners**:
     - `PLAINTEXT://kafka:9092` — для межконтейнерного взаимодействия внутри сети Docker.
     - `PLAINTEXT_HOST://localhost:9094` — для бэкендов, запускаемых локально на хосте разработчика.
   - **Kafka-UI (`provectuslabs/kafka-ui:latest`)**: веб-интерфейс на порту `8080` для инспекции топиков, сообщений, партиций и консьюмер-групп.

---

## 3. Что создано и почему (Шаг 0.2: Каркас FastAPI на Python 3.14)

### 3.1 Зависимости и пакетный менеджер
- [backend-python/pyproject.toml](../../backend-python/pyproject.toml):
  - Использован **`uv`** (пакетный менеджер нового поколения на Rust) по стандарту **PEP 621**.
  - Зафиксирован стек: `fastapi`, `uvicorn[standard]`, `pydantic-settings`, `sqlalchemy[asyncio]`, `asyncpg`, `redis`, `aiokafka`, `pwdlib[argon2]`, `pyjwt`, `httpx`.
  - Настроен линтер `ruff` и тестовый фреймворк `pytest-asyncio`.

### 3.2 Конфигурация и валидация
- [backend-python/src/core/config.py](../../backend-python/src/core/config.py):
  - Класс `Settings` на базе `pydantic-settings`.
  - Принцип **Fail-Fast**: приложение не запустится, если в `.env` отсутствует обязательный секрет или передан неверный тип данных.
  - Синглтон через `@lru_cache def get_settings()`: позволяет легко подменять конфигурацию в тестах через `app.dependency_overrides`.

### 3.3 Подключение к PostgreSQL
- [backend-python/src/core/database.py](../../backend-python/src/core/database.py):
  - Асинхронный движок `create_async_engine` с диалектом `postgresql+asyncpg`.
  - `pool_pre_ping=True`: автоматический пинг (`SELECT 1`) перед выдачей соединения из пула, защищает от разрывов соединений по таймауту.
  - `expire_on_commit=False`: предотвращает падение `MissingGreenlet` при чтении полей ORM после коммита.
  - Базовый класс `Base(AsyncAttrs, DeclarativeBase)` для моделей.
  - Функция генератора сессий `get_db()` с автооткатом (`rollback()`) при ошибках.

### 3.4 Контракты ответов и ошибок
- [backend-python/src/common/envelope.py](../../backend-python/src/common/envelope.py):
  - Единый контракт `ApiResponse[T]` (PEP 695 Type Parameter Syntax): `{ "ok": true, "data": ..., "error": null }`.
- [backend-python/src/common/exceptions.py](../../backend-python/src/common/exceptions.py):
  - Иерархия доменных исключений (`AppException`, `NotFoundException`, `UnauthorizedException`, `ForbiddenException`, `ConflictException`, `TenantAccessDeniedException`).
  - Бизнес-логика сервисов изолирована от HTTP-слоя.

### 3.5 Middleware и перехватчики
- [backend-python/src/middleware/correlation.py](../../backend-python/src/middleware/correlation.py):
  - Сквозной Correlation ID через корутино-безопасный `contextvars.ContextVar`.
  - Проброс и возврат заголовка `X-Request-ID`.
- [backend-python/src/middleware/error_handler.py](../../backend-python/src/middleware/error_handler.py):
  - Глобальный перехват `AppException`, `RequestValidationError` и общих `Exception`.
  - Защита от утечки системных данных и стектрейсов наружу.

### 3.6 Модуль здоровья и точка входа
- [backend-python/src/modules/health/routes.py](../../backend-python/src/modules/health/routes.py):
  - `/api/v1/health/live`: проверка живости процесса.
  - `/api/v1/health/ready`: реальный пинг PostgreSQL (`SELECT 1`) и Redis (`PING`). При сбое возвращает `503 Service Unavailable`.
- [backend-python/src/main.py](../../backend-python/src/main.py):
  - Современный `@asynccontextmanager lifespan` с корректным Graceful Shutdown (`async_engine.dispose()`).
  - Подключение CORS, CorrelationId, Exception Handlers и роутера.

---

## 4. Что создано и почему (Шаг 0.4: Каркас Go 1.26+)

### 4.1 Структура проекта (Standard Go Layout)
- [backend-go/cmd/server/main.go](../../backend-go/cmd/server/main.go):
  - Минималистичная точка входа. Настраивает структурированный JSON-логгер `log/slog`, читает переменные окружения и подписывается на сигналы завершения ОС (`SIGINT`, `SIGTERM`) через `signal.NotifyContext`.
- [backend-go/internal/app/app.go](../../backend-go/internal/app/app.go):
  - Оркестратор жизненного цикла приложения. Инициализирует пул PostgreSQL, клиент Redis, собирает Chi-роутер с middleware, стартует HTTP-сервер в отдельной горутине и реализует **Graceful Shutdown** с таймаутом дренажа соединений (10 секунд).

### 4.2 Конфигурация и окружение
- [backend-go/internal/config/config.go](../../backend-go/internal/config/config.go):
  - Парсинг конфигурации через библиотеку `caarlos0/env/v11` со строгой типизацией (`AppConfig`, `PostgresConfig`, `RedisConfig`).
  - Поддержка чтения `.env` файлов на этапе разработки через `godotenv`.

### 4.3 Доступ к данным и pgvector
- [backend-go/internal/database/postgres.go](../../backend-go/internal/database/postgres.go):
  - Пул соединений на базе `github.com/jackc/pgx/v5/pgxpool`.
  - Конфигурация пула (`MinConns`, `MaxConns`, `MaxConnIdleTime`, `HealthCheckPeriod`).
  - Fail-Fast проверка связи при старте (`pool.Ping`).
  - Идемпотентная активация расширения для векторного поиска (`CREATE EXTENSION IF NOT EXISTS vector;`).

### 4.4 Зеркальный API Envelope и Middleware
- [backend-go/pkg/response/envelope.go](../../backend-go/pkg/response/envelope.go):
  - Дженерик-структура `ApiResponse[T any]` полностью совпадает по формату с Python-версией: `{ "ok": bool, "data": *T, "error": *ErrorDetail }`.
  - Хелперы `response.JSON(w, statusCode, data)` и `response.Error(w, statusCode, code, message, details)`.
- [backend-go/internal/middleware/correlation.go](../../backend-go/internal/middleware/correlation.go):
  - Middleware для извлечения или генерации сквозного `X-Request-ID` (UUID v4), сохраняющий его в `context.Context` запроса и в ответные HTTP-заголовки.

### 4.5 Модуль Healthcheck и тестирование
- [backend-go/internal/modules/health/handler.go](../../backend-go/internal/modules/health/handler.go):
  - Роуты `/api/v1/health/live` (проверка жизнеспособности) и `/api/v1/health/ready` (проверка пинга Postgres и Redis).
- [backend-go/internal/modules/health/handler_test.go](../../backend-go/internal/modules/health/handler_test.go):
  - Юнит-тестирование эндпоинта живости через `httptest.ResponseRecorder` с валидацией структуры ответа.

---

## 5. Что создано и почему (Шаг 0.5: Каркас Frontend на Vue 3, Vite, Tailwind CSS и Storybook)

### 5.1 Архитектурный стек фронтенда
- [frontend/package.json](../../frontend/package.json):
  - **Vue 3 (`<script setup lang="ts">`)**: Современный Composition API с нативным Script Setup — устраняет бойлерплейт `return { ... }` и обеспечивает авто-экспонирование переменных в шаблон.
  - **Vite 8**: Сверхбыстрый HMR на базе нативных ES-модулей браузера в dev-режиме и оптимизированный Rollup-бандлер для продакшена.
  - **TypeScript 7.0**: Строгая типизация пропсов, контрактов и отсутствие неявного `any`.

### 5.2 Стилизация и адаптивная тема (Tailwind CSS v4)
- [frontend/src/style.css](../../frontend/src/style.css) и [frontend/tailwind.config.js](../../frontend/tailwind.config.js):
  - **Tailwind v4 Gotcha**: директива `@import "tailwindcss";` в связке с `@config "../tailwind.config.js";` для обратной совместимости с пользовательскими палитрами (`brand-500` индиго) и ручным переключением тем `darkMode: 'class'`.
  - **Data-Dense палитра**: компактные размеры и CSS-переменные для светлой (`:root`) и темной (`.dark`) тем.
  - **Интеграция с `@vueuse/core` (`useDark`)**: автоматическое считывание системной темы ОС (`prefers-color-scheme`) и сохранение ручного выбора пользователя в `localStorage`.

### 5.3 Типизированный сетевой мост (Axios)
- [frontend/src/api/types.ts](../../frontend/src/api/types.ts):
  - Зеркальный контракт `ApiResponse<T>`: строго соответствует бэкендам FastAPI и Go (`{ ok, data, error }`).
  - Класс доменных ошибок `AppApiError`.
- [frontend/src/api/client.ts](../../frontend/src/api/client.ts):
  - **Request Interceptor**: автоматическая генерация сквозного UUID `X-Request-ID` (Correlation ID) через `crypto.randomUUID()`.
  - **Response Interceptor**: авто-распаковка Envelope (`payload?.data ?? payload`) при успехе и трансляция серверных кодов ошибок в типизированный `AppApiError`.
  - Лаконичные хелперы `api.get<T>()` и `api.post<T>()`.
- [frontend/src/api/modules/health.ts](../../frontend/src/api/modules/health.ts):
  - Методы `checkLiveness()` и `checkReadiness()`, подключенные к реальным серверам.

### 5.4 Компонентный подход: дизайн-примитив `Button`
- [frontend/src/lib/utils.ts](../../frontend/src/lib/utils.ts):
  - Утилита `cn(...)` на базе `clsx` и `tailwind-merge` для безопасного разрешения конфликтов классов Tailwind при переопределении стилей родителем.
- [frontend/src/components/ui/button/Button.vue](../../frontend/src/components/ui/button/Button.vue):
  - **Lookup Table паттерн**: словари `variantStyles` и `sizeStyles` типа `Record<string, string>` вместо громоздких `if/else`.
  - Доступность (**a11y**): использование `focus-visible:` вместо `focus:` предотвращает «залипание» фокуса при кликах мышью, сохраняя обводку при навигации клавишей `Tab`.
  - Варианты: `primary`, `secondary`, `outline`, `ghost`, `danger`.
  - Состояния: встроенный спиннер `Loader2` (`loading`), `disabled`.
  - Паттерн «входных ворот» (**Barrel Export**): [frontend/src/components/ui/button/index.ts](../../frontend/src/components/ui/button/index.ts) для лаконичного импорта.

### 5.5 Изолированная разработка в Storybook 8
- [frontend/.storybook/main.ts](../../frontend/.storybook/main.ts) и [frontend/.storybook/preview.ts](../../frontend/.storybook/preview.ts):
  - Интеграция с `@storybook/vue3-vite` и `@storybook/addon-essentials`.
  - Импорт `../src/style.css` для рендеринга стилей Tailwind внутри Canvas.
  - Центрированный лейаут (`layout: 'centered'`).
- [frontend/src/components/ui/button/Button.stories.ts](../../frontend/src/components/ui/button/Button.stories.ts):
  - Спецификация CSF 3 со всеми историями (`Primary`, `Secondary`, `Outline`, `Ghost`, `Danger`, `Loading`, `Disabled`) и интерактивной панелью Controls.

---

## 6. Доступные веб-интерфейсы инфраструктуры

| Сервис | Назначение | URL | Логин / Пароль (из .env) |
|---|---|---|---|
| **Frontend CRM (Vite Dev)** | Веб-интерфейс платформы CRM AI | [http://localhost:5173](http://localhost:5173) | Без авторизации |
| **Storybook** | Каталог UI-компонентов и дизайн-система | [http://localhost:6006](http://localhost:6006) | Без авторизации |
| **FastAPI Swagger UI** | Автодокументация API | [http://localhost:8000/docs](http://localhost:8000/docs) | Без авторизации |
| **Kafka UI** | Мониторинг топиков и событий Kafka | [http://localhost:8080](http://localhost:8080) | Без авторизации |
| **MinIO Console** | S3 веб-хранилище документов | [http://localhost:9001](http://localhost:9001) | `minio_admin` / `minio_password` |
| **RabbitMQ UI** | Мониторинг очередей задач | [http://localhost:15672](http://localhost:15672) | `rabbit_user` / `rabbit_password` |
| **PostgreSQL** | Реляционная СУБД + pgvector | `localhost:5432` | `crm_admin` / `crm_password` (`crm_db`) |
| **Redis** | In-memory кэш и Pub/Sub | `localhost:6379` | пароль: `redis_password` |
