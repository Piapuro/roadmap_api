.PHONY: run build dev sqlc migrate

run:
	go run main.go

build:
	go build -o bin/api main.go

dev:
	air

sqlc:
	sqlc generate

migrate:
	psql $(DATABASE_URL) -f sql/schema.sql
