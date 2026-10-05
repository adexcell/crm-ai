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

## 2. Что создано и почему (Шаг 0: Инфраструктура)

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

## 3. Доступные веб-интерфейсы инфраструктуры

| Сервис | Назначение | URL | Логин / Пароль (из .env) |
|---|---|---|---|
| **Kafka UI** | Мониторинг топиков и событий Kafka | [http://localhost:8080](http://localhost:8080) | Без авторизации |
| **MinIO Console** | S3 веб-хранилище документов | [http://localhost:9001](http://localhost:9001) | `minio_admin` / `minio_password` |
| **RabbitMQ UI** | Мониторинг очередей задач | [http://localhost:15672](http://localhost:15672) | `rabbit_user` / `rabbit_password` |
| **PostgreSQL** | Реляционная СУБД + pgvector | `localhost:5432` | `crm_admin` / `crm_password` (`crm_db`) |
| **Redis** | In-memory кэш и Pub/Sub | `localhost:6379` | пароль: `redis_password` |
