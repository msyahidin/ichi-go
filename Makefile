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

#test: ## Run tests
#	@echo "$(COLOR_INFO)🧪 Running tests...$(COLOR_RESET)"
#	@go test ./... -v
#
#test-coverage: ## Run tests with coverage
#	@echo "$(COLOR_INFO)🧪 Running tests with coverage...$(COLOR_RESET)"
#	@go test ./... -coverprofile=coverage.out
#	@go tool cover -html=coverage.out -o coverage.html
#	@echo "$(COLOR_SUCCESS)✅ Coverage report: coverage.html$(COLOR_RESET)"

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

IMAGE_NAME := ichi-go
IMAGE_TAG := latest
COMPOSE_PROD := docker-compose.yml
COMPOSE_DEV := docker-compose.dev.yml

docker-build: ## Build Docker image
	@echo "$(COLOR_INFO)🐳 Building Docker image...$(COLOR_RESET)"
	@docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "$(COLOR_SUCCESS)✅ Docker image built: $(IMAGE_NAME):$(IMAGE_TAG)$(COLOR_RESET)"

docker-build-dev: ## Build development Docker image
	@echo "$(COLOR_INFO)🐳 Building development Docker image...$(COLOR_RESET)"
	@docker build -f Dockerfile.dev -t $(IMAGE_NAME):dev .
	@echo "$(COLOR_SUCCESS)✅ Development image built$(COLOR_RESET)"


docker-build-dev-no-cache: ## Build development Docker image
	@echo "$(COLOR_INFO)🐳 Building development Docker image...$(COLOR_RESET)"
	@docker build --no-cache -f Dockerfile.dev -t $(IMAGE_NAME):dev .
	@echo "$(COLOR_SUCCESS)✅ Development image built$(COLOR_RESET)"

docker-build-no-cache: ## Build Docker image without cache
	@echo "$(COLOR_INFO)🐳 Building Docker image (no cache)...$(COLOR_RESET)"
	@docker build --no-cache -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "$(COLOR_SUCCESS)✅ Docker image built$(COLOR_RESET)"

docker-run: ## Run Docker container
	@echo "$(COLOR_INFO)🐳 Running Docker container...$(COLOR_RESET)"
	@docker run -p 8080:8080 -p 8081:8081 $(IMAGE_NAME):$(IMAGE_TAG)

docker-up: ## Start production Docker Compose services
	@echo "$(COLOR_INFO)🐳 Starting production services...$(COLOR_RESET)"
	@docker-compose -f $(COMPOSE_PROD) up -d
	@echo "$(COLOR_SUCCESS)✅ Services started$(COLOR_RESET)"
	@echo "$(COLOR_INFO)📍 API: http://localhost:8080$(COLOR_RESET)"
	@echo "$(COLOR_INFO)📍 Swagger: http://localhost:8080/swagger/index.html$(COLOR_RESET)"

docker-up-dev: ## Start development Docker Compose services with hot reload
	@echo "$(COLOR_INFO)🐳 Starting development services...$(COLOR_RESET)"
	@docker-compose -f $(COMPOSE_DEV) up
	@echo "$(COLOR_SUCCESS)✅ Development services started$(COLOR_RESET)"

docker-down: ## Stop Docker Compose services
	@echo "$(COLOR_WARNING)🐳 Stopping services...$(COLOR_RESET)"
	@docker-compose -f $(COMPOSE_PROD) down
	@docker-compose -f $(COMPOSE_DEV) down
	@echo "$(COLOR_SUCCESS)✅ Services stopped$(COLOR_RESET)"

docker-restart: ## Restart Docker services
	@echo "$(COLOR_INFO)🔄 Restarting services...$(COLOR_RESET)"
	@docker-compose -f $(COMPOSE_PROD) restart
	@echo "$(COLOR_SUCCESS)✅ Services restarted$(COLOR_RESET)"

docker-logs: ## View Docker logs
	@docker-compose -f $(COMPOSE_PROD) logs -f

docker-logs-app: ## View app container logs only
	@docker-compose -f $(COMPOSE_PROD) logs -f app

docker-ps: ## Show Docker container status
	@docker-compose -f $(COMPOSE_PROD) ps

docker-shell: ## Open shell in app container
	@echo "$(COLOR_INFO)🐚 Opening shell in app container...$(COLOR_RESET)"
	@docker-compose -f $(COMPOSE_PROD) exec app sh

docker-shell-root: ## Open shell as root in app container
	@echo "$(COLOR_INFO)🐚 Opening root shell in app container...$(COLOR_RESET)"
	@docker-compose -f $(COMPOSE_PROD) exec -u root app sh

docker-clean: ## Remove containers, volumes, and images
	@echo "$(COLOR_WARNING)🧹 Cleaning Docker resources...$(COLOR_RESET)"
	@docker-compose -f $(COMPOSE_PROD) down -v --rmi all
	@docker-compose -f $(COMPOSE_DEV) down -v --rmi all
	@echo "$(COLOR_SUCCESS)✅ Docker resources cleaned$(COLOR_RESET)"

docker-prune: ## Remove unused Docker resources
	@echo "$(COLOR_WARNING)🧹 Pruning unused Docker resources...$(COLOR_RESET)"
	@docker system prune -a -f --volumes
	@echo "$(COLOR_SUCCESS)✅ Docker resources pruned$(COLOR_RESET)"

docker-health: ## Check service health
	@echo "$(COLOR_INFO)🏥 Checking service health...$(COLOR_RESET)"
	@docker-compose -f $(COMPOSE_PROD) ps
	@echo ""
	@echo "$(COLOR_INFO)📊 Health endpoint:$(COLOR_RESET)"
	@curl -s http://localhost:8080/health | jq . 2>/dev/null || curl -s http://localhost:8080/health || echo "App not responding"

docker-migrate-up: ## Run database migrations in container
	@echo "$(COLOR_INFO)⬆️  Running migrations in container...$(COLOR_RESET)"
	@docker-compose -f $(COMPOSE_PROD) exec app goose -dir db/migrations/schema mysql "ichi_user:ichi_password@tcp(mysql:3306)/ichi_db" up
	@echo "$(COLOR_SUCCESS)✅ Migrations completed$(COLOR_RESET)"

docker-migrate-down: ## Rollback last migration in container
	@echo "$(COLOR_WARNING)⬇️  Rolling back migration in container...$(COLOR_RESET)"
	@docker-compose -f $(COMPOSE_PROD) exec app goose -dir db/migrations/schema mysql "ichi_user:ichi_password@tcp(mysql:3306)/ichi_db" down
	@echo "$(COLOR_SUCCESS)✅ Rollback completed$(COLOR_RESET)"

docker-seed: ## Run database seeds in container
	@echo "$(COLOR_INFO)🌱 Running seeds in container...$(COLOR_RESET)"
	@docker-compose -f $(COMPOSE_PROD) exec app goose -dir db/migrations/seeds mysql "ichi_user:ichi_password@tcp(mysql:3306)/ichi_db" up
	@echo "$(COLOR_SUCCESS)✅ Seeds completed$(COLOR_RESET)"

docker-mysql: ## Access MySQL shell in container
	@docker-compose -f $(COMPOSE_PROD) exec mysql mysql -u ichi_user -pichi_password ichi_db

docker-redis: ## Access Redis CLI in container
	@docker-compose -f $(COMPOSE_PROD) exec redis redis-cli

docker-rabbitmq: ## Show RabbitMQ status
	@docker-compose -f $(COMPOSE_PROD) exec rabbitmq rabbitmqctl status

docker-deploy: ## Deploy using automation script
	@echo "$(COLOR_INFO)🚀 Deploying with automation script...$(COLOR_RESET)"
	@./scripts/deploy-docker.sh production

docker-deploy-dev: ## Deploy development environment
	@echo "$(COLOR_INFO)🚀 Deploying development environment...$(COLOR_RESET)"
	@./scripts/deploy-docker.sh dev

docker-stats: ## Show Docker container resource usage
	@docker stats

docker-swagger: ## Regenerate Swagger docs in container
	@echo "$(COLOR_INFO)📝 Regenerating Swagger docs in container...$(COLOR_RESET)"
	@docker-compose -f $(COMPOSE_PROD) exec app swag init -g cmd/main.go -o docs --parseDependency --parseInternal
	@echo "$(COLOR_SUCCESS)✅ Swagger docs regenerated$(COLOR_RESET)"

docker-help: ## Show Docker commands help
	@echo "$(COLOR_INFO)╔════════════════════════════════════════════════╗$(COLOR_RESET)"
	@echo "$(COLOR_INFO)║          Docker Commands Help                  ║$(COLOR_RESET)"
	@echo "$(COLOR_INFO)╚════════════════════════════════════════════════╝$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_SUCCESS)🐳 Build Commands:$(COLOR_RESET)"
	@echo "  $(COLOR_INFO)make docker-build$(COLOR_RESET)          - Build production image"
	@echo "  $(COLOR_INFO)make docker-build-dev$(COLOR_RESET)      - Build development image"
	@echo "  $(COLOR_INFO)make docker-build-no-cache$(COLOR_RESET) - Build without cache"
	@echo ""
	@echo "$(COLOR_SUCCESS)🚀 Run Commands:$(COLOR_RESET)"
	@echo "  $(COLOR_INFO)make docker-up$(COLOR_RESET)             - Start production services"
	@echo "  $(COLOR_INFO)make docker-up-dev$(COLOR_RESET)         - Start dev services (hot reload)"
	@echo "  $(COLOR_INFO)make docker-down$(COLOR_RESET)           - Stop all services"
	@echo "  $(COLOR_INFO)make docker-restart$(COLOR_RESET)        - Restart services"
	@echo ""
	@echo "$(COLOR_SUCCESS)🔧 Management Commands:$(COLOR_RESET)"
	@echo "  $(COLOR_INFO)make docker-logs$(COLOR_RESET)           - View all logs"
	@echo "  $(COLOR_INFO)make docker-shell$(COLOR_RESET)          - Open app shell"
	@echo "  $(COLOR_INFO)make docker-health$(COLOR_RESET)         - Check health status"
	@echo "  $(COLOR_INFO)make docker-ps$(COLOR_RESET)             - Show containers"
	@echo ""
	@echo "$(COLOR_SUCCESS)🗄️  Database Commands:$(COLOR_RESET)"
	@echo "  $(COLOR_INFO)make docker-migrate-up$(COLOR_RESET)     - Run migrations"
	@echo "  $(COLOR_INFO)make docker-seed$(COLOR_RESET)           - Run seeds"
	@echo "  $(COLOR_INFO)make docker-mysql$(COLOR_RESET)          - MySQL shell"
	@echo ""
	@echo "$(COLOR_SUCCESS)🧹 Cleanup Commands:$(COLOR_RESET)"
	@echo "  $(COLOR_INFO)make docker-clean$(COLOR_RESET)          - Remove containers & volumes"
	@echo "  $(COLOR_INFO)make docker-prune$(COLOR_RESET)          - Prune unused resources"
	@echo ""
	@echo "$(COLOR_SUCCESS)📚 Quick Start:$(COLOR_RESET)"
	@echo "  1. $(COLOR_INFO)make docker-deploy$(COLOR_RESET)      - Automated deployment"
	@echo "  2. View Swagger: $(COLOR_INFO)http://localhost:8080/swagger/index.html$(COLOR_RESET)"
	@echo ""

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

# Testing Makefile Targets for ichi-go
# Add these to your main Makefile

##@ Testing

.PHONY: test test-unit test-integration test-auth test-coverage test-verbose test-race test-bench test-clean

# Colors for output
COLOR_TEST = \033[35m

test: ## Run all tests
	@echo "$(COLOR_INFO)🧪 Running all tests...$(COLOR_RESET)"
	@go test ./... -v
	@echo "$(COLOR_SUCCESS)✅ All tests passed$(COLOR_RESET)"

test-unit: ## Run unit tests only
	@echo "$(COLOR_INFO)🧪 Running unit tests...$(COLOR_RESET)"
	@go test ./internal/applications/*/service/... ./pkg/... -v -short
	@echo "$(COLOR_SUCCESS)✅ Unit tests passed$(COLOR_RESET)"

test-integration: ## Run integration tests only
	@echo "$(COLOR_INFO)🧪 Running integration tests...$(COLOR_RESET)"
	@go test ./... -v -run Integration
	@echo "$(COLOR_SUCCESS)✅ Integration tests passed$(COLOR_RESET)"

test-auth: ## Run auth service tests only
	@echo "$(COLOR_INFO)🔐 Running auth service tests...$(COLOR_RESET)"
	@go test ./internal/applications/auth/service/... -v
	@echo "$(COLOR_SUCCESS)✅ Auth tests passed$(COLOR_RESET)"

test-coverage: ## Run tests with coverage report
	@echo "$(COLOR_INFO)📊 Running tests with coverage...$(COLOR_RESET)"
	@go test ./... -cover -coverprofile=coverage.out
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_SUCCESS)✅ Coverage report generated: coverage.html$(COLOR_RESET)"
	@echo "$(COLOR_INFO)📈 Opening coverage report in browser...$(COLOR_RESET)"
	@open coverage.html 2>/dev/null || xdg-open coverage.html 2>/dev/null || echo "Please open coverage.html manually"

test-coverage-auth: ## Run auth service tests with coverage
	@echo "$(COLOR_INFO)📊 Running auth tests with coverage...$(COLOR_RESET)"
	@go test ./internal/applications/auth/service/... -cover -coverprofile=auth_coverage.out
	@go tool cover -func=auth_coverage.out
	@go tool cover -html=auth_coverage.out -o auth_coverage.html
	@echo "$(COLOR_SUCCESS)✅ Auth coverage report generated: auth_coverage.html$(COLOR_RESET)"

test-verbose: ## Run tests with verbose output
	@echo "$(COLOR_INFO)🧪 Running tests with verbose output...$(COLOR_RESET)"
	@go test ./... -v -count=1

test-race: ## Run tests with race detector
	@echo "$(COLOR_WARNING)⚠️  Running tests with race detector...$(COLOR_RESET)"
	@go test ./... -race -v
	@echo "$(COLOR_SUCCESS)✅ No race conditions detected$(COLOR_RESET)"

test-bench: ## Run benchmark tests
	@echo "$(COLOR_INFO)⚡ Running benchmark tests...$(COLOR_RESET)"
	@go test ./... -bench=. -benchmem -run=^$
	@echo "$(COLOR_SUCCESS)✅ Benchmarks completed$(COLOR_RESET)"

test-bench-auth: ## Run auth service benchmarks
	@echo "$(COLOR_TEST)⚡ Running auth service benchmarks...$(COLOR_RESET)"
	@go test ./internal/applications/auth/service/... -bench=. -benchmem -run=^$

test-watch: ## Watch and run tests on file changes (requires entr)
	@echo "$(COLOR_INFO)👀 Watching for changes...$(COLOR_RESET)"
	@echo "$(COLOR_WARNING)Requires 'entr' to be installed$(COLOR_RESET)"
	@find . -name '*.go' | entr -c go test ./... -v

test-clean: ## Clean test cache and coverage files
	@echo "$(COLOR_WARNING)🧹 Cleaning test artifacts...$(COLOR_RESET)"
	@go clean -testcache
	@rm -f coverage.out coverage.html auth_coverage.out auth_coverage.html
	@rm -f *.test
	@echo "$(COLOR_SUCCESS)✅ Test artifacts cleaned$(COLOR_RESET)"

test-failfast: ## Run tests and stop on first failure
	@echo "$(COLOR_INFO)🧪 Running tests with fail-fast...$(COLOR_RESET)"
	@go test ./... -v -failfast

test-timeout: ## Run tests with custom timeout (default: 30s)
	@echo "$(COLOR_INFO)⏱️  Running tests with 30s timeout...$(COLOR_RESET)"
	@go test ./... -v -timeout 30s

test-json: ## Run tests with JSON output
	@echo "$(COLOR_INFO)📄 Running tests with JSON output...$(COLOR_RESET)"
	@go test ./... -json > test-results.json
	@echo "$(COLOR_SUCCESS)✅ Results saved to test-results.json$(COLOR_RESET)"

test-profile: ## Run tests and generate CPU profile
	@echo "$(COLOR_INFO)🔬 Running tests with CPU profiling...$(COLOR_RESET)"
	@go test ./... -cpuprofile=cpu.prof -memprofile=mem.prof
	@echo "$(COLOR_SUCCESS)✅ Profiles generated: cpu.prof, mem.prof$(COLOR_RESET)"
	@echo "$(COLOR_INFO)💡 Analyze with: go tool pprof cpu.prof$(COLOR_RESET)"

test-list: ## List all available tests
	@echo "$(COLOR_INFO)📋 Available tests:$(COLOR_RESET)"
	@go test ./... -list . 2>/dev/null | grep -v "^ok" | grep -v "^?"

test-deps: ## Check test dependencies
	@echo "$(COLOR_INFO)📦 Checking test dependencies...$(COLOR_RESET)"
	@go list -f '{{.TestImports}}' ./... | tr ' ' '\n' | sort -u | grep -v "^$"

##@ Test Generators

test-install-mockery: ## Install mockery tool (v3.6+)
	@echo "$(COLOR_INFO)📦 Installing mockery v3...$(COLOR_RESET)"
	@go install github.com/vektra/mockery/v3@latest
	@echo "$(COLOR_SUCCESS)✅ Mockery v3 installed$(COLOR_RESET)"

test-generate-mocks: test-install-mockery ## Generate mocks using portable script (works with any module name)
	@echo "$(COLOR_INFO)🎭 Generating mocks (module-agnostic)...$(COLOR_RESET)"
	@./scripts/generate-mocks.sh
	@echo "$(COLOR_SUCCESS)✅ Mocks generated successfully$(COLOR_RESET)"

test-generate-mocks-config: test-install-mockery ## Generate mocks using .mockery.yaml (requires module name update)
	@echo "$(COLOR_INFO)🎭 Generating mocks from .mockery.yaml...$(COLOR_RESET)"
	@mockery --config .mockery.yaml
	@echo "$(COLOR_SUCCESS)✅ Mocks generated successfully$(COLOR_RESET)"

test-generate-mocks-all: test-install-mockery ## Generate all mocks (force regenerate)
	@echo "$(COLOR_INFO)🎭 Force regenerating all mocks...$(COLOR_RESET)"
	@rm -rf internal/*/mocks pkg/*/mocks
	@./scripts/generate-mocks.sh
	@echo "$(COLOR_SUCCESS)✅ All mocks regenerated$(COLOR_RESET)"

test-clean-mocks: ## Clean generated mock files
	@echo "$(COLOR_WARNING)🧹 Cleaning mock files...$(COLOR_RESET)"
	@find . -type d -name "mocks" -exec rm -rf {} + 2>/dev/null || true
	@echo "$(COLOR_SUCCESS)✅ Mock files cleaned$(COLOR_RESET)"

##@ Quick Test Commands

qt: test ## Quick alias for test
qta: test-auth ## Quick alias for test-auth
qtc: test-coverage ## Quick alias for test-coverage
qtca: test-coverage-auth ## Quick alias for auth coverage
qtr: test-race ## Quick alias for test-race
qtb: test-bench ## Quick alias for benchmark

##@ CI/CD Testing

test-ci: ## Run tests in CI environment
	@echo "$(COLOR_INFO)🤖 Running tests for CI...$(COLOR_RESET)"
	@go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out
	@echo "$(COLOR_SUCCESS)✅ CI tests completed$(COLOR_RESET)"

test-ci-short: ## Run fast tests for CI
	@echo "$(COLOR_INFO)🤖 Running short tests for CI...$(COLOR_RESET)"
	@go test ./... -short -race -coverprofile=coverage.out
	@echo "$(COLOR_SUCCESS)✅ Short CI tests completed$(COLOR_RESET)"

##@ Test Reports

test-report: test-coverage ## Generate comprehensive test report
	@echo "$(COLOR_INFO)📊 Generating test report...$(COLOR_RESET)"
	@mkdir -p reports
	@go test ./... -v -json > reports/test-results.json
	@go test ./... -coverprofile=reports/coverage.out
	@go tool cover -html=reports/coverage.out -o reports/coverage.html
	@go tool cover -func=reports/coverage.out > reports/coverage.txt
	@echo "$(COLOR_SUCCESS)✅ Reports generated in ./reports/$(COLOR_RESET)"
	@echo "$(COLOR_INFO)📁 Files:$(COLOR_RESET)"
	@echo "   - reports/test-results.json"
	@echo "   - reports/coverage.html"
	@echo "   - reports/coverage.txt"

##@ Test Help

test-help: ## Show detailed testing help
	@echo "$(COLOR_INFO)╔════════════════════════════════════════════════╗$(COLOR_RESET)"
	@echo "$(COLOR_INFO)║          Testing Commands Help                 ║$(COLOR_RESET)"
	@echo "$(COLOR_INFO)╚════════════════════════════════════════════════╝$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_SUCCESS)📚 Basic Commands:$(COLOR_RESET)"
	@echo "  $(COLOR_INFO)make test$(COLOR_RESET)              - Run all tests"
	@echo "  $(COLOR_INFO)make test-unit$(COLOR_RESET)         - Run unit tests only"
	@echo "  $(COLOR_INFO)make test-auth$(COLOR_RESET)         - Run auth service tests"
	@echo "  $(COLOR_INFO)make test-coverage$(COLOR_RESET)     - Generate coverage report"
	@echo ""
	@echo "$(COLOR_SUCCESS)🔍 Advanced Commands:$(COLOR_RESET)"
	@echo "  $(COLOR_INFO)make test-race$(COLOR_RESET)         - Detect race conditions"
	@echo "  $(COLOR_INFO)make test-bench$(COLOR_RESET)        - Run benchmarks"
	@echo "  $(COLOR_INFO)make test-verbose$(COLOR_RESET)      - Verbose output"
	@echo "  $(COLOR_INFO)make test-failfast$(COLOR_RESET)     - Stop on first failure"
	@echo ""
	@echo "$(COLOR_SUCCESS)⚡ Quick Aliases:$(COLOR_RESET)"
	@echo "  $(COLOR_INFO)make qt$(COLOR_RESET)                - Alias for test"
	@echo "  $(COLOR_INFO)make qta$(COLOR_RESET)               - Alias for test-auth"
	@echo "  $(COLOR_INFO)make qtc$(COLOR_RESET)               - Alias for test-coverage"
	@echo "  $(COLOR_INFO)make qtr$(COLOR_RESET)               - Alias for test-race"
	@echo ""
	@echo "$(COLOR_SUCCESS)🤖 CI/CD Commands:$(COLOR_RESET)"
	@echo "  $(COLOR_INFO)make test-ci$(COLOR_RESET)           - Full CI test suite"
	@echo "  $(COLOR_INFO)make test-ci-short$(COLOR_RESET)     - Quick CI tests"
	@echo ""
	@echo "$(COLOR_SUCCESS)💡 Tips:$(COLOR_RESET)"
	@echo "  • Run $(COLOR_INFO)make test-coverage$(COLOR_RESET) before committing"
	@echo "  • Use $(COLOR_INFO)make test-race$(COLOR_RESET) to catch concurrency issues"
	@echo "  • Run $(COLOR_INFO)make test-bench$(COLOR_RESET) to monitor performance"
	@echo "  • Use $(COLOR_INFO)make test-clean$(COLOR_RESET) if tests behave oddly"
	@echo ""

# Example GitHub Actions workflow
.PHONY: test-github-actions
test-github-actions: ## Show example GitHub Actions config
	@echo "$(COLOR_INFO)Example GitHub Actions workflow:$(COLOR_RESET)"
	@cat << 'EOF'
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      - name: Run tests
        run: make test-ci
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out