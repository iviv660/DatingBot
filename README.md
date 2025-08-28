# DatingBot 💘

## Содержание
- [Краткое описание](#Краткое-описание)
- [Стек технологий](#-стек-технологий)
- [Структура проекта](#структура-проекта)
- [Установка и настройка](#-установка-и-запуск)
---

## 📄 Краткое описание
Телеграм-бот для знакомств. Позволяет пользователям находить интересных собеседников, обмениваться лайками и начинать общение.
Основные возможности:
- 📌 Регистрация и создание профиля
- 🔍 Подбор потенциальных пар
- ❤️ Лайки и дизлайки для поиска совпадени
- 🔔 Уведомления о новых лайках и мэтчах
----

## 🚀 Стек технологий
- **Go** 1.24
- **gRPC** (protobuf)
- **PostgreSQL** (хранение пользователей и матчей)
- **Redis** (кэш)
- **MinIO** (хранение фотографий)
- **Docker & docker-compose**
- **Telebot v4** (Telegram Bot SDK)
---

## 💡 Структура проекта
```bash
match # Match Service
notifier # TelegramBot service
uesr # User Service
docker-compose.yml
Dockerfile
README.md
```

## 🛠️ Установка и запуск

#### 1.Клонируй проект:
```bash
git clone https://github.com/iviv660/DatingBot.git
cd tasker
```
#### 2.Создай .env (пример ниже):
```bash
#---------------- User Service ----------------
USER_POSTGRES_USER=user_postgres
USER_POSTGRES_PASSWORD=070823
USER_POSTGRES_DB=user_db
USER_POSTGRES_PORT=5432
USER_POSTGRES_DSN=postgres://user_postgres:070823@user_postgres:5432/user_db?sslmode=disable

USER_REDIS_DSN=redis://user_redis:6379/0

USER_MINIO_ENDPOINT=http://user_minio:9000
USER_MINIO_BASE_URL=http://user_minio:9000
USER_MINIO_ACCESS_KEY=user_minio
USER_MINIO_SECRET_KEY=070823Vo
USER_MINIO_BUCKET=users
USER_MINIO_USE_SSL=false

USER_GRPC_PORT=:50051

#---------------- Match Service ---------------
MATCH_POSTGRES_USER=match_postgres
MATCH_POSTGRES_PASSWORD=070823
MATCH_POSTGRES_DB=match_db
MATCH_POSTGRES_PORT=5432
MATCH_POSTGRES_DSN=postgres://match_postgres:070823@match_postgres:5432/match_db?sslmode=disable

MATCH_USER_CLIENT=user_service:50051
MATCH_GRPC_PORT=:50052

#---------------- Notifier Service ---------------
TELEGRAM_BOT_TOKEN=!
```
#### 3.Запусти в Docker:
```bash
docker compose up --build
```

