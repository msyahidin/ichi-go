.PHONY: ichigen-build ichigen-install ichigen-clean ichigen-test ichigen-example migration-help

BINARY_NAME=ichigen
INSTALL_PATH=./

ichigen-build:
	@go build -o $(BINARY_NAME) pkg/ichigen/cmd/main.go
	@echo "✓ Built $(BINARY_NAME)"

ichigen-install: ichigen-build
	@sudo mv $(BINARY_NAME) $(INSTALL_PATH)/
	@echo "✓ Installed to $(INSTALL_PATH)"

ichigen-clean:
	@rm -f $(BINARY_NAME)
	@echo "✓ Cleaned"

ichigen-test:
	@go test ./internal/generator/...

# Examples
ichigen-example-full:
	@go run ./pkg/ichigen/cmd/main.go g full product --domain=catalog --crud

ichigen-example-controller:
	@go run ./pkg/ichigen/cmd/main.go g controller notification --domain=alert

# Colors for output
COLOR_RESET   = \033[0m
COLOR_INFO    = \033[36m
COLOR_SUCCESS = \033[32m
COLOR_WARNING = \033[33m
COLOR_ERROR   = \033[31m

##@ Database Migration Commands

migration-help: ## Display this help message
	@echo "$(COLOR_INFO)📚 Database Migration Manager$(COLOR_RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(COLOR_INFO)<target>$(COLOR_RESET)\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(COLOR_INFO)%-20s$(COLOR_RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(COLOR_SUCCESS)%s$(COLOR_RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Schema Migrations

migration-create: ## Create new schema migration (usage: make migration-create name=create_products_table)
	@if [ -z "$(name)" ]; then \
		echo "$(COLOR_ERROR)❌ Error: name parameter is required$(COLOR_RESET)"; \
		echo "$(COLOR_INFO)Usage: make migration-create name=create_products_table$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_INFO)📝 Creating schema migration: $(name)$(COLOR_RESET)"
	@go run db/cmd/migrate.go create $(name) sql
	@echo "$(COLOR_SUCCESS)✅ Migration created successfully$(COLOR_RESET)"

migration-up: ## Run all pending schema migrations
	@echo "$(COLOR_INFO)⬆️  Running schema migrations...$(COLOR_RESET)"
	@go run db/cmd/migrate.go up
	@echo "$(COLOR_SUCCESS)✅ Migrations completed$(COLOR_RESET)"

migration-down: ## Rollback last schema migration
	@echo "$(COLOR_WARNING)⬇️  Rolling back last migration...$(COLOR_RESET)"
	@go run db/cmd/migrate.go down
	@echo "$(COLOR_SUCCESS)✅ Rollback completed$(COLOR_RESET)"

migration-status: ## Show migration status
	@echo "$(COLOR_INFO)📊 Migration status:$(COLOR_RESET)"
	@go run db/cmd/migrate.go status

migration-redo: ## Re-run last migration (down + up)
	@echo "$(COLOR_INFO)🔄 Re-running last migration...$(COLOR_RESET)"
	@go run db/cmd/migrate.go redo
	@echo "$(COLOR_SUCCESS)✅ Redo completed$(COLOR_RESET)"

migration-reset: ## Rollback ALL migrations (DANGEROUS!)
	@echo "$(COLOR_ERROR)⚠️  WARNING: This will rollback ALL migrations!$(COLOR_RESET)"
	@echo "$(COLOR_WARNING)Press Ctrl+C to cancel, or wait 5 seconds to continue...$(COLOR_RESET)"
	@sleep 5
	@go run db/cmd/migrate.go reset
	@echo "$(COLOR_SUCCESS)✅ Reset completed$(COLOR_RESET)"

migration-up-to: ## Migrate to specific version (usage: make migration-up-to version=20250407085044)
	@if [ -z "$(version)" ]; then \
		echo "$(COLOR_ERROR)❌ Error: version parameter is required$(COLOR_RESET)"; \
		echo "$(COLOR_INFO)Usage: make migration-up-to version=20250407085044$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_INFO)⬆️  Migrating to version $(version)...$(COLOR_RESET)"
	@go run db/cmd/migrate.go up-to $(version)
	@echo "$(COLOR_SUCCESS)✅ Migration completed$(COLOR_RESET)"

migration-down-to: ## Rollback to specific version (usage: make migration-down-to version=20250407085044)
	@if [ -z "$(version)" ]; then \
		echo "$(COLOR_ERROR)❌ Error: version parameter is required$(COLOR_RESET)"; \
		echo "$(COLOR_INFO)Usage: make migration-down-to version=20250407085044$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_WARNING)⬇️  Rolling back to version $(version)...$(COLOR_RESET)"
	@go run db/cmd/migrate.go down-to $(version)
	@echo "$(COLOR_SUCCESS)✅ Rollback completed$(COLOR_RESET)"

##@ Data Migrations

data-migration-create: ## Create new data migration (usage: make data-migration-create name=fix_legacy_emails)
	@if [ -z "$(name)" ]; then \
		echo "$(COLOR_ERROR)❌ Error: name parameter is required$(COLOR_RESET)"; \
		echo "$(COLOR_INFO)Usage: make data-migration-create name=fix_legacy_emails$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_INFO)📝 Creating data migration: $(name)$(COLOR_RESET)"
	@go run db/cmd/migrate.go --table=data create $(name) sql
	@echo "$(COLOR_SUCCESS)✅ Data migration created successfully$(COLOR_RESET)"

data-migration-up: ## Run all pending data migrations
	@echo "$(COLOR_INFO)⬆️  Running data migrations...$(COLOR_RESET)"
	@go run db/cmd/migrate.go --table=data up
	@echo "$(COLOR_SUCCESS)✅ Data migrations completed$(COLOR_RESET)"

data-migration-down: ## Rollback last data migration
	@echo "$(COLOR_WARNING)⬇️  Rolling back last data migration...$(COLOR_RESET)"
	@go run db/cmd/migrate.go --table=data down
	@echo "$(COLOR_SUCCESS)✅ Rollback completed$(COLOR_RESET)"

data-migration-status: ## Show data migration status
	@echo "$(COLOR_INFO)📊 Data migration status:$(COLOR_RESET)"
	@go run db/cmd/migrate.go --table=data status

##@ Database Seeders

seed-run: ## Run all seed files
	@echo "$(COLOR_INFO)🌱 Running all seeders...$(COLOR_RESET)"
	@go run db/cmd/migrate.go seed run
	@echo "$(COLOR_SUCCESS)✅ All seeds completed$(COLOR_RESET)"

seed-file: ## Run specific seed file (usage: make seed-file name=00_base_roles.sql)
	@if [ -z "$(name)" ]; then \
		echo "$(COLOR_ERROR)❌ Error: name parameter is required$(COLOR_RESET)"; \
		echo "$(COLOR_INFO)Usage: make seed-file name=00_base_roles.sql$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_INFO)🌱 Running seed: $(name)$(COLOR_RESET)"
	@go run db/cmd/migrate.go seed run $(name)
	@echo "$(COLOR_SUCCESS)✅ Seed completed$(COLOR_RESET)"

seed-list: ## List all available seed files
	@echo "$(COLOR_INFO)📋 Available seed files:$(COLOR_RESET)"
	@go run db/cmd/migrate.go seed list

##@ Database Management

db-reset-dev: ## Reset database for development (reset + migrate + seed)
	@echo "$(COLOR_WARNING)🔄 Resetting development database...$(COLOR_RESET)"
	@echo "$(COLOR_WARNING)Press Ctrl+C to cancel, or wait 3 seconds to continue...$(COLOR_RESET)"
	@sleep 3
	@$(MAKE) migration-reset
	@$(MAKE) migration-up
	@$(MAKE) data-migration-up
	@$(MAKE) seed-run
	@echo "$(COLOR_SUCCESS)✅ Development database reset completed$(COLOR_RESET)"

db-fresh: ## Fresh database setup (up + seed)
	@echo "$(COLOR_INFO)🆕 Setting up fresh database...$(COLOR_RESET)"
	@$(MAKE) migration-up
	@$(MAKE) data-migration-up
	@$(MAKE) seed-run
	@echo "$(COLOR_SUCCESS)✅ Fresh database setup completed$(COLOR_RESET)"

db-status: ## Show complete database status
	@echo "$(COLOR_INFO)📊 Complete Database Status$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_INFO)Schema Migrations:$(COLOR_RESET)"
	@$(MAKE) migration-status
	@echo ""
	@echo "$(COLOR_INFO)Data Migrations:$(COLOR_RESET)"
	@$(MAKE) data-migration-status
	@echo ""
	@echo "$(COLOR_INFO)Available Seeds:$(COLOR_RESET)"
	@$(MAKE) seed-list

##@ Swagger Documentation

swagger-init: ## Initialize Swagger documentation
	@echo "$(COLOR_INFO)📝 Initializing Swagger documentation...$(COLOR_RESET)"
	@swag init -g cmd/main.go -o docs --parseDependency --parseInternal
	@echo "$(COLOR_SUCCESS)✅ Swagger documentation generated$(COLOR_RESET)"
	@echo "$(COLOR_INFO)📍 Access at: http://localhost:8080/swagger/index.html$(COLOR_RESET)"

swagger-fmt: ## Format Swagger comments
	@echo "$(COLOR_INFO)🔧 Formatting Swagger comments...$(COLOR_RESET)"
	@swag fmt
	@echo "$(COLOR_SUCCESS)✅ Swagger comments formatted$(COLOR_RESET)"

swagger-gen: swagger-fmt swagger-init ## Format and generate Swagger docs
	@echo "$(COLOR_SUCCESS)✅ Swagger documentation updated$(COLOR_RESET)"

swagger-validate: ## Validate Swagger documentation
	@echo "$(COLOR_INFO)🔍 Validating Swagger spec...$(COLOR_RESET)"
	@if [ -f "docs/swagger.json" ]; then \
		echo "$(COLOR_SUCCESS)✅ Swagger spec exists$(COLOR_RESET)"; \
		echo "$(COLOR_INFO)📄 JSON spec: docs/swagger.json$(COLOR_RESET)"; \
		echo "$(COLOR_INFO)📄 YAML spec: docs/swagger.yaml$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_ERROR)❌ Swagger spec not found. Run 'make swagger-init'$(COLOR_RESET)"; \
		exit 1; \
	fi

swagger-clean: ## Clean generated Swagger files
	@echo "$(COLOR_WARNING)🗑️  Cleaning Swagger documentation...$(COLOR_RESET)"
	@rm -rf docs/docs.go docs/swagger.json docs/swagger.yaml
	@echo "$(COLOR_SUCCESS)✅ Swagger documentation cleaned$(COLOR_RESET)"

swagger-help: ## Show Swagger usage help
	@echo "$(COLOR_INFO)╔════════════════════════════════════════════════╗$(COLOR_RESET)"
	@echo "$(COLOR_INFO)║     Swagger Documentation Commands            ║$(COLOR_RESET)"
	@echo "$(COLOR_INFO)╚════════════════════════════════════════════════╝$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_SUCCESS)📚 Available Commands:$(COLOR_RESET)"
	@echo "  $(COLOR_INFO)make swagger-init$(COLOR_RESET)       - Generate Swagger documentation"
	@echo "  $(COLOR_INFO)make swagger-fmt$(COLOR_RESET)        - Format Swagger comments"
	@echo "  $(COLOR_INFO)make swagger-gen$(COLOR_RESET)        - Format and generate docs (recommended)"
	@echo "  $(COLOR_INFO)make swagger-validate$(COLOR_RESET)   - Validate Swagger spec"
	@echo "  $(COLOR_INFO)make swagger-clean$(COLOR_RESET)      - Remove generated files"
	@echo "  $(COLOR_INFO)make swagger-help$(COLOR_RESET)       - Show this help message"
	@echo ""
	@echo "$(COLOR_SUCCESS)🚀 Quick Start:$(COLOR_RESET)"
	@echo "  1. Generate docs:   $(COLOR_INFO)make swagger-gen$(COLOR_RESET)"
	@echo "  2. Start server:    $(COLOR_INFO)go run cmd/main.go$(COLOR_RESET) or $(COLOR_INFO)air$(COLOR_RESET)"
	@echo "  3. Visit:           $(COLOR_INFO)http://localhost:8080/swagger/index.html$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_SUCCESS)💡 Tips:$(COLOR_RESET)"
	@echo "  • Always run $(COLOR_INFO)make swagger-gen$(COLOR_RESET) after changing controllers"
	@echo "  • Swagger UI lets you test endpoints interactively"
	@echo "  • Use the 'Authorize' button to add Bearer token"
	@echo ""

##@ Application

run: ## Run the application
	@echo "$(COLOR_INFO)🚀 Starting application...$(COLOR_RESET)"
	@go run cmd/main.go

run-dev: ## Run the application with hot reload (requires air)
	@echo "$(COLOR_INFO)🔥 Starting application with hot reload...$(COLOR_RESET)"
	@air

build: ## Build the application
	@echo "$(COLOR_INFO)🔨 Building application...$(COLOR_RESET)"
	@go build -o bin/ichi-go cmd/main.go
	@echo "$(COLOR_SUCCESS)✅ Build completed: bin/ichi-go$(COLOR_RESET)"

test: ## Run tests
	@echo "$(COLOR_INFO)🧪 Running tests...$(COLOR_RESET)"
	@go test ./... -v

test-coverage: ## Run tests with coverage
	@echo "$(COLOR_INFO)🧪 Running tests with coverage...$(COLOR_RESET)"
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_SUCCESS)✅ Coverage report: coverage.html$(COLOR_RESET)"

lint: ## Run linter
	@echo "$(COLOR_INFO)🔍 Running linter...$(COLOR_RESET)"
	@golangci-lint run

fmt: ## Format code
	@echo "$(COLOR_INFO)💅 Formatting code...$(COLOR_RESET)"
	@go fmt ./...
	@echo "$(COLOR_SUCCESS)✅ Code formatted$(COLOR_RESET)"

clean: ## Clean build artifacts and generated files
	@echo "$(COLOR_WARNING)🧹 Cleaning...$(COLOR_RESET)"
	@rm -rf bin/
	@rm -rf coverage.out coverage.html
	@echo "$(COLOR_SUCCESS)✅ Clean completed$(COLOR_RESET)"

##@ Docker

docker-build: ## Build Docker image
	@echo "$(COLOR_INFO)🐳 Building Docker image...$(COLOR_RESET)"
	@docker build -t ichi-go:latest .
	@echo "$(COLOR_SUCCESS)✅ Docker image built$(COLOR_RESET)"

docker-run: ## Run Docker container
	@echo "$(COLOR_INFO)🐳 Running Docker container...$(COLOR_RESET)"
	@docker run -p 8080:8080 ichi-go:latest

docker-compose-up: ## Start Docker Compose services
	@echo "$(COLOR_INFO)🐳 Starting Docker Compose services...$(COLOR_RESET)"
	@docker-compose up -d
	@echo "$(COLOR_SUCCESS)✅ Services started$(COLOR_RESET)"

docker-compose-down: ## Stop Docker Compose services
	@echo "$(COLOR_WARNING)🐳 Stopping Docker Compose services...$(COLOR_RESET)"
	@docker-compose down
	@echo "$(COLOR_SUCCESS)✅ Services stopped$(COLOR_RESET)"

##@ Help

help: ## Display this help message
	@echo "$(COLOR_INFO)╔════════════════════════════════════════════════╗$(COLOR_RESET)"
	@echo "$(COLOR_INFO)║        Ichi-Go Makefile Commands               ║$(COLOR_RESET)"
	@echo "$(COLOR_INFO)╚════════════════════════════════════════════════╝$(COLOR_RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(COLOR_INFO)<target>$(COLOR_RESET)\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(COLOR_INFO)%-25s$(COLOR_RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(COLOR_SUCCESS)%s$(COLOR_RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(COLOR_SUCCESS)💡 Quick Commands:$(COLOR_RESET)"
	@echo "  $(COLOR_INFO)make help$(COLOR_RESET)              - Show all commands"
	@echo "  $(COLOR_INFO)make swagger-help$(COLOR_RESET)     - Show Swagger commands"
	@echo "  $(COLOR_INFO)make migration-help$(COLOR_RESET)   - Show migration commands"
	@echo ""

.DEFAULT_GOAL := help