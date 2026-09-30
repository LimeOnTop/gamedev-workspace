# Gamedev Workspace

Рабочее пространство для документации по разработке игры. Слева в боковом меню находится дерево папок и файлов. Главная страница показывает карточки разделов. Файл описывает один игровой объект (замок, оружие, персонажа) и содержит:

- **референсы**: изображения, видео или PDF (загрузка через кнопку или drag & drop);
- **описание**;
- **характеристики**: упорядоченные пары «параметр — значение»;
- **механики взаимодействия**.

Встроенный **MCP-сервер** даёт Claude (Claude Code, Claude Desktop) прямой доступ к рабочему пространству. Claude сам создаёт папки и файлы, заполняет описания и характеристики и прикладывает референсы через запущенный сервис. Пересобирать проект или перезапускать контейнеры для этого не нужно.

```
┌──────────────┐     ┌───────────────────────── backend (Go) ─────────────────────────┐
│  React + TS  │────▶│ controller (gin, REST) ─┐                                       │
│   (nginx)    │     │                         ├─▶ usecase ◀── service ──▶ repository ─┼─▶ PostgreSQL
└──────────────┘     │ mcp (Streamable HTTP) ──┘   (интерфейсы)   │                    │
┌──────────────┐     │                                            ├──▶ treecache ──────┼─▶ Redis
│ Claude (MCP) │────▶│                                            └──▶ filestorage ────┼─▶ volume /data/uploads
└──────────────┘     └────────────────────────────────────────────────────────────────┘
```

## Быстрый старт (Docker)

Нужен Docker с Compose v2.

```bash
cp .env.example .env
```

```bash
docker compose up -d --build
```

Откройте http://localhost:3300. При первом запуске миграции создают схему и пример структуры: «Карта → Замок, Деревня», «Оружие → Пистолет», «Персонажи».

| Сервис     | Назначение                                        | Порт на хосте                |
|------------|---------------------------------------------------|------------------------------|
| `frontend` | nginx: SPA и прокси `/api`, `/uploads`, `/mcp`    | `127.0.0.1:3300` (`WEB_PORT`) |
| `backend`  | Go-сервис: REST API и MCP                         | только внутри сети compose   |
| `migrate`  | golang-migrate, применяет `backend/migrations`    | —                            |
| `postgres` | данные                                            | `127.0.0.1:55432`            |
| `redis`    | кеш дерева папок                                  | `127.0.0.1:56379`            |

Данные хранятся в volume'ах `postgres-data` и `uploads`. Полный сброс вместе с данными:

```bash
docker compose down -v
```

Обновить сервис после изменений в коде:

```bash
docker compose up -d --build backend frontend
```

## Подключение Claude через MCP

MCP-эндпоинт (Streamable HTTP): `http://localhost:3300/mcp`.

**Claude Code.** В корне репозитория уже лежит [`.mcp.json`](.mcp.json): при запуске `claude` из этой папки сервер `gamedev-workspace` подхватится сам (Claude Code спросит подтверждение). Подключить его глобально, для любых папок, в том числе для Unity-проекта:

```bash
claude mcp add --transport http --scope user gamedev-workspace http://localhost:3300/mcp
```

**Claude Desktop.** Добавьте в `claude_desktop_config.json` (нужен Node.js):

```json
{
  "mcpServers": {
    "gamedev-workspace": {
      "command": "npx",
      "args": ["-y", "mcp-remote", "http://localhost:3300/mcp"]
    }
  }
}
```

### Инструменты

| Инструмент                  | Что делает                                                                      |
|-----------------------------|---------------------------------------------------------------------------------|
| `get_tree`                  | Всё дерево в виде текстового outline с ID узлов                                 |
| `get_node`                  | Папка или файл целиком: описание, характеристики, механики, референсы, путь, дети |
| `search_nodes`              | Поиск по названиям, описаниям, механикам и характеристикам                      |
| `create_folder`             | Создать папку (в корне или внутри другой)                                       |
| `create_file`               | Создать файл сразу с описанием, характеристиками и механиками                   |
| `update_node`               | Частичное обновление; `merge_characteristics` для upsert по ключу, `remove_characteristic_keys` для удаления |
| `move_node`                 | Переместить узел (защищено от циклов)                                           |
| `delete_node`               | Удалить узел со всем поддеревом и файлами референсов                            |
| `add_reference_from_url`    | Скачать изображение, видео или PDF по http(s)-ссылке и прикрепить к файлу       |
| `add_reference_from_base64` | Прикрепить файл из base64 или `data:` URL                                        |
| `delete_reference`          | Удалить референс                                                                |

Каждый результат содержит `web_url`, ссылку на изменённую страницу. Базовый адрес задаёт `PUBLIC_URL`.

Примеры запросов к Claude:

> Создай в «Оружии» файлы для дробовика, снайперской винтовки и ножа с характеристиками урона, скорострельности и веса.

> Пройдись по всем файлам в «Карте» и допиши механики взаимодействия там, где их нет.

> Приложи к «Замку» референс по ссылке https://…/castle.jpg

Открытая вкладка сайта обновляет дерево при возврате фокуса, поэтому изменения от Claude видны сразу.

## Структура бэкенда

Слои clean architecture по аналогии с сервисами interverse. Зависимости направлены внутрь, а слои общаются через интерфейсы из `usecase`.

```
backend/
├── cmd/
│   ├── main.go                 # сборка зависимостей, HTTP-сервер, graceful shutdown
│   └── config/config.go        # конфигурация из env / .env
├── internal/
│   ├── entity/                 # доменные сущности: Node, Reference, Characteristic
│   ├── usecase/                # ИНТЕРФЕЙСЫ и DTO
│   │   ├── node.go             #   usecase.Node: сценарии работы с деревом
│   │   ├── reference.go        #   usecase.Reference: сценарии работы с референсами
│   │   ├── node_repository.go  #   порты хранилищ…
│   │   ├── reference_repository.go
│   │   ├── tree_cache.go
│   │   ├── file_storage.go
│   │   └── dto.go
│   ├── service/                # реализация usecase.Node / usecase.Reference (бизнес-логика, валидация)
│   ├── repository/             # PostgreSQL (database/sql + lib/pq)
│   ├── treecache/              # Redis-реализация usecase.TreeCache (+ Noop без Redis)
│   ├── filestorage/            # хранение файлов на диске, реализация usecase.FileStorage
│   ├── controller/             # REST-транспорт на gin, маппинг DTO → JSON
│   ├── mcp/                    # MCP-транспорт поверх тех же usecase-интерфейсов
│   ├── middleware/             # CORS, логирование
│   └── apperr/                 # доменные ошибки и безопасные сообщения для клиента
├── migrations/                 # golang-migrate: 000001_*.up.sql / .down.sql
├── Makefile                    # run, build, test, lint, migrate-up/down/version
└── Dockerfile
```

REST и MCP — два транспорта над одними и теми же use case'ами, поэтому поведение, валидация и инвалидация кеша у них одинаковые.

## REST API

| Метод    | Путь                          | Описание                                                            |
|----------|-------------------------------|---------------------------------------------------------------------|
| `GET`    | `/api/tree`                   | Дерево папок и файлов (кешируется в Redis)                          |
| `GET`    | `/api/search?q=&limit=`       | Поиск                                                               |
| `POST`   | `/api/nodes`                  | `{parent_id?, kind: "folder"\|"file", name, description?, mechanics?, characteristics?}` |
| `GET`    | `/api/nodes/:id`              | Узел с референсами и путём                                          |
| `PATCH`  | `/api/nodes/:id`              | Частичное обновление `{name?, description?, mechanics?, characteristics?}` |
| `POST`   | `/api/nodes/:id/move`         | `{parent_id}`: `null` переносит в корень                            |
| `DELETE` | `/api/nodes/:id`              | Удалить узел и поддерево                                            |
| `POST`   | `/api/nodes/:id/references`   | `multipart/form-data`, поле `file`                                  |
| `DELETE` | `/api/references/:id`         | Удалить референс                                                    |
| `GET`    | `/uploads/:name`              | Файл референса                                                      |
| `GET`    | `/health`                     | Healthcheck                                                         |

## Локальная разработка без Docker для кода

Поднимите только инфраструктуру и примените миграции:

```bash
docker compose up -d postgres redis migrate
```

Бэкенд (порт `8081`, настройки в `backend/.env`):

```bash
cd backend && cp .env.example .env && make run
```

Фронтенд (Vite на http://localhost:5173 проксирует `/api` и `/uploads` на `localhost:8081`):

```bash
cd frontend && npm install && npm run dev
```

Тесты и линтер бэкенда:

```bash
cd backend && make test lint
```

Новая миграция: добавьте пару `backend/migrations/00000N_name.up.sql` и `.down.sql`, затем выполните `docker compose up migrate`.

## Конфигурация

Корневой `.env` читает docker compose (см. [`.env.example`](.env.example)):

| Переменная          | По умолчанию            | Описание                                                     |
|---------------------|-------------------------|--------------------------------------------------------------|
| `WEB_PORT`          | `3300`                  | Порт сайта и MCP на хосте                                    |
| `WEB_BIND`          | `127.0.0.1`             | Интерфейс; `0.0.0.0` открывает доступ из локальной сети      |
| `PUBLIC_URL`        | `http://localhost:3300` | Базовый адрес для ссылок в ответах MCP                       |
| `MAX_UPLOAD_MB`     | `25`                    | Максимальный размер референса                                |
| `TREE_CACHE_TTL`    | `10m`                   | TTL кеша дерева в Redis                                      |
| `DEV_MODE`          | `false`                 | Подробные тексты внутренних ошибок в ответах                 |
| `POSTGRES_*`        | `workspace`             | Пользователь, пароль и имя БД                                |
| `POSTGRES_PORT` / `REDIS_PORT` | `55432` / `56379` | Порты для локальной разработки (только 127.0.0.1)        |

## Безопасность

Авторизации в сервисе нет: он рассчитан на локальное использование одним человеком или небольшой командой в доверенной сети. Поэтому порт по умолчанию слушает только `127.0.0.1`. Для доступа извне поставьте перед сервисом прокси с аутентификацией. MCP даёт полный доступ на запись и удаление.

Загруженные файлы отдаются с `Content-Security-Policy: sandbox` и `X-Content-Type-Options: nosniff`, а тип определяется по содержимому. Поэтому загруженный HTML или SVG со скриптом не выполнится в контексте сайта.
