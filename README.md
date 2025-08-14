## WB Orders — сервис заказов с Kafka, PostgreSQL, Echo и кэшом в памяти

Микросервис для приёма и чтения заказов:
- HTTP API на Echo (`/order`, `/order/{id}`, CRUD и `/health`).
- Чтение событий заказов из Kafka и сохранение в PostgreSQL.
- In‑memory кэш заказов с TTL, прогрев из БД при старте.
- OpenAPI спецификация и встроенный Swagger UI.

### Технологии
- **Go** 1.24
- **Echo** (HTTP сервер, CORS, middleware)
- **GORM** + **PostgreSQL** (миграции через AutoMigrate)
- **Kafka (Sarama)** — consumer и тестовый producer
- **cleanenv**, **godotenv** — конфигурация
- **slog** — логирование

## Быстрый старт

### Вариант 1 — через Docker Compose (рекомендуется)
Запустит PostgreSQL, Zookeeper, Kafka, Kafka‑UI и API‑сервис.

```bash
docker-compose up -d
```

- API: `http://localhost:8081`
- Swagger UI: `http://localhost:8081/swagger`
- OpenAPI YAML: `http://localhost:8081/openapi.yaml`
- Kafka UI: `http://localhost:8080`

Остановка:

```bash
docker-compose down -v
```

Локальная Make‑цель:

```bash
make run_full
```

### Вариант 2 — локальный запуск (без Docker для API)
Инфраструктуру (Postgres, Kafka) можно поднять из compose, а API запускать локально.

1) Поднять Postgres и Kafka (из compose) или иным способом.


go run cmd/api/main.go
```

Короткая цель для локального запуска:

```bash
make run
```

## Конфигурация

Сервис читает конфигурацию из `CONFIG_PATH` (YAML) и ряд параметров из env:

```yaml
env: dev
http_server:
  port: "8081"
  timeout: 5s
  idle_timeout: 60s
kafka:
  brokers:
    - "kafka:9092"
  topic: "orders"
  group_id: "wb-consumer"
  version: "2.8.0"
cache:
  ttl: 10m
```

Обязательные env:

| Переменная | Назначение | Пример |
|---|---|---|
| `CONFIG_PATH` | Путь к YAML‑конфигу | `/app/config/config.yaml` |
| `DB_CONNECTION_STRING` | Строка подключения к Postgres | `postgres://user:pass@host:5432/db?sslmode=disable` |

Дополнительные env:

| Переменная | Назначение | Пример |
|---|---|---|
| `KAFKA_BROKERS` | Переопределение брокеров из конфига | `localhost:9092` или `kafka:9092` |
| `KAFKA_TOPIC` | Топик для тестового producer`а (cmd/producer)` | `orders` |

Примечания:
- Миграции выполняются автоматически через GORM `AutoMigrate` при старте API.
- В БД создаётся уникальный индекс `orders(order_uid)` для идемпотентности.

## Архитектура

Структура каталогов:

```
api/openapi.yaml          # OpenAPI спецификация
cmd/api/main.go           # Точка входа HTTP API + запуск Kafka consumer
cmd/producer/main.go      # Пример producer'а: публикует тестовый заказ в Kafka
config/config.yaml        # Конфигурация по умолчанию (используется в compose)
internal/cache/           # In-memory кэш заказов (TTL, cleaner)
internal/config/          # Загрузка конфигурации из YAML/env
internal/handler/         # HTTP‑обработчики (CRUD заказов)
internal/kafka/           # Consumer и утилита‑producer
internal/lib/logger/      # Настройка slog
internal/models/          # GORM‑модели
internal/server/          # Echo server и маршруты, Swagger UI
internal/storage/postgres # Инициализация GORM, методы доступа к БД
Dockerfile                # Сборка API образа
docker-compose.yml        # Инфраструктура и API
Makefile                  # Утилитарные команды
```

Потоки данных:
- HTTP: запросы попадают в `internal/server/routes.go` → `internal/handler/*` → БД через `internal/storage/postgres` и/или кэш `internal/cache`.
- Kafka: `internal/kafka/consumer.go` читает сообщения, валидирует, сохраняет в БД и кладёт в кэш.

Кэш:
- Прогревается содержимым БД при старте.
- TTL управляется `cache.ttl` и фоновой очисткой.

## API

- Health: `GET /health`
- Заказы:
  - `GET /order` — список (кэш → БД при необходимости)
  - `GET /order/{id}` — по `order_uid`
  - `POST /order` — создать
  - `PUT /order` — обновить
  - `DELETE /order` — удалить (тело: `{ "order_uid": "..." }`)

Swagger UI: `http://localhost:8081/swagger`

Примеры:

```bash
curl http://localhost:8081/health

curl http://localhost:8081/order

curl http://localhost:8081/order/123e4567-e89b-12d3-a456-426614174000

curl -X POST http://localhost:8081/order \
  -H "Content-Type: application/json" \
  -d '{
    "track_number":"WBILMTESTTRACK",
    "entry":"WB",
    "delivery":{"name":"John Doe","phone":"+70000000000","zip":"123456","city":"Moscow","address":"Tverskaya 1","region":"Moscow","email":"john@example.com"},
    "payment":{"transaction":"tx-1","currency":"RUB","provider":"wbpay","amount":1000,"payment_dt":1710000000,"bank":"SBER","delivery_cost":100,"goods_total":900,"custom_fee":0},
    "items":[{"chrt_id":1,"track_number":"WBILMTESTTRACK","price":900,"rid":"rid-1","name":"Item","sale":0,"size":"M","total_price":900,"nm_id":1000,"brand":"WB","status":202}],
    "locale":"ru","customer_id":"cust-1","delivery_service":"meest","shardkey":"9","sm_id":99,"oof_shard":"1"
  }'
```

## Kafka: consumer и тестовый producer

Consumer запускается вместе с API и читает топик, указанный в конфигурации (`kafka.topic`). Версия брокера задаётся `kafka.version`.

Тестовый producer можно запустить так:

```bash
go run cmd/producer/main.go
```

Опционально задать окружение:

```bash
# один или несколько брокеров через запятую
export KAFKA_BROKERS=localhost:9092
export KAFKA_TOPIC=orders
```

Producer публикует JSON, соответствующий `OrderResponse`, сгенерирует случайный `order_uid` и текущие метки времени.

## Тонкости и заметки

- Для старта API обязательны `CONFIG_PATH` и `DB_CONNECTION_STRING`. При их отсутствии приложение завершится с ошибкой.
- В compose эти переменные уже заданы для контейнера `api`.
- `init.sql`: миграции БД выполняются автоматически через GORM, отдельный SQL не обязателен.
- Создаётся уникальный индекс `orders(order_uid)` для идемпотентности сохранения заказов.

## Разработка

- Локально включено детальное логирование (`env: dev`). В конфиге можно переключить уровни логов.
- Формат кода, именование и структура соответствуют привычным практикам Go; хранение бизнес‑логики — в `internal/*`.



