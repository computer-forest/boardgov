.PHONY: build run test lint migrate docker

APP=boardgov
CMD=./cmd/server
BUILD_FLAGS=-ldflags="-w -s"

build:
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -o bin/$(APP) $(CMD)

run:
	go run $(CMD)/main.go

test:
	go test ./... -v -race -timeout 60s

lint:
	golangci-lint run ./...

migrate:
	psql $${DATABASE_DSN} -f migrations/001_initial_schema.sql

docker:
	docker build -t $(APP):latest .

docker-run:
	docker run --env-file .env -p 8080:8080 $(APP):latest
