.PHONY: run worker migrate-up seed docker-up docker-down docker-migrate docker-seed docker-fresh test-docs test

run:
	go run cmd/api/main.go

worker:
	go run ./cmd/worker

migrate-up:
	go run cmd/migrate/main.go -direction=up -path=migration

seed:
	go run cmd/seed/main.go

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

test-docs:
	docker compose up -d --wait postgres
	env -u GOROOT GOTOOLCHAIN=auto GOCACHE=/tmp/booking-service-go-cache go test -count=1 -v ./test

test:
	env -u GOROOT GOTOOLCHAIN=auto GOCACHE=/tmp/booking-service-go-cache go test ./...
