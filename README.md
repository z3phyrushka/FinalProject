Планировщик задач (Выпускной проект курса "Go-разработчик с нуля")

Это веб-приложение для управления задачами, которое поддерживает одноразовые и повторяющиеся задачи, редактирование, удаление, отметку выполнения, поиск, аутентификацию и запуск через Docker.

Данные хранятся в SQLite.


1. Назначение проекта

Проект реализует серверную часть простого, но функционального планировщика задач с веб-интерфейсом:

> создание задач;

> поддержка правил повторения (d, y, w, m);

> редактирование задач;

> выполнение и автоматический расчёт следующей даты;

> удаление задач;

> поиск по тексту и по дате;

> аутентификация по паролю через JWT;

> Docker-образ для запуска на любой машине.


2. Выполненные задания со звёздочкой

✔ Полная поддержка правил повторения: d, y, w, m
✔ Полная поддержка поиска (search → текст и дата 02.01.2006)
✔ JWT-аутентификация через cookie token
✔ Dockerfile для сборки образа
✔ docker-compose (compose.yml)
✔ Полное прохождение всех тестов (go test ./tests)


3. Структура проекта
Final-project/
│
├── .github/
│   └── workflows/
│       └── tests.yml        # CI для тестов
│
├── pkg/
│   ├── api/
│   │   ├── addtask.go       # POST /api/task
│   │   ├── api.go           # регистрация маршрутов
│   │   ├── auth.go          # POST /api/signin и middleware auth
│   │   ├── done.go          # POST /api/task/done
│   │   ├── nextdate.go      # логика NextDate()
│   │   └── tasks.go         # GET /api/tasks
│   │
│   └── db/
│       ├── db.go            # инициализация SQLite
│       └── task.go          # CRUD-функции: Add, Update, Delete, Get, List
│
├── tests/                   # тесты проекта
│
├── web/                     # фронтенд
│
├── compose.yml              # docker-compose для запуска
├── Dockerfile               # сборка Docker-образа
├── envFile.env              # пример env-файла
├── scheduler.db             # SQLite база
├── go.mod
├── go.sum
├── main.go
└── README.md


4. Локальный запуск
Переменные окружения

Файл .env (пример):

TODO_PORT=7540
TODO_DBFILE=scheduler.db
TODO_PASSWORD=12345


Можно загрузить их:

set -a
source .env
set +a

Запуск
go run .

Открыть в браузере
http://localhost:7540


Если задан пароль (TODO_PASSWORD):

http://localhost:7540/login.html


5. API тестирование
Для запуска всех тестов:
go test ./tests

Если включена аутентификация (TODO_PASSWORD ≠ "")

Получить токен:

curl -X POST -d "{\"password\":\"12345\"}" http://localhost:7540/api/signin


Вставить токен в tests/settings.go:

Token = "<полученный токен>"


Запустить тесты:

go test ./tests


6. Docker
Сборка изображения
docker build -t scheduler .

Запуск через docker-compose
docker compose up -d


После запуска:

http://localhost:7540


База данных хранится в volume или привязке к хосту — зависит от compose.yml
(обычно ./scheduler.db:/app/scheduler.db).

Ручной запуск без compose
docker run -d \
  -p 7540:7540 \
  -e TODO_PORT=7540 \
  -e TODO_DBFILE=/data/scheduler.db \
  -e TODO_PASSWORD=12345 \
  -v ./scheduler.db:/data/scheduler.db \
  --name scheduler \
  scheduler


7. Запуск сервера в продакшене

собрать бинарь для Linux:

GOOS=linux GOARCH=amd64 go build -o scheduler


загрузить бинарь + папку web/ на сервер;

настроить systemd или supervisor;

использовать reverse-proxy (nginx) при необходимости.


8. Список API
Метод	Путь	Описание
POST	/api/signin	Аутентификация
GET	/api/tasks	Список задач + поиск
POST	/api/task	Добавить задачу
GET	/api/task?id=ID	Получить задачу
PUT	/api/task	Обновить задачу
DELETE	/api/task?id=ID	Удалить задачу
POST	/api/task/done?id=ID	Отметить выполненной


9. Проект готов к проверке

> Все тесты проходят

> Все задания (включая задания со звёздочками) выполнены

> Docker и README добавлены

> Код структурирован и соответствует ТЗ