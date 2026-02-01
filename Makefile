.PHONY: help all backend frontend db db-stop clean install test test-backend test-frontend test-coverage test-watch test-unit test-integration benchmark migrate-up migrate-down migrate-create migrate-version migrate-force docs-preview

# Load environment variables from .env
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sed 's/Makefile://' | sort | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

all: db ## Start database (run 'make backend' and 'make frontend' in separate terminals)
	@echo "Database is running. In separate terminals, run:"
	@echo "  make backend"
	@echo "  make frontend"

db: ## Start PostgreSQL database
	@echo "Starting PostgreSQL..."
	@docker compose up -d db
	@echo "Waiting for database to be ready..."
	@until docker exec nomdb-db pg_isready -U nomdb -d nomdb > /dev/null 2>&1; do sleep 1; done
	@echo "Database ready at postgres://nomdb:nomdb_secret@localhost:5432/nomdb"

db-stop: ## Stop PostgreSQL database
	@docker compose stop db

backend: ## Run backend with Go
	@cd backend && go run ./cmd/server

frontend: install ## Run frontend with Vite dev server
	@cd frontend && npm run dev

install: ## Install frontend dependencies
	@cd frontend && npm install

clean: ## Stop and remove database container and volume
	@docker compose down -v

test: test-backend test-frontend ## Run all tests (backend and frontend)

test-backend: ## Run backend tests
	@echo "Running backend tests..."
	@cd backend && go test -v ./...

test-frontend: install ## Run frontend tests
	@echo "Running frontend tests..."
	@cd frontend && npm test

test-coverage: ## Run tests with coverage report
	@echo "Running backend tests with coverage..."
	@cd backend && go test -v -cover -coverprofile=coverage.out ./...
	@cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "Running frontend tests with coverage..."
	@cd frontend && npm run test:coverage
	@echo "Coverage reports generated:"
	@echo "  Backend: backend/coverage.html"
	@echo "  Frontend: frontend/coverage/index.html"

test-watch: ## Run frontend tests in watch mode
	@echo "Running frontend tests in watch mode..."
	@cd frontend && npm run test:watch

test-unit: ## Run unit tests only (backend)
	@echo "Running backend unit tests..."
	@cd backend && go test -v -short ./internal/middleware/... ./internal/services/...

test-integration: ## Run integration tests only (requires database)
	@echo "Running integration tests..."
	@cd backend && go test -v -run Integration ./...

benchmark: ## Run backend benchmarks
	@echo "Running backend benchmarks..."
	@cd backend && go test -bench=. -benchmem ./internal/middleware/...

migrate-up: ## Run database migrations
	@echo "Running database migrations..."
	@cd backend && go run cmd/migrate/main.go up

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	@cd backend && go run cmd/migrate/main.go down

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(NAME)"
	@cd backend && go run cmd/migrate/main.go create $(NAME)

migrate-version: ## Show current migration version
	@cd backend && go run cmd/migrate/main.go version

migrate-force: ## Force migration version (usage: make migrate-force VERSION=4)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make migrate-force VERSION=4"; \
		exit 1; \
	fi
	@echo "Forcing migration version to $(VERSION)..."
	@cd backend && go run cmd/migrate/main.go force $(VERSION)

docs-preview: ## Run documentation site locally (Hugo + Lotus Docs)
	@cd docs && hugo server -D -e production --baseURL http://localhost:1313/
