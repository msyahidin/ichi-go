# Docker Deployment Guide for ichi-go

Complete guide for deploying ichi-go using Docker with YAML configuration.

## Quick Start

```bash
# 1. Update production config
nano config.production.yaml
# Change JWT secret and other sensitive values

# 2. Deploy
make docker-deploy

# 3. Access
# API: http://localhost:8080
# Swagger: http://localhost:8080/swagger/index.html
```

## Table of Contents

- [Configuration](#configuration)
- [Deployment](#deployment)
- [Development](#development)
- [Services](#services)
- [Makefile Commands](#makefile-commands)
- [Kubernetes](#kubernetes)
- [Troubleshooting](#troubleshooting)

## Configuration

### YAML-Based Configuration

The application uses YAML files for all configuration (no `.env` files):

| File | Environment | Usage |
|------|-------------|-------|
| `config.production.yaml` | Production | `APP_ENV=production` |
| `config.staging.yaml` | Staging | `APP_ENV=staging` |
| `config.local.yaml` | Development | `APP_ENV=local` |
| `config.example.yaml` | Template | Reference only |

### Configuration Structure

```yaml
app:
  env: "production"
  name: "ichi-go"
  debug: false

http:
  port: 8080
  timeout: 10000
  cors:
    allow_origins:
      - "https://yourdomain.com"

database:
  driver: "mysql"
  host: "mysql"          # Docker service name
  port: 3306
  user: "ichi_user"
  password: "ichi_password"
  name: "ichi_db"
  max_open_conns: 200

cache:
  driver: "redis"
  host: "redis"          # Docker service name
  port: 6379
  pool_size: 50

queue:
  enabled: true
  rabbitmq:
    enabled: true
    connection:
      host: "rabbitmq"   # Docker service name
      port: 5672
      username: "admin"
      password: "admin_password"

auth:
  jwt:
    enabled: true
    signing_method: "HS256"
    secret_key: "CHANGE-THIS-IN-PRODUCTION"  # Min 32 chars
    access_token_ttl: "15m"
    refresh_token_ttl: "168h"
```

### Important: Production Configuration

Before deploying to production, update `config.production.yaml`:

```yaml
# 1. JWT Secret (REQUIRED)
auth:
  jwt:
    secret_key: "your-strong-random-secret-minimum-32-characters"

# 2. CORS Origins
http:
  cors:
    allow_origins:
      - "https://yourdomain.com"
      - "https://api.yourdomain.com"

# 3. Issuer/Audience
auth:
  jwt:
    issuer: "your-company-name"
    audience:
      - "your-app-users"
```

The deployment script validates that you've changed the default secrets.

## Deployment

### Production Deployment

#### Method 1: Automated Script (Recommended)

```bash
make docker-deploy
```

This will:
1. Check dependencies (Docker, Docker Compose)
2. Validate configuration (checks for default secrets)
3. Stop existing containers
4. Build Docker images
5. Start all services
6. Wait for health checks
7. Run database migrations
8. Verify deployment
9. Show access URLs

#### Method 2: Manual Steps

```bash
# Build images
make docker-build

# Start services
make docker-up

# Wait for services to be healthy
make docker-health

# Run migrations
make docker-migrate-up

# Optional: Seed database
make docker-seed
```

### Docker Compose Stack

The production stack (`docker-compose.yml`) includes:

- **Application**: Go REST API (port 8080, 8081)
- **MySQL 8.0**: Database with persistent volume
- **Redis 7**: Cache with persistent volume
- **RabbitMQ 3.12**: Message queue with management UI

All services have health checks and automatic restarts.

## Development

### Development with Hot Reload

```bash
# Start development environment
make docker-up-dev

# Your code changes will automatically reload!
```

The development setup:
- Uses `config.local.yaml` automatically
- Mounts your source code as a volume
- Includes Air for hot reload
- Shows logs in the terminal
- Uses development credentials

### Development Workflow

```bash
# Start with logs visible
docker-compose -f docker-compose.dev.yml up

# In another terminal, make code changes
# The app will automatically rebuild and restart

# View specific service logs
docker-compose -f docker-compose.dev.yml logs -f app

# Run migrations in dev
docker-compose -f docker-compose.dev.yml exec app \
  goose -dir db/migrations/schema mysql \
  "root:password@tcp(mysql:3306)/mydb" up

# Stop
docker-compose -f docker-compose.dev.yml down
```

## Services

### Service Access

| Service | URL | Credentials | Notes |
|---------|-----|-------------|-------|
| REST API | http://localhost:8080 | JWT token | Main API endpoints |
| Web Server | http://localhost:8081 | - | Template routes |
| Swagger UI | http://localhost:8080/swagger/index.html | Public | API documentation |
| Health Check | http://localhost:8080/health | Public | Service health |
| RabbitMQ UI | http://localhost:15672 | admin/admin_password | Queue management |
| MySQL | localhost:3306 | ichi_user/ichi_password | Database |
| Redis | localhost:6379 | No auth | Cache |

### Database Credentials

**Production** (from `config.production.yaml` and `docker-compose.yml`):
- User: `ichi_user`
- Password: `ichi_password`
- Database: `ichi_db`

**Development** (from `config.local.yaml` and `docker-compose.dev.yml`):
- User: `root`
- Password: `password`
- Database: `mydb`

### Swagger API Documentation

Access interactive API documentation at:
```
http://localhost:8080/swagger/index.html
```

Features:
- **Interactive Testing**: Try endpoints directly from browser
- **JWT Authentication**: Click "Authorize" and enter `Bearer <token>`
- **Multi-language**: Supports English/Indonesian validation
- **Auto-generated**: Updates when you change code

To regenerate Swagger docs:
```bash
make docker-swagger
# or locally
make swagger-gen
```

## Makefile Commands

### Build Commands

```bash
make docker-build              # Build production image
make docker-build-dev          # Build development image
make docker-build-no-cache     # Build without cache
```

### Deployment Commands

```bash
make docker-deploy             # Automated production deployment
make docker-deploy-dev         # Automated development deployment
make docker-up                 # Start production services
make docker-up-dev             # Start development with hot reload
make docker-down               # Stop all services
make docker-restart            # Restart services
```

### Management Commands

```bash
make docker-logs               # View all container logs
make docker-logs-app           # View app logs only
make docker-shell              # Open shell in app container
make docker-shell-root         # Open root shell in app
make docker-ps                 # Show container status
make docker-health             # Check service health
make docker-stats              # Show resource usage
```

### Database Commands

```bash
make docker-migrate-up         # Run database migrations
make docker-migrate-down       # Rollback last migration
make docker-seed               # Run database seeds
make docker-mysql              # Access MySQL shell
make docker-redis              # Access Redis CLI
make docker-rabbitmq           # Show RabbitMQ status
make docker-swagger            # Regenerate Swagger docs
```

### Cleanup Commands

```bash
make docker-clean              # Remove containers, volumes, images
make docker-prune              # Remove all unused Docker resources
```

### Help Commands

```bash
make docker-help               # Show Docker commands
make help                      # Show all Makefile commands
```

## Kubernetes

### Kubernetes Deployment

For production Kubernetes deployments, use the provided manifests:

```bash
# 1. Update secrets in k8s/deployment.yaml
nano k8s/deployment.yaml

# 2. Update ingress domain
nano k8s/ingress.yaml

# 3. Apply
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/ingress.yaml

# 4. Check status
kubectl get pods
kubectl get ingress
```

### Kubernetes Features

The K8s deployment includes:

- **Deployment**: 3 replicas with rolling updates
- **HorizontalPodAutoscaler**: Auto-scale 3-10 pods based on CPU/memory
- **PodDisruptionBudget**: Maintain availability during updates
- **Health Checks**: Liveness and readiness probes
- **ConfigMap**: YAML configuration management
- **Secrets**: Sensitive data protection
- **Ingress**: TLS, CORS, rate limiting, security headers
- **Resource Limits**: CPU and memory constraints

### Configuration in Kubernetes

The ConfigMap includes `config.production.yaml`. Update it:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ichi-go-config
data:
  config.production.yaml: |
    app:
      env: "production"
      name: "ichi-go"
    # ... rest of your config
```

Secrets are stored separately:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ichi-go-secrets
type: Opaque
stringData:
  database.password: "your-strong-password"
  rabbitmq.password: "your-strong-password"
  jwt.secret: "your-strong-secret-min-32-chars"
```

## Troubleshooting

### Common Issues

#### 1. Port Already in Use

**Error**: `Bind for 0.0.0.0:8080 failed: port is already allocated`

**Solution**: Update `config.production.yaml`:
```yaml
http:
  port: 9090  # Change to available port
```

Then restart:
```bash
make docker-down
make docker-up
```

#### 2. Default Secret Detected

**Error**: `Production config contains default JWT secret!`

**Solution**: Edit `config.production.yaml`:
```yaml
auth:
  jwt:
    secret_key: "generate-a-strong-secret-at-least-32-characters-long"
```

#### 3. Database Connection Failed

**Error**: `Error 1045: Access denied`

**Solution**:
```bash
# Check MySQL is running
make docker-ps

# Check MySQL logs
docker-compose logs mysql

# Reset database
make docker-down
docker volume rm rathalos-kit_mysql_data
make docker-up
```

#### 4. RabbitMQ Plugin Not Enabled

**Error**: `exchange type x-delayed-message not found`

**Solution**:
```bash
# Check plugin file
cat docker/rabbitmq/enabled_plugins
# Should contain: [rabbitmq_management,rabbitmq_delayed_message_exchange].

# Restart RabbitMQ
docker-compose restart rabbitmq
```

#### 5. Swagger Not Loading

**Solution**:
```bash
# Regenerate Swagger docs
make docker-swagger

# Or rebuild image
make docker-build-no-cache
make docker-up
```

### Debug Commands

```bash
# View all logs
docker-compose logs

# Follow specific service
docker-compose logs -f app

# Check container health
docker inspect --format='{{json .State.Health}}' ichi-go-app | jq

# Check network
docker network inspect ichi-network

# Check volumes
docker volume ls
docker volume inspect rathalos-kit_mysql_data

# Manual health check
curl -v http://localhost:8080/health

# Access MySQL manually
docker-compose exec mysql mysql -u ichi_user -pichi_password ichi_db

# Test Redis
docker-compose exec redis redis-cli ping
```

### Complete Reset

If all else fails:

```bash
# Stop and remove everything
make docker-clean

# Remove all Docker resources
make docker-prune

# Rebuild from scratch
make docker-build-no-cache
make docker-deploy
```

## Best Practices

### Security

1. **Change Secrets**: Always change JWT secret and passwords in `config.production.yaml`
2. **Use Strong Passwords**: Minimum 16 characters for production
3. **Enable TLS**: Use HTTPS in production (configure in Ingress)
4. **Limit CORS**: Only allow specific domains
5. **Scan Images**: Run `docker scan ichi-go:latest` regularly
6. **Update Dependencies**: Keep base images and dependencies updated

### Performance

1. **Connection Pooling**: Tune `database.max_open_conns` in config
2. **Redis Pool Size**: Adjust `cache.pool_size` based on load
3. **Worker Pools**: Configure RabbitMQ `worker_pool_size`
4. **Resource Limits**: Set in docker-compose or K8s
5. **Enable Caching**: Use Redis for frequently accessed data

### Monitoring

1. **Health Checks**: Monitor `/health` endpoint
2. **Logs**: Centralize with ELK, CloudWatch, etc.
3. **Metrics**: Export Prometheus metrics
4. **Alerts**: Set up alerts for failures
5. **APM**: Consider Application Performance Monitoring

### Backups

```bash
# Backup MySQL
docker-compose exec mysql mysqldump -u ichi_user -pichi_password ichi_db > backup.sql

# Restore MySQL
docker-compose exec -T mysql mysql -u ichi_user -pichi_password ichi_db < backup.sql

# Backup Redis
docker-compose exec redis redis-cli SAVE
docker cp ichi-go-redis:/data/dump.rdb ./redis-backup.rdb
```

## Files Reference

### Core Docker Files

- `Dockerfile` - Multi-stage production build (~20MB)
- `Dockerfile.dev` - Development image with hot reload
- `.dockerignore` - Build optimization
- `docker-compose.yml` - Production stack
- `docker-compose.dev.yml` - Development stack

### Configuration

- `config.production.yaml` - Production configuration
- `config.staging.yaml` - Staging configuration
- `config.local.yaml` - Local development
- `config.example.yaml` - Template/reference

### Scripts

- `scripts/deploy-docker.sh` - Automated deployment

### Kubernetes

- `k8s/deployment.yaml` - Deployment, Service, HPA, PDB
- `k8s/ingress.yaml` - Ingress with TLS

### CI/CD

- `.github/workflows/docker-build.yml` - Automated builds

## Architecture

```
┌─────────────────────────────────────────────┐
│         Docker Compose Stack                 │
│                                              │
│  ┌──────────────────────────────────────┐  │
│  │  Application Container                │  │
│  │  - ichi-go binary                     │  │
│  │  - config.production.yaml             │  │
│  │  - Ports: 8080, 8081                  │  │
│  └──────────────────────────────────────┘  │
│       │                                      │
│       ├─── MySQL (persistent volume)        │
│       ├─── Redis (persistent volume)        │
│       └─── RabbitMQ (persistent volume)     │
│                                              │
│  Configuration: APP_ENV=production          │
│  Config File: config.production.yaml        │
└─────────────────────────────────────────────┘
```

## Support

- **Documentation**: See `CLAUDE.md` for architecture details
- **Docker Help**: `make docker-help`
- **All Commands**: `make help`
- **Swagger UI**: http://localhost:8080/swagger/index.html

---

**Environment**: Configure via `config.{APP_ENV}.yaml`
**No .env files needed** - All configuration in YAML!
