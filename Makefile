.PHONY: run build tidy test test-integration test-all coverage migrate-up diagrams docker-up docker-up-postgres frontend-install frontend-dev frontend-build sync-swagger

run: sync-swagger
	go run ./cmd/server

sync-swagger:
	cp docs/swagger.yaml internal/swagger/openapi.yaml

build: sync-swagger
	go build -o bin/server ./cmd/server

tidy:
	go mod tidy

test:
	go test ./...

test-integration:
	go test -tags=integration ./...

test-all: test test-integration

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

migrate-up:
	flyway -url=jdbc:postgresql://localhost:5432/loggerdb -user=postgres -password=postgres -locations=filesystem:database/migrations migrate

diagrams:
	docker run --rm -v "$(PWD)/docs/diagrams:/data" plantuml/plantuml -tpng -o generated /data/*.puml

docker-up:
	docker compose up --build

docker-up-postgres:
	docker compose -f docker-compose.yml -f docker-compose.postgres.yml up --build

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build
