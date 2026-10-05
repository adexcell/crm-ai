# Спринт 0: Инфраструктурный фундамент и старт проектов

## 1. Введение и цели спринта
Цель Спринта 0 — подготовить отказоустойчивую среду для локальной разработки, исключив любые «костыли».
Мы закладываем:
1. Мультисервисный стек в Docker Compose:
   - **PostgreSQL 16 с расширением pgvector** (для реляционных данных и RAG AI).
   - **Redis 7** (для кэша сессий, rate limiting и WebSocket Pub/Sub).
   - **RabbitMQ 3.13** (брокер для тяжелых фоновых задач).
   - **MinIO** (S3-совместимое хранилище файлов документов и записей).
2. Каркас **FastAPI (Python 3.14)**:
   - Pydantic Settings для строгой валидации переменных окружения.
   - Единый **API Envelope** (`ok`, `data`, `error`).
   - Иерархия доменных исключений (`AppException`).
   - **Correlation ID Middleware** для сквозного отслеживания запросов (`X-Request-ID`).
   - Асинхронное ядро базы данных на базе **SQLAlchemy 2.0 Async + asyncpg**.
3. Каркас **Go 1.26+**:
   - Чистая архитектура (Clean Architecture / Standard Go Layout).
   - Пул соединений `pgxpool`.
   - Зеркальный API Envelope.
4. Каркас **Frontend (Vue 3 + TypeScript + Tailwind + Storybook)**.

---

## 2. Что уже создано и почему

### 2.1 Инфраструктура (`infra/`)
- [infra/docker-compose.yml](../../infra/docker-compose.yml):
  - **Образ БД `pgvector/pgvector:pg16`**: стандартный образ postgres не содержит расширения для векторов. Использование официального образа pgvector экономит время сборки и гарантирует совместимость с HNSW индексами.
  - **Автосоздание бакетов MinIO (`minio-init-buckets`)**: сервис-инициализатор, который запускает `mc` после старта MinIO и создает бакеты `crm-documents`, `crm-attachments`, `crm-recordings`.
- [infra/.env](../../infra/.env): переменные для портов и паролей.

### 2.2 Каркас FastAPI (`backend-python/`)
- [pyproject.toml](../../backend-python/pyproject.toml): декларация зависимостей современного стека (FastAPI 0.115+, SQLAlchemy 2.0 Async, Pydantic v2, Ruff).
- [src/core/config.py](../../backend-python/src/core/config.py):
  - **Зачем BaseSettings**: типизированные переменные с валидацией при старте приложения. Если в `.env` отсутствует критичный параметр или передан неверный тип (например порт строкой, которая не парсится в int), сервер упадет сразу на старте с понятной ошибкой (Fail-Fast принцип).
- [src/common/envelope.py](../../backend-python/src/common/envelope.py):
  - **Зачем API Envelope**: единый контракт фронтенда и сторонних API-клиентов. Клиенту никогда не нужно гадать, в каком ключе лежит ошибка или результат. Всегда есть булевый флаг `ok`.
- [src/common/exceptions.py](../../backend-python/src/common/exceptions.py):
  - **Зачем доменные исключения**: изоляция бизнес-слоя от HTTP-специфики. Сервисы не должны знать про FastAPI `HTTPException(status_code=404)`. Сервис бросает `NotFoundException("User", id)`, а глобальный обработчик маппит его в HTTP 404 и красивый JSON Envelope.
- [src/middleware/correlation.py](../../backend-python/src/middleware/correlation.py):
  - **Зачем Correlation ID**: каждый входящий запрос получает UUID. В распределенной системе или микросервисах по этому ID в логах можно отследить весь путь запроса (от Nginx через API в базу данных и фоновый воркер).
