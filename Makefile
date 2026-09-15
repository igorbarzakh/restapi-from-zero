.PHONY: up down logs test run

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f api

test:
	go test ./...

run:
	@set -a; \
	if [ -f .env ]; then . ./.env; fi; \
	export PGDATABASE="$${POSTGRES_DB:-tasks}"; \
	export PGUSER="$${POSTGRES_USER:-tasks}"; \
	export PGPASSWORD="$${POSTGRES_PASSWORD:-local_dev_password}"; \
	export DATABASE_URL="$${DATABASE_URL:-host=localhost port=5432 sslmode=disable}"; \
	export SERVER_PORT="$${SERVER_PORT:-8080}"; \
	exec go run ./cmd/api
