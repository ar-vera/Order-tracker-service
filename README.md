# Order-tracker-service

Демо‑сервис отслеживания заказов на PostgreSQL, Kafka (Confluent), HTTP API (Gin) и простой веб‑интерфейс.

## Быстрый старт

Требования:
- Docker + Docker Compose
- Make

Запуск всех сервисов:
```bash
make up
```

Открыть UI:
- http://localhost:8080/

Полезные команды:
```bash
make logs         # логи всех сервисов (follow)
make logs-app     # логи приложения
make logs-producer
make ps           # статус контейнеров
make sh-app       # shell в контейнер приложения
make sh-db        # psql в контейнер Postgres
make migrate      # ручной прогон SQL миграции (обычно не нужен — запускается при старте app)
```

## Что запущено

- Postgres (БД orders) — на хосте порт 5433 (в контейнере 5432)
- Zookeeper + Kafka
  - адрес брокера для контейнеров: `kafka:9092`
  - доступ с хоста (CLI): `localhost:29092`
- Приложение (HTTP + Kafka consumer)
- Продьюсер (периодически отправляет случайные заказы в Kafka)

## API

- Проверка здоровья: `GET /api/v1/health`
- Получить заказ по UID: `GET /api/v1/orders/{id}`
- Получить все (заглушка): `GET /api/v1/orders?page=1&limit=10`

## Веб‑интерфейс

- Корневая страница `GET /` содержит поле для ввода Order UID и кнопку «Загрузить». Результат отображается в удобном формате.

## База данных

Авто‑миграции применяются при старте приложения из папки `migrations/`.

Ручные проверки:
```bash
make sh-db
\dt
SELECT order_uid FROM orders ORDER BY id DESC LIMIT 5;
```

## Диагностика проблем

- Вижу только JSON — это нормально для API‑маршрутов; для HTML откройте `http://localhost:8080/`.
- Ошибка «relation orders does not exist» — убедитесь, что приложение запустилось после Postgres и применило миграции; при необходимости выполните `make migrate`.
- Подключение к Kafka с хоста — используйте `localhost:29092`; внутри контейнеров — `kafka:9092`.