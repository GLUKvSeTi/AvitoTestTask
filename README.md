# AvitoTestTask

Микросервис для назначения ревьюеров для Pull Request’ов

## Описание

Проект представляет собой REST API сервис для  назначения ревьюеров для Pull Request’ов с возможностью автоматического добавления/удаления пользователей.

## Технологии

- **Go 1.24+** - основной язык программирования
- **PostgreSQL** - база данных
- **Docker** - контейнеризация
- **Make** - автоматизация сборки (опционально)

## Требования

- Go 1.24.4 или выше
- Docker и Docker Compose
- Git

## Проблемы
- Дебаг модулей - решение: Посидел почитал как исправляют ошибки
- Заболел - решение: Полечился и потерял несколько дней
- Новые технологии (новый рутер)
- Сделал сам себе доп работу добавив CRUD'ы

## Быстрый старт (рекомедованный)
```bash
# Клонирование репозитория
git clone https://github.com/GLUKvSeTi/AvitoTestTask.git
cd AvitoTestTask

# Запуск приложения и базы данных
docker compose up --build
```

## Для Linux/MacOS
```bash

git clone https://github.com/GLUKvSeTi/AvitoTestTask.git
cd AvitoTestTask

# Сборка приложения
make build

# Запуск приложения
make run

# Запуск через Docker Compose
make docker-compose-up
```
## Локальный запуск
```bash
git clone https://github.com/GLUKvSeTi/AvitoTestTask.git
cd AvitoTestTask

# Установка зависимостей
go mod download

# Запуск приложения
go run ./cmd/app
```

## сервис работает на localhost:8080

## Пример запроса 
```bash
curl -X POST localhost:8080/user/create \
  -H "Content-Type: application/json" \
  -d '{
    "user_id":"11111111-1111-1111-1111-111111111111",
    "username":"alice",
    "team_id": null
  }'
```
## Ожидаемый ответ
```bash
{"user_id":"11111111-1111-1111-1111-111111111111"}
```




