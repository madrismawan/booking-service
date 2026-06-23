run:
	go run cmd/api/main.go

worker:
	go run ./cmd/worker

init: migrate-up seed

migrate-up:
	go run cmd/migrate/main.go -direction=up -path=migration

seed:
	go run cmd/seed/main.go

docker-build:
	docker compose build

docker-up:
	docker compose up --build

docker-down:
	docker compose down

docker-migrate:
	docker compose up -d --wait postgres
	docker compose run --rm api /app/migrate -direction=up -path=/app/migration

docker-seed:
	docker compose up -d --wait postgres
	docker compose run --rm api /app/seed

docker-fresh:
	docker compose up -d --wait postgres
	docker compose run --rm api /app/migrate -direction=fresh -path=/app/migration
	docker compose run --rm api /app/seed