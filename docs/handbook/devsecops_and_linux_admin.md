# Руководство по DevSecOps, Linux System Administration & Безопасности

> **Цель этого документа**: Практическое руководство по управлению серверами Ubuntu/Linux (и Termux), глубокому пониманию процессов, работе с Docker из чистой консоли, защите боевых серверов (Hardening) и настройке автоматического CI/CD с Telegram-уведомлениями.

---

## 1. Управление процессами, сессиями и портами

### 1.1 `tmux` (Terminal Multiplexer)
Позволяет создавать изолированные сессии терминала, которые не прерываются при закрытии окна, разрыве SSH или потере Wi-Fi.

| Команда | Описание |
| :--- | :--- |
| `tmux new -s <имя>` | Создать новую сессию с понятным именем (например: `tmux new -s backend`) |
| `Ctrl + B`, затем `D` | **Detach**: отключиться от сессии, оставив процессы работать в фоне |
| `tmux ls` | Показать список всех активных сессий |
| `tmux attach -t <имя>` | Подключиться обратно к работающей сессии |
| `Ctrl + B`, затем `%` | Разделить текущий экран **по вертикали** на 2 панели |
| `Ctrl + B`, затем `"` | Разделить текущий экран **по горизонтали** на 2 панели |
| `Ctrl + B`, затем стрелки | Перемещение между панелями |
| `tmux kill-session -t <имя>` | Завершить и закрыть сессию со всеми процессами |

---

### 1.2 Поиск и освобождение портов
Когда сервер выдает ошибку `Address already in use: 8000`:

```bash
# 1. Поиск процесса через ss (Socket Statistics):
# -t (TCP), -u (UDP), -l (Listening), -p (Process name/PID), -n (Numeric ports)
ss -tulpn | grep 8000

# 2. Поиск через lsof (List Open Files):
sudo lsof -i :8000

# 3. Завершение процесса:
# Вежливое завершение (SIGTERM = 15) — даем процессу закрыть БД и освободить ресурсы:
kill <PID>

# Принудительное завершение ядра (SIGKILL = 9) — если процесс завис:
kill -9 <PID>

# Завершение по шаблону имени:
pkill -f "uvicorn"
pkill -f "vite"
```

---

### 1.3 Интерактивный мониторинг ресурсов
- **`btop`**: Лучший интерактивный TUI-монитор нагрузки CPU, памяти, дисков, процессов и сетевого ввода/вывода (установка: `sudo apt install -y btop`).
- **`htop`**: Классический системный монитор с фильтрацией по F4 и деревом процессов по F5.

---

## 2. Администрирование Docker напрямую из консоли

### 2.1 Изоляция сетей (Docker Networks)
Контейнеры общаются между собой по внутренним именам хостов только тогда, когда находятся в одной пользовательской сети.

```bash
# 1. Создание сети bridge:
docker network create crm-net

# 2. Список сетей:
docker network ls

# 3. Инспекция сети (показывает подключенные контейнеры и их внутренние IP):
docker network inspect crm-net
```

---

### 2.2 Запуск сервисов чистыми командами
Пример запуска базы данных PostgreSQL с векторным расширением:

```bash
docker run -d \
  --name crm-postgres \
  --network crm-net \
  -p 127.0.0.1:5432:5432 \
  -e POSTGRES_USER=crm_admin \
  -e POSTGRES_PASSWORD=crm_password \
  -e POSTGRES_DB=crm_db \
  -v crm_pgdata:/var/lib/postgresql/data \
  --restart unless-stopped \
  pgvector/pgvector:pg16
```

> [!IMPORTANT]
> **Правило безопасности сетевых портов**:
> Обратите внимание на флаг `-p 127.0.0.1:5432:5432` вместо `-p 5432:5432`.
> Если написать просто `-p 5432:5432`, Docker автоматически откроет порт на `0.0.0.0` (весь интернет), обойдя правила файрвола `ufw`! Указание `127.0.0.1:` привязывает порт строго к локальной петле сервера.

---

### 2.3 Чтение логов, мониторинг и отладка
```bash
# 1. Просмотр логов в реальном времени:
docker logs -f --tail 100 crm-postgres

# 2. Интерактивный вход внутрь контейнера:
docker exec -it crm-postgres bash

# 3. Мониторинг потребления ресурсов контейнерами (CPU, Memory, Network I/O):
docker stats

# 4. Очистка неиспользуемых зависших контейнеров и слоев:
docker system prune -f
```

---

## 3. Server Hardening: Максимальная безопасность Linux-сервера

Когда сервер разворачивается в публичном интернете (VPS):

### 3.1 Защита SSH (`/etc/ssh/sshd_config`)
1. **Генерация стойкого ключа на компьютере разработчика**:
   ```bash
   ssh-keygen -t ed25519 -C "admin@crm-ai"
   ```
2. **Копирование ключа на сервер**:
   ```bash
   ssh-copy-id -i ~/.ssh/id_ed25519.pub user@server_ip
   ```
3. **Отключение входа по паролям и под пользователем root**:
   В файле `/etc/ssh/sshd_config`:
   ```text
   PermitRootLogin no
   PasswordAuthentication no
   PubkeyAuthentication yes
   MaxAuthTries 3
   ```
   Перезапуск демона: `sudo systemctl restart sshd`.

---

### 3.2 Файрвол `ufw` (Uncomplicated Firewall)
Открываем только то, что строго необходимо для работы пользователей:

```bash
# По умолчанию запретить весь входящий трафик и разрешить исходящий:
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Разрешить SSH:
sudo ufw allow 22/tcp

# Разрешить веб-трафик HTTP и HTTPS:
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Включить файрвол:
sudo ufw enable

# Проверить статус правил:
sudo ufw status verbose
```

---

### 3.3 Защита от брутфорс-атак (`fail2ban`)
Автоматически анализирует системные логи и блокирует IP-адреса ботов, которые пытаются подобрать учетные данные:
```bash
sudo apt install -y fail2ban
sudo systemctl enable --now fail2ban
# Проверить статус бана SSH:
sudo fail2ban-client status sshd
```

---

## 4. Архитектура CI/CD (GitHub Actions + Telegram Bot)

### 4.1 Пайплайн непрерывной интеграции и доставки (CI/CD)
Файл `.github/workflows/deploy.yml`:
1. **Этап 1: Линтинг и тесты (CI)**:
   - Запуск `ruff check` и `pytest` для Python.
   - Запуск `golangci-lint` и `go test` для Go.
   - Сборка TypeScript/Vite для Frontend.
2. **Этап 2: Сборка и публикация контейнеров (Build & Push)**:
   - Сборка образов для backend-python, backend-go, frontend.
   - Пуш в защищенный реестр `ghcr.io` (GitHub Packages).
3. **Этап 3: Деплой на боевой сервер (CD)**:
   - Подключение по SSH-ключу к серверу.
   - Выполнение `docker compose pull && docker compose up -d --remove-orphans`.
   - Запуск миграций Alembic: `docker compose exec backend-python uv run alembic upgrade head`.
4. **Этап 4: Уведомление в Telegram**:
   - Отправка статуса деплоя (успех / падение) с указанием автора, коммита и времени.

---

## 5. Полезные CLI-утилиты (Ubuntu & Zsh)

- **`zoxide`**: Умный переход по директориям (`z crm` вместо `cd /home/abakar/projects/crm-ai-python`).
- **`fzf`**: Интерактивный нечеткий поиск по истории (комбинация клавиш `Ctrl + R`).
- **`bat`**: Просмотр файлов с подсветкой синтаксиса (замена `cat`).
- **`eza`**: Современная замена `ls` с отображением прав, иконок и цветов.
