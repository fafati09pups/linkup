# LINKUP — веб-мессенджер (Go + PostgreSQL)

## Запуск

```bash
# 1. База данных
createdb linkup
psql linkup -f database/schema.sql
psql linkup -f database/seed.sql   # по желанию: тестовые пользователи anna, boris, clara (пароль 123456)

# 2. Сервер
cd backend
go mod tidy          # скачает pgx, gorilla/websocket, bcrypt
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/linkup"
export SECRET="любая-длинная-случайная-строка"
go run .
```

Откройте http://localhost:8080 в двух браузерах (или в обычном и приватном окне),
зарегистрируйте двух пользователей и начните диалог.

## API

| Метод | Адрес                      | Назначение                      |
| ----- | -------------------------- | ------------------------------- |
| POST  | /api/register              | регистрация                     |
| POST  | /api/login                 | вход (cookie с токеном)         |
| POST  | /api/logout                | выход                           |
| GET   | /api/me                    | текущий пользователь            |
| GET   | /api/users?q=              | поиск пользователей             |
| GET   | /api/chats                 | список чатов с последним сообщением |
| POST  | /api/chats                 | открыть или создать личный чат  |
| GET   | /api/chats/{id}/messages   | история переписки               |
| POST  | /api/chats/{id}/messages   | отправить сообщение             |
| GET   | /ws                        | WebSocket для новых сообщений   |

## Структура

```text
LINKUP/
├── frontend/   index.html, style.css, script.js
├── backend/    main.go, handlers/, models/, database/, ws/, go.mod
├── database/   schema.sql, seed.sql, queries.sql
└── docs/       ER-диаграмма (добавить)
```
