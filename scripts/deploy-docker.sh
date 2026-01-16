#!/bin/bash

# Docker Deployment Script for ichi-go
# Usage: ./scripts/deploy-docker.sh [environment]
# Environment: local, staging, production (default: production)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
ENVIRONMENT=${1:-production}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
COMPOSE_FILE="docker-compose.yml"

if [ "$ENVIRONMENT" = "local" ] || [ "$ENVIRONMENT" = "dev" ]; then
    COMPOSE_FILE="docker-compose.dev.yml"
fi

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_dependencies() {
    log_info "Checking dependencies..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi

    log_info "Dependencies check passed"
}

validate_config() {
    CONFIG_FILE="$PROJECT_DIR/config.$ENVIRONMENT.yaml"

    if [ ! -f "$CONFIG_FILE" ]; then
        log_error "Configuration file not found: $CONFIG_FILE"
        log_info "Available configs:"
        ls -1 "$PROJECT_DIR"/config.*.yaml 2>/dev/null || echo "  No config files found"
        exit 1
    fi

    # Check for default/insecure values in production
    if [ "$ENVIRONMENT" = "production" ]; then
        log_info "Validating production configuration..."

        if grep -q "CHANGE-THIS-SECRET-KEY" "$CONFIG_FILE"; then
            log_error "Production config contains default JWT secret! Please update config.$ENVIRONMENT.yaml"
            exit 1
        fi

        if grep -q "dev-secret-key" "$CONFIG_FILE"; then
            log_error "Production config contains development secret! Please update config.$ENVIRONMENT.yaml"
            exit 1
        fi

        log_info "Configuration validation passed"
    fi
}


stop_existing_containers() {
    log_info "Stopping existing containers..."
    cd "$PROJECT_DIR"
    docker-compose -f "$COMPOSE_FILE" down || true
    log_info "Existing containers stopped"
}

build_images() {
    log_info "Building Docker images..."
    cd "$PROJECT_DIR"
    docker-compose -f "$COMPOSE_FILE" build --no-cache
    log_info "Docker images built successfully"
}

start_containers() {
    log_info "Starting containers..."
    cd "$PROJECT_DIR"
    docker-compose -f "$COMPOSE_FILE" up -d
    log_info "Containers started"
}

wait_for_health() {
    log_info "Waiting for services to be healthy..."

    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if docker-compose -f "$COMPOSE_FILE" ps | grep -q "healthy"; then
            log_info "Services are healthy"
            return 0
        fi

        log_info "Waiting for services... (attempt $attempt/$max_attempts)"
        sleep 5
        ((attempt++))
    done

    log_error "Services did not become healthy in time"
    docker-compose -f "$COMPOSE_FILE" logs
    return 1
}

run_migrations() {
    log_info "Running database migrations..."
    cd "$PROJECT_DIR"

    # Wait for database to be ready
    sleep 10

    docker-compose -f "$COMPOSE_FILE" exec -T app \
        goose -dir db/migrations/schema mysql \
        "ichi_user:ichi_password@tcp(mysql:3306)/ichi_db" up || {
        log_warn "Migration failed or no new migrations to run"
    }

    log_info "Migrations completed"
}

run_seeds() {
    if [ "$ENVIRONMENT" != "production" ]; then
        log_info "Running database seeds..."
        cd "$PROJECT_DIR"

        docker-compose -f "$COMPOSE_FILE" exec -T app \
            goose -dir db/migrations/seeds mysql \
            "ichi_user:ichi_password@tcp(mysql:3306)/ichi_db" up || {
            log_warn "Seeds failed or already run"
        }

        log_info "Seeds completed"
    fi
}

verify_deployment() {
    log_info "Verifying deployment..."

    # Check if containers are running
    if ! docker-compose -f "$COMPOSE_FILE" ps | grep -q "Up"; then
        log_error "Containers are not running!"
        docker-compose -f "$COMPOSE_FILE" logs
        exit 1
    fi

    # Check health endpoint
    log_info "Checking health endpoint..."
    sleep 5

    local max_attempts=10
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -f http://localhost:8080/health &> /dev/null; then
            log_info "Health check passed!"
            return 0
        fi

        log_info "Waiting for app to be ready... (attempt $attempt/$max_attempts)"
        sleep 3
        ((attempt++))
    done

    log_error "Health check failed!"
    docker-compose -f "$COMPOSE_FILE" logs app
    exit 1
}

show_status() {
    log_info "Deployment Status:"
    echo ""
    docker-compose -f "$COMPOSE_FILE" ps
    echo ""
    log_info "Services:"
    echo "  - REST API: http://localhost:8080"
    echo "  - Swagger:  http://localhost:8080/swagger/index.html"
    echo "  - Web:      http://localhost:8081"
    echo "  - RabbitMQ: http://localhost:15672 (admin/admin_password)"
    echo ""
    log_info "To view logs: docker-compose -f $COMPOSE_FILE logs -f"
    log_info "To stop:      docker-compose -f $COMPOSE_FILE down"
}

# Main deployment flow
main() {
    log_info "Starting deployment for environment: $ENVIRONMENT"
    echo ""

    check_dependencies
    validate_config
    stop_existing_containers

    log_info "Building and starting services..."
    build_images
    start_containers

    wait_for_health
    run_migrations
    run_seeds
    verify_deployment

    echo ""
    log_info "Deployment completed successfully!"
    echo ""
    show_status
}

# Run main function
main
