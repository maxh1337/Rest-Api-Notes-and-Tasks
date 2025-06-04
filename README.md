# 📝 Rest API Notes and Tasks (Echo + GORM)

> Мой первый проект на Go с использованием фреймворка Echo(Первый RestApi на Golang в целом). Построен как RESTful API для управления пользователями, задачами и подзадачами. Основная цель — разобраться с архитектурой, роутингом, слоями (handler → service → repository) и применением GORM + PostgreSQL.
> У меня уже есть огромный опыт написания Fullstack проектов на TypeScript (NestJs,NextJs), так как я много писал на NestJs, а там используется структура controller → service, в данном проекте я решил убрать слой usecases. В первую очередь из-за непривычности разбиения логики на такие маленькие куски с логикой. Во вторую очередь мне кажется, что в таком маленьком проекте, даже подход с Clean Architecture является избыточным.

---

## ⚙️ Стек технологий

### Задействовано

- [Go (Golang)](https://golang.org/)
- [Echo](https://echo.labstack.com/) — HTTP веб-фреймворк
- [GORM](https://gorm.io/) — ORM для работы с PostgreSQL
- [PostgreSQL](https://www.postgresql.org/) — база данных
- [Redis](https://pkg.go.dev/github.com/redis/go-redis/v9) - Кэширование пользовательских сессий и токенов
- [JWT](https://pkg.go.dev/github.com/golang-jwt/jwt/v5) - http-only cookie accessToken и refreshToken для аутентификации пользователя.
- UUID — в качестве идентификаторов вместо `uint`

### Планируется

- Дореализовать метод для переобновления токенов
- Добавить Middlewares - auth, role, cors, ratelimiting, CSRF
- Полная реализации логики работы с задачами и подазадачами
- 2FA с помощью кода Telegram Codes

---

## 📁 Структура проекта

```

├── cmd/server/ # Точка входа: main / инициализация сервера
├── internal/
│ ├── api/ # Delivery слой
│ │ ├── handlers/ # HTTP-обработчики (Echo)
│ │ ├── middleware/ # Middleware (логирование, аутентификация и т.д.)
│ │ └── routes/ # Роутинг
│ ├── config/ # Загрузка конфигурации и .env
│ ├── domain/ # Сущности, интерфейсы, бизнес-логика
│ │ ├── entities/ # Модели (User, Task, SubTask и т.д.)
│ │ ├── repositories/ # Интерфейсы и реализация репозиториев
│ │ ├── services/ # Сервисный слой (логика приложения)
│ │ └── validator/ # Первичная валидация для json полей
│ └── infrastructure/
│ └── database/ # Инициализация базы данных (PostgreSQL)
├── docker/ # Dockerfile + docker-compose.yml
├── scripts/ # Вспомогательные скрипты
├── tmp/ # Временные файлы
└── Makefile # Команды сборки, запуска, миграций и т.п.

```

---

## 🚀 Как запустить

1. Установи зависимости:

   ```bash
   go mod tidy

   ```

2. Настрой переменные окружения (`.env` или через `os.Getenv`):

   ```
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=password
   DB_NAME=notes_db
   PORT=8080
   ```

3. Запусти сервер:

   ```bash
   make dev

   ```

4. API будет доступно по адресу:

   http://localhost:8080

```

```
