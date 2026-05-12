# Makefile for go-genpos-app

# Load environment variables from .env
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Database Connection String for golang-migrate
# Constructed from individual DB variables in .env
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

MIGRATION_PATH=./migrations
MIGRATE=migrate

.PHONY: help migrate-create migrate-up migrate-down migrate-force migrate-status migrate-drop migrate-fresh

help: ## Show this help menu
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

migrate-create: ## Create a new migration file. Usage: make migrate-create name=init_schema
	@echo "Creating migration: $(name)"
	@$(MIGRATE) create -ext sql -dir $(MIGRATION_PATH) -seq $(name)

migrate-up: ## Run all up migrations
	@echo "Running up migrations..."
	@$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DB_URL)" up

migrate-down: ## Rollback the last migration
	@echo "Running down migrations..."
	@$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DB_URL)" down 1

migrate-force: ## Force migration to a specific version. Usage: make migrate-force version=N
	@echo "Force-setting version to $(version)..."
	@$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DB_URL)" force $(version)

migrate-status: ## Show migration status
	@$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DB_URL)" version

migrate-drop: ## Drop all tables in the database (CAUTION)
	@echo "Dropping all tables..."
	@$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DB_URL)" drop -f

migrate-fresh: migrate-drop migrate-up db-seed ## Drop all tables, run migrations and seed

db-seed: ## Seed the database with initial data
	@echo "Seeding database..."
	@go run cmd/seed/main.go

api-bundle: ## Bundle OpenAPI files into a single YAML
	@echo "Bundling OpenAPI documentation..."
	@npx @redocly/cli bundle docs/api/openapi/index.yaml -o docs/api/openapi.bundle.yaml

api-lint: ## Lint OpenAPI files
	@echo "Linting OpenAPI documentation..."
	@npx @redocly/cli lint docs/api/openapi/index.yaml
