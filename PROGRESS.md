# 🧭 Operational Progress Tracker: CRM AI

> **Последнее обновление**: 2026-10-31
> **Текущий фокус**: Спринт 1 — Аутентификация, сессии и мультитенантность (Go-First)
> **Статус компиляции**: `go build ./...` — ✅ PASS | `go vet ./...` — ✅ PASS

---

## 🎯 Стратегический курс (Architecture Roadmap)
- **Основной бэкенд**: **100% Go (Golang 1.23+)** — ядро системы, REST/WebSocket API, PostgreSQL, Redis, безопасность, транзакции и мультитенантность.
- **Python-бэкенд**: Отложен. Будет реализован точечно в виде выделенного микросервиса **только для AI/ML задач** (LLM-агенты, Whisper ASR транскрибация звонков Asterisk).
- **Инфраструктура & CI/CD**: Домашний кластер **k3s** на ноутбуке, self-hosted **Gitea Actions**, **Cloudflare Tunnel**, с последующим зеркалированием стабильного кода в **GitHub**.
- **Фронтенд**: **Vue 3** (Composition API, TypeScript, Pinia, Tailwind CSS, Storybook).

---

## 🏆 Что уже выполнено (Done)

### 1. Спецификации и База знаний
- [x] Единая инженерная спецификация модуля Auth: `docs/specs/auth.md` (инварианты, матрица краевых кейсов EC-01 — EC-06).
- [x] База знаний в Obsidian (`Проектирование сервисов/`):
  - Модель угроз и мультитенантность (01).
  - Матрица 15 эндпоинтов и API envelope (02).
  - Криптография Argon2id, токены, Timing attacks (03).
  - Сценарии онбординга, инвайтов, сброса паролей (04).
  - Grace Period (15с) и Web Locks при рефреше токенов (05).
  - Паттерны отказоустойчивости и долговечности (Resilience: Timeouts, RateLimit, Singleflight, CircuitBreaker).

### 2. Backend Go (`backend-go`) — Полная реализация модуля Auth
- [x] **Шаг 1 (Онбординг)**: `EmailSendCode`, `EmailVerifyCode`, `Register` (атомарная транзакция User + Org + Member).
- [x] **Шаг 2 (Сессии и Токены)**: `Login`, `RefreshToken` (RTR, Replay Attack Killswitch, Grace Period 15с), `Logout`, `GetSessions`, `RevokeSession`, `RevokeAllSessions`.
- [x] **Шаг 3 (Восстановление доступа)**: `ForgotPassword` (anti-enumeration через 200 OK + Redis TTL 15m), `ResetPassword` (SHA-256 токен, `GetDel`, атомарная смена пароля и сброс всех сессий).
- [x] **Шаг 4 (Профиль и Мультитенантность)**: `GetProfile` (данные + членства), `UpdateProfile`, `SwitchOrganization` (IDOR защита, проверка членства, перевыпуск токена).
- [x] **Шаг 5 (DI & Lifecycle)**:
  - `internal/infra/http/request/request.go`: безопасный декодер с защитой от DoS (1MB MaxBytesReader) и валидацией.
  - `internal/app/auth_domain.go`: чистая сборка зависимостей, валидация DI, возврат `nil, err`, регистрация роутов `/api/v1/auth`.
  - `internal/app/app.go`: Graceful Shutdown через `<-ctx.Done()` (управление контекстом сигналов из `main.go`), закрытие ресурсов через `defer` в LIFO-порядке (Server -> Redis -> Postgres).

---

## ⏳ В работе прямо сейчас (NOW)
- [x] **Ручная проверка эндпоинтов модуля Auth (`contracts/http/auth.http`)**:
  - `POST /register`: ✅ 201 Created (хэш Argon2id, slug collision check, cookie).
  - `POST /login`: ✅ 200 OK (проверка пароля, выдача токенов).
  - `GET /me`: ✅ 200 OK (валидация claims, профиль и список организаций).
  - `PATCH /me`: ✅ 200 OK (обновление имени и фамилии).
  - `POST /refresh`: ✅ 200 OK (RTR ротация токенов в атомарной транзакции).
  - **Replay Attack Killswitch**: ✅ 401 Unauthorized (мгновенный отзыв всех сессий пользователя при повторе токена).
  - `POST /logout`: ✅ 200 OK (сброс куки, маркировка сессии отозванной).
  - `GET /sessions`: ✅ 200 OK (список активных устройств с `is_current: true`).
  - `POST /forgot-password`: ✅ 200 OK (anti-enumeration, SHA-256 токен в Redis).
  - `POST /reset-password`: ✅ 200 OK (GETDEL из Redis, Argon2id, атомарный сброс сессий).
- [x] **Финальные проверки перед коммитом**:
  - Проверка логина с новым паролем (`POST /login`) и отказ со старым (401) — ✅ PASS.
  - Проверка смены организации (`POST /switch-org`) и защита от IDOR (403 Forbidden + security WARN) — ✅ PASS.
  - Проверка ручного отзыва сессий (`DELETE /sessions/{id}`, `POST /sessions/revoke-all`) — ✅ PASS.
- [ ] **Фиксация Спринта 1 в Git и CI/CD**:
  - Атомарный коммит реализации модуля Auth на Go в репозиторий.
  - Настройка пайплайна CI/CD в Gitea (`.gitea/workflows/ci.yaml`): линтинг (`golangci-lint`), компиляция, тесты, авто-зеркалирование в GitHub.

---

## 📋 Следующие шаги (Next Steps)
1. **Интеграционные тесты Go**:
   - Конкурентный тест (`-race`) на 10 одновременных запросов `/refresh` для проверки Grace Period 15 секунд.
2. **Подготовка манифестов k3s**:
   - Deployment, Service, ConfigMap, Secret, Ingress для запуска Go-бэкенда в k3s.
   - Подключение Cloudflare Tunnel для доступа снаружи.
3. **Frontend (Vue 3 + Storybook)**:
   - Инициализация Pinia store для авторизации (хранение токена в RAM).
   - Настройка интерцепторов с поддержкой Web Locks API для рефреша токена без гонок между вкладками.

---

## 🛡️ Архитектурные соглашения (Gotchas & Invariants)
- **Lifecycle**: В `app.go` слушаем `<-ctx.Done()`, внешние сигналы ловит `main.go`.
- **Ресурсы**: Закрываются строго через `defer` сразу после успешной инициализации.
- **Хранение токенов**: Refresh-токены в PostgreSQL хранятся **только** в виде SHA-256 хэшей.
- **Безопасность кук**: Refresh токен передается строго в `Set-Cookie: HttpOnly; Secure; SameSite=Lax`.
- **API Envelope**: Все ответы оборачиваются в `{ "ok": true, "data": ..., "error": null }`.
