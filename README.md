# Demo Service

Demo Service — микросервис для просмотра заказов, реализующий работу с PostgreSQL, Kafka и HTTP API.  
Проект поддерживает кэширование заказов, интеграцию с Kafka, Swagger-документацию и простое веб-приложение для просмотра заказов.

---

## Состав репозитория

- **cmd/main.go** — точка входа, запуск через Fx DI.
- **internal/**
  - **app/** — модели данных (Order, Delivery, Payment, Item).
  - **cache/** — LRU-кэш заказов.
  - **config/** — загрузка конфигурации из YAML.
  - **db/** — работа с PostgreSQL (CRUD, кэш-загрузка).
  - **di/** — DI-компоненты для Fx.
  - **kafka/** — консьюмер Kafka.
  - **web/** — HTTP-обработчики и роутер.
- **config/local.yaml** — пример конфигурации.
- **migrations/** — SQL-миграции для PostgreSQL.
- **scripts/messages.go** — отправка тестовых сообщений в Kafka.
- **docs/** — Swagger-документация.
- **web/index.html** — простая страница для просмотра заказов.
- **docker-compose.yml** — запуск PostgreSQL и Kafka через Docker.

---

## Быстрый старт

### 1. Запуск инфраструктуры

```sh
docker-compose up -d
```
Запустит PostgreSQL (порт 5433), Kafka (порт 9092), Zookeeper.

### 2. Применение миграций

Примените миграции из папки `migrations/` к базе данных `demo_service` (например, через [migrate](https://github.com/golang-migrate/migrate)):

```sh
migrate -path migrations -database "postgres://user:password@localhost:5433/demo_service?sslmode=disable" up
```

### 3. Настройка конфигурации

Проверьте файл [`config/local.yaml`](config/local.yaml):

```yaml
env: local
http_port: 8080
db_name: demo_service
db_user: user
db_password: password
db_url: localhost
db_port: 5433
cache_size: 15
```

Укажите путь к конфигу через переменную окружения:

```sh
export CONFIG_PATH=config/local.yaml
```

### 4. Запуск сервиса

```sh
go run ./cmd/main.go
```

Сервис стартует на порту 8080.

---

## API

- **GET /order/{orderUID}** — получить заказ по UID (из кэша или базы).
- **Swagger**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

---

## Веб-интерфейс

Откройте [`web/index.html`](web/index.html) в браузере.  
Введите Order UID и получите заказ через API.

---

## Тесты

- Юнит-тесты: `go test ./internal/...`
- Интеграционные тесты для PostgreSQL: `go test ./internal/db/tests`
  - Перед запуском задайте переменные окружения для тестовой БД.

---

## Kafka

Для отправки тестовых сообщений используйте [`scripts/messages.go`](scripts/messages.go):

```sh
go run scripts/messages.go
```

---

## Миграции

- `migrations/000001_create_tables.up.sql` — создание таблиц.
- `migrations/000001_create_tables.down.sql` — удаление таблиц.

---

## Зависимости

- Go 1.20+
- PostgreSQL 16+
- Kafka 7.5+
- Docker (для инфраструктуры)
- Swagger (для документации)

---

## Swagger

- Swagger: [docs/swagger.yaml](docs/swagger.yaml)
- Документация генерируется автоматически и доступна по `/swagger/*`.

---

## Профилирование и оптимизация

### Инструменты
- `net/http/pprof` — сбор профилей CPU и памяти
- `go tool pprof` — анализ профилей
- `go test -bench` — бенчмарки
- `benchstat` — сравнение результатов
- `go tool trace` — трассировка выполнения

### Что было сделано
1. Изменение in-memory cache. В кэше теперь хранится не экземпляр Order. А сериализованный JSON с этой структурой.
2. Использование `sync.Pool` для буферов. Раньше encoder json писал напрямую в ResponseWriter, что создавало нагрузку на CPU большим количеством запросов. Теперь все пишется в буффер, и разом постится в Response.

### Результаты BenchStat

- goos: darwin
- goarch: arm64
- pkg: demoservice/internal/web
- cpu: Apple M1

|                      | bench_before.txt<br/>sec/op  | bench_after.txt<br/>sec/op | diff                  |
|----------------------|------------------------------|----------------------------|-----------------------|
| GetOrder_CacheHit-8  | 1576.0n ± ∞ ¹                | 950.7n ± ∞ ¹               | -39.68% (p=0.008 n=5) |
| GetOrder_CacheMiss-8 | 1514.0n ± ∞ ¹                | 869.0n ± ∞ ¹               | -42.60% (p=0.008 n=5) |
| geomean              | 1.545µ                       | 908.9n                     | -41.16%               |

|                      | bench_before.txt<br/>B/op    | bench_after.txt<br/>B/op | diff                  |
|----------------------|------------------------------|----------------------------|-----------------------|
| GetOrder_CacheHit-8  | 2.126Ki ± ∞ ¹                | 2.112Ki ± ∞ ¹               | -0.64% (p=0.008 n=5)  |
| GetOrder_CacheMiss-8 | 2.126Ki ± ∞ ¹                | 2.112Ki ± ∞ ¹               | -0.64% (p=0.008 n=5)  |
| geomean              | 2.126Ki                      | 2.112Ki                     | -0.64%                |

|                      | bench_before.txt<br/>allocs/op  | bench_after.txt<br/>allocs/op | diff                 |
|----------------------|------------------------------|----------------------------|----------------------|
| GetOrder_CacheHit-8  | 14.00 ± ∞ ¹                | 15.00 ± ∞ ¹               | +7.14% (p=0.008 n=5) |
| GetOrder_CacheMiss-8 | 14.00 ± ∞ ¹                | 15.00 ± ∞ ¹               | +7.14% (p=0.008 n=5) |
| geomean              | 14.00                      | 15.00                     | +7.14%               |

- Скорость выполнения увеличилась на 41% при увеличении расхода памяти всего на 7%.
