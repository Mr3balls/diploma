# ACE Esports Platform

Платформа для управления киберспортивными турнирами. Поддерживает несколько форматов сеток, командную и индивидуальную регистрацию, реалтайм уведомления и чат.

## Стек

| Слой | Технологии |
|---|---|
| Backend | Go 1.23, Chi, PostgreSQL 16, Redis 7 |
| Frontend | React 18, TypeScript, Vite, Tailwind CSS |
| Realtime | SSE (Server-Sent Events), Web Push |
| Инфраструктура | Docker Compose, golang-migrate |

## Быстрый старт

### Docker (рекомендуется)

```bash
cd back
cp .env.example .env   # отредактируйте при необходимости
make docker-up
```

API: `http://localhost:8080` · Postgres: `localhost:55432`

### Локально

```bash
# Backend
cd back
cp .env.example .env
# В .env заменить хосты: postgres → localhost, redis → localhost
make migrate-up
make run

# Frontend (отдельный терминал)
cd front
npm install
npm run dev
```

Frontend: `http://localhost:5173` · API: `http://localhost:8080`

## Переменные окружения

Файл `back/.env` (пример в `back/.env.example`):

```env
APP_ENV=development
HTTP_PORT=8080

# JWT
ACCESS_TOKEN_SECRET=super-secret-change-me
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=720h

# БД и кэш
DATABASE_URL=postgres://postgres:postgres@postgres:5432/esports?sslmode=disable
REDIS_URL=redis://redis:6379/0

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# Email (опционально — отключить, оставив пустым)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@gmail.com
SMTP_PASSWORD=xxxx xxxx xxxx xxxx
SMTP_FROM=your@gmail.com

# Google Sheets импорт (опционально)
GOOGLE_SERVICE_ACCOUNT_FILE=
GOOGLE_DEFAULT_WORKSHEET=Sheet1

# Web Push (опционально)
VAPID_PRIVATE_KEY=
VAPID_PUBLIC_KEY=
VAPID_EMAIL=

AUTH_RATE_LIMIT_PER_MINUTE=30
```

## Команды

```bash
make run            # запуск API
make build          # сборка бинарника
make test           # тесты
make migrate-up     # применить миграции
make migrate-down   # откатить последнюю миграцию
make seed           # загрузить seed-данные
make docker-up      # docker compose up --build
make docker-down    # docker compose down -v
```

## Функциональность

### Форматы турниров
- **Single Elimination** — классическая олимпийская сетка
- **Double Elimination** — с сеткой проигравших (Winner/Loser Bracket + Grand Final)
- **Group Stage** — групповой этап с round-robin
- **Group + DE** — групповой этап с последующим double elimination

### Управление
- Создание турниров с настройкой формата, лимита команд, дат
- Командная и индивидуальная регистрация участников
- Генерация и сброс сетки, пересев
- Выставление времени и места матчей
- Запись и подтверждение результатов матчей
- Импорт участников из Google Sheets

### Уведомления
- Realtime через SSE (Server-Sent Events)
- Web Push (PWA)
- Email (SMTP)
- Управление предпочтениями уведомлений

### Прочее
- Чат турнира в реальном времени
- Карта места проведения (Leaflet / OpenStreetMap)
- Аудит-лог всех действий
- Мультиязычность (RU / EN / KAZ)
- Challonge-совместимый API

## Структура проекта

```
├── back/                   # Go backend
│   ├── cmd/api/main.go     # точка входа
│   ├── internal/
│   │   ├── entity/         # модели данных
│   │   ├── repository/     # слой БД
│   │   ├── service/        # бизнес-логика
│   │   ├── transport/http/ # HTTP хендлеры и роутер
│   │   └── pkg/            # утилиты (notif, xjson, …)
│   ├── migrations/         # SQL миграции
│   └── docker-compose.yml
│
└── front/                  # React frontend
    └── src/
        ├── app/            # провайдеры, роутер
        ├── features/       # фичи (tournaments, matches, …)
        ├── pages/          # страницы
        ├── shared/         # UI-компоненты, типы, хуки
        └── widgets/        # navbar
```

## API

Базовый URL: `http://localhost:8080`

Аутентификация: Bearer JWT в заголовке `Authorization`.

Основные группы эндпоинтов:

| Группа | Префикс |
|---|---|
| Аутентификация | `/auth/*` |
| Пользователи | `/users/*` |
| Турниры | `/tournaments/*` |
| Матчи | `/matches/*` |
| Команды | `/teams/*` |
| Уведомления | `/notifications/*` |
| Challonge API | `/challonge/*` |
| Платформ-admin | `/platform-admin/*` |

## Роли

| Роль | Возможности |
|---|---|
| Игрок | регистрация в турнирах, управление командой |
| Менеджер турнира | управление турниром, результатами |
| Владелец турнира | все права + добавление менеджеров |
| Platform Admin | полный доступ ко всем турнирам и пользователям |
