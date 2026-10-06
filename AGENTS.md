# CRM AI: Engineering & Mentorship Rules

## 1. Роль и формат менторства
- **Роль**: Архитектор и Senior/Lead ментор.
- **Подход**:
  - Разбираем каждое архитектурное решение («почему именно так», компромиссы, безопасность, масштабируемость).
  - Пишем код сразу **Production Ready** — без временных хаков, `TODO: fix later`, нетипизированного кода и захардкоженных секретов.
  - Двойная реализация бэкэнда: **FastAPI (Python)** и **Go** идут синхронно по модулям, разделяя одинаковые API-контракты и схему БД.
  - Фронтенд пишется модуль за модулем на **Vue 3**, сопровождаясь дизайн-системой и фиксацией компонентов в **Storybook**.

---

## 2. Архитектура репозитория (Monorepo)
```text
crm-ai/
├── contracts/               # Единые OpenAPI спецификации, общие схемы данных
├── infra/                   # Docker Compose, Nginx, конфигурации PostgreSQL, Redis, MinIO
├── backend-python/          # Python 3.14+ (FastAPI, SQLAlchemy 2.0 Async, Pydantic v2, Ruff)
├── backend-go/              # Go 1.26+ (Standard Project Layout, Chi/Fiber, pgx/sqlc, slog)
├── frontend/                # Vue 3, Vite, TypeScript, Tailwind CSS, Pinia, Storybook
└── TECHNICAL.md             # Исходное техническое задание
```

---

## 3. Стандарты Production-Ready: Python (FastAPI)
1. **Слоистая архитектура в каждом модуле (`modules/<domain>/`)**:
   - `models.py`: SQLAlchemy 2.0 ORM модели (строгая типизация `Mapped[...]`, `mapped_column`).
   - `schemas.py`: Pydantic v2 схемы (разделение `CreateDTO`, `UpdateDTO`, `ResponseDTO`, `config = ConfigDict(from_attributes=True)`).
   - `repository.py`: Изолированный слой доступа к данным. Все запросы через `AsyncSession`, фильтрация по `organization_id` обязательна.
   - `service.py`: Доменная бизнес-логика, транзакционные границы, проверка бизнес-инвариантов.
   - `routes.py`: Тонкие HTTP/WebSocket эндпоинты, маппинг на статус-коды, Dependency Injection через `Depends()`.
2. **API Envelope**:
   - Все REST-ответы строго заворачиваются в контракт:
     ```json
     { "ok": true, "data": { ... }, "error": null }
     ```
   - Ошибки:
     ```json
     { "ok": false, "data": null, "error": { "code": "STRING_CODE", "message": "Human readable", "details": {} } }
     ```
3. **Безопасность и Middleware**:
   - HttpOnly Secure SameSite куки для JWT (`access_token`, `refresh_token`), Token Rotation + Blacklist/Redis.
   - Мультитенантность: `current_org_id` резолвится через DI/Middleware, ни один запрос не должен утекать между тенантами.
   - Correlation ID (RequestId) во всех запросах и структурированных логах.
   - Rate limiting, валидация размера тела запросов, защита от SQL injection и XSS.

---

## 4. Стандарты Production-Ready: Go
1. **Архитектурный стандарт**:
   - `cmd/server/main.go`: точка входа, конфигурация, graceful shutdown (`os.Interrupt`, `syscall.SIGTERM`).
   - `internal/config/`: строгая конфигурация через переменные окружения (`caarlos0/env` или `cleanenv`).
   - `internal/domain/`: доменные сущности, ошибки, интерфейсы репозиториев и сервисов.
   - `internal/modules/<domain>/`: реализация (Handler -> Service -> Repository).
   - `pkg/`: переиспользуемые библиотеки (httpresponse, database, jwt, logger).
2. **База данных и драйвер**:
   - Использование `pgx/v5` (пул соединений `pgxpool`, поддержка транзакций, строгая типизация параметров).
   - Контексты `context.Context` пробрасываются во все методы без исключения.
3. **Идиоматика Go**:
   - Никаких неконтролируемых `panic()` в рантайме.
   - Обработка ошибок через `fmt.Errorf("%w")` и `errors.Is`/`errors.As`.
   - Структурированное логирование через `log/slog` с атрибутами `correlation_id`, `org_id`, `user_id`.
   - Идентичный Python-версии JSON Envelope для всех ответов и ошибок.

---

## 5. Стандарты Production-Ready: Frontend (Vue 3 + Storybook)
1. **Стек и структура**:
   - Vue 3 (`<script setup lang="ts">`), Pinia для глобального клиентского стейта.
   - Vue Router 4 с навигационными гардами (Auth, Organization context).
   - Axios или Fetch с интерцепторами для авто-рефреша токенов и перехвата 401.
2. **Дизайн и UI/UX**:
   - Премиальный дизайн: темная/светлая темы, акцентные палитры, плавные анимации, data-dense интерфейс для CRM.
   - UI-примитивы (Button, Input, Dropdown, Modal, Toast, Badge, DataGrid) выносятся в дизайн-систему.
   - Доступность (a11y), клавиатурная навигация, понятные состояния загрузки (скелетоны, спиннеры) и пустые состояния (empty states).
3. **Storybook**:
   - Для каждого компонента в `src/components/ui/` и модульных компонентов пишется `.stories.ts`.
   - Истории покрывают все варианты: Primary, Secondary, Disabled, Loading, WithIcon, DarkMode, Error.

---

## 6. Порядок работы над каждым модулем
1. **Обсуждение концепта и схемы**: Разбор доменной логики, сущностей, связей и эндпоинтов.
2. **БД и Спецификация**: Определение схемы миграции и контракта API (Request/Response).
3. **FastAPI реализация**: Модель -> Репозиторий -> Сервис -> Роуты -> Интеграционный тест.
4. **Go реализация**: Сущность -> Репозиторий -> Сервис -> Хэндлер -> Интеграционный тест.
5. **Vue 3 UI + Storybook**: Создание визуальных компонентов в Storybook -> Сборка модуля -> Подключение API.
6. **Ревью и фиксация**: Проверка соответствия DOD (Definition of Done) и переход к следующему модулю.
