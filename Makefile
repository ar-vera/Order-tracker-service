SHELL := /bin/zsh

.PHONY: up down restart logs logs-app logs-producer ps sh-app sh-producer sh-db seed migrate

up:
	docker compose up -d --build

down:
	docker compose down -v

restart:
	$(MAKE) down && $(MAKE) up

logs:
	docker compose logs -f

logs-app:
	docker compose logs -f app

logs-producer:
	docker compose logs -f producer

ps:
	docker compose ps

sh-app:
	docker exec -it orders_app sh

sh-producer:
	docker exec -it orders_producer sh

sh-db:
	docker exec -it orders_postgres psql -U postgres -d orders

migrate:
	# Миграции запускаются автоматически при старте app, цель для ручного прогона SQL (если нужно)
	docker exec -i orders_postgres psql -U postgres -d orders < migrations/000001_create_tables.up.sql

seed:
	# Проверить последние 5 заказов в БД
	docker exec -it orders_postgres psql -U postgres -d orders -c "SELECT order_uid, date_created FROM orders ORDER BY id DESC LIMIT 5;"


